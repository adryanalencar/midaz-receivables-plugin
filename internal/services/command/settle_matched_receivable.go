package command

import (
	"context"
	"fmt"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
)

type MatchedReceivableItem struct {
	SourceARecordID string          `json:"source_a_record_id"` // e.g. "REC-SUMUP-987654321-1"
	Amount          decimal.Decimal `json:"amount"`             // e.g. 97.00
}

type SettlementDeductionItem struct {
	FeeType     string          `json:"feeType"`     // e.g. "MACHINE_RENTAL", "CHARGEBACK", "ANTICIPATION_FEE"
	ChargedBy   string          `json:"chargedBy"`   // e.g. "SUMUP", "PAGBANK"
	Description string          `json:"description"` // e.g. "Monthly POS Terminal Rental"
	Amount      decimal.Decimal `json:"amount"`
}

type SettleMatchedReceivableInput struct {
	OrganizationID           string                    `json:"organizationId"`
	LedgerID                 string                    `json:"ledgerId"`
	MatchID                  string                    `json:"matchId"`
	MatcherRunID             string                    `json:"matcherRunId"`
	BankTransactionID        string                    `json:"bankTransactionId"`        // e.g. "OFP-SICREDI-998877"
	TargetOpenFinanceAccount string                    `json:"targetOpenFinanceAccount"` // e.g. "@contadomeucliente/openfinance"
	TotalDepositAmount       decimal.Decimal           `json:"totalDepositAmount"`       // Net amount deposited in bank
	SettlementDate           time.Time                 `json:"settlementDate"`
	Items                    []MatchedReceivableItem   `json:"items"`
	Deductions               []SettlementDeductionItem `json:"deductions,omitempty"` // Tariffs/deductions deducted by acquirer
}

type IOpenFinanceNotifier interface {
	MarkTransactionProcessed(ctx context.Context, bankTxID, midazTxCode string) error
}

type SettleMatchedReceivableHandler struct {
	repo               domain.IReceivablesRepository
	midazClient        domain.IMidazLedgerClient
	openFinanceNotifier IOpenFinanceNotifier
}

func NewSettleMatchedReceivableHandler(
	repo domain.IReceivablesRepository,
	midazClient domain.IMidazLedgerClient,
) *SettleMatchedReceivableHandler {
	return &SettleMatchedReceivableHandler{
		repo:        repo,
		midazClient: midazClient,
	}
}

func (h *SettleMatchedReceivableHandler) WithOpenFinanceNotifier(notifier IOpenFinanceNotifier) *SettleMatchedReceivableHandler {
	h.openFinanceNotifier = notifier
	return h
}

func (h *SettleMatchedReceivableHandler) Execute(ctx context.Context, input SettleMatchedReceivableInput) error {
	if len(input.Items) == 0 {
		return fmt.Errorf("settle_matched_receivable: no matched items provided")
	}

	settlementDate := input.SettlementDate
	if settlementDate.IsZero() {
		settlementDate = time.Now().UTC()
	}

	// 1. Resolve and Commit each pending receivable transaction in Midaz
	var sourceLegs []domain.TransactionSourceLeg
	var resolvedInstallments []*domain.Installment
	var parentAccount *domain.ReceivablesAccount

	for _, item := range input.Items {
		acquirer, nsu, installmentNum, err := domain.ParseDeterministicReceivableCode(item.SourceARecordID)
		if err != nil {
			return fmt.Errorf("settle_matched_receivable: failed parsing source_a code %s: %w", item.SourceARecordID, err)
		}

		inst, err := h.repo.FindInstallmentByReceivableCode(ctx, acquirer, nsu, installmentNum)
		if err != nil {
			return fmt.Errorf("settle_matched_receivable: database query failed: %w", err)
		}
		if inst == nil {
			return fmt.Errorf("settle_matched_receivable: installment not found for code %s", item.SourceARecordID)
		}

		// Idempotency: skip already reconciled installments
		if inst.Status == domain.InstallmentStatusReconciled {
			continue
		}

		acc, err := h.repo.FindReceivablesAccountByID(ctx, inst.ReceivablesAccountID)
		if err != nil {
			return fmt.Errorf("settle_matched_receivable: account lookup failed: %w", err)
		}
		if acc == nil {
			return fmt.Errorf("settle_matched_receivable: parent receivables account %s not found", inst.ReceivablesAccountID)
		}
		parentAccount = acc

		// COMMIT the pending transaction in Midaz (releases on-hold in @acquirer/<name> and credits @receivables)
		if inst.MidazReceivableTxID != nil && *inst.MidazReceivableTxID != "" {
			if err := h.midazClient.CommitTransaction(ctx, input.OrganizationID, input.LedgerID, *inst.MidazReceivableTxID); err != nil {
				return fmt.Errorf("settle_matched_receivable: failed committing pending receivable %s: %w", item.SourceARecordID, err)
			}
		}

		itemAmount := item.Amount
		if itemAmount.IsZero() {
			itemAmount = inst.NetAmount
		}

		sourceLegs = append(sourceLegs, domain.TransactionSourceLeg{
			AccountAlias: acc.AccountAlias,
			Amount:       itemAmount,
			Description:  fmt.Sprintf("Receivable %s (NSU: %s)", item.SourceARecordID, inst.NSU),
			Metadata: map[string]any{
				"receivable_code":    item.SourceARecordID,
				"nsu":                inst.NSU,
				"installment_number": inst.InstallmentNumber,
				"acquirer":           string(inst.Acquirer),
				"receivable_id":      inst.ID,
			},
		})

		resolvedInstallments = append(resolvedInstallments, inst)
	}

	// If all items were already reconciled, exit cleanly
	if len(sourceLegs) == 0 {
		return nil
	}

	totalSourcesAmount := decimal.Zero
	for _, leg := range sourceLegs {
		totalSourcesAmount = totalSourcesAmount.Add(leg.Amount)
	}

	targetOpenFinanceAccount := input.TargetOpenFinanceAccount
	if targetOpenFinanceAccount == "" && parentAccount != nil && parentAccount.CheckingAccountAlias != "" {
		targetOpenFinanceAccount = parentAccount.CheckingAccountAlias
	}
	if targetOpenFinanceAccount == "" {
		return fmt.Errorf("settle_matched_receivable: targetOpenFinanceAccount is required and no default checking account alias is set")
	}

	feeAccountAlias := ""
	if parentAccount != nil {
		feeAccountAlias = parentAccount.FeeAccountAlias
		if feeAccountAlias == "" && parentAccount.AccountAlias != "" {
			feeAccountAlias = parentAccount.AccountAlias + "/fees"
		}
	}

	// Build optional deduction legs
	var deductionLegs []domain.TransactionDestinationLeg
	totalDeductions := decimal.Zero
	for _, ded := range input.Deductions {
		if ded.Amount.IsPositive() && feeAccountAlias != "" {
			totalDeductions = totalDeductions.Add(ded.Amount)
			deductionLegs = append(deductionLegs, domain.TransactionDestinationLeg{
				AccountAlias: feeAccountAlias,
				Amount:       ded.Amount,
				Description:  fmt.Sprintf("%s Deduction: %s", ded.ChargedBy, ded.Description),
				Metadata: map[string]any{
					"fee_type":   ded.FeeType,
					"charged_by": ded.ChargedBy,
					"reason":     ded.Description,
					"bank_tx_id": input.BankTransactionID,
				},
			})
		}
	}

	// 2. Post atomic N:M transfer: N sources (@XPTO/receivables) -> Bank Account + Fee Subaccounts
	settlePayload := &domain.MultiSourceTransactionPayload{
		ChartOfAccountsGroupName: "OPEN_FINANCE_SYNC",
		Description: fmt.Sprintf("Open Finance Deposit %s (Conciliated with %d Receivables)",
			input.BankTransactionID, len(sourceLegs)),
		Code:                    input.BankTransactionID,
		TotalAmount:             totalSourcesAmount,
		AssetCode:               "BRL",
		Sources:                 sourceLegs,
		DestinationAccountAlias: targetOpenFinanceAccount,
		Deductions:              deductionLegs,
		Metadata: map[string]any{
			"matcher_run_id":          input.MatcherRunID,
			"match_id":                input.MatchID,
			"bank_tx_id":              input.BankTransactionID,
			"receivables_count":       len(sourceLegs),
			"deductions_count":        len(deductionLegs),
			"settlement_date":         settlementDate.Format("2006-01-02"),
			"reconciliation_strategy": "MULTI_SOURCE_SETTLEMENT",
		},
	}

	settleTxID, err := h.midazClient.PostMultiSourceTransaction(ctx, input.OrganizationID, input.LedgerID, settlePayload)
	if err != nil {
		return fmt.Errorf("settle_matched_receivable: failed posting multi-source settlement to Midaz: %w", err)
	}

	// 3. Mark each installment as RECONCILED in repository
	for _, inst := range resolvedInstallments {
		if err := h.repo.MarkInstallmentReconciled(ctx, inst.ID, settleTxID, input.MatcherRunID, settlementDate); err != nil {
			continue
		}
	}

	// 4. Inter-plugin notification: notify Open Finance that this bank deposit was settled via reconciliation
	if h.openFinanceNotifier != nil && input.BankTransactionID != "" {
		_ = h.openFinanceNotifier.MarkTransactionProcessed(ctx, input.BankTransactionID, settleTxID)
	}

	return nil
}
