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

type SettleMatchedReceivableInput struct {
	OrganizationID           string                  `json:"organizationId"`
	LedgerID                 string                  `json:"ledgerId"`
	MatchID                  string                  `json:"matchId"`
	MatcherRunID             string                  `json:"matcherRunId"`
	BankTransactionID        string                  `json:"bankTransactionId"`        // e.g. "OFP-SICREDI-998877"
	TargetOpenFinanceAccount string                  `json:"targetOpenFinanceAccount"` // e.g. "@contadomeucliente/openfinance"
	TotalDepositAmount       decimal.Decimal         `json:"totalDepositAmount"`       // e.g. 300.00
	SettlementDate           time.Time               `json:"settlementDate"`
	Items                    []MatchedReceivableItem `json:"items"`
}

type SettleMatchedReceivableHandler struct {
	repo        domain.IReceivablesRepository
	midazClient domain.IMidazLedgerClient
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

		// COMMIT the pending transaction in Midaz (moves funds from pending to committed in @receivables)
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

	totalAmount := input.TotalDepositAmount
	if totalAmount.IsZero() {
		for _, leg := range sourceLegs {
			totalAmount = totalAmount.Add(leg.Amount)
		}
	}

	targetOpenFinanceAccount := input.TargetOpenFinanceAccount
	if targetOpenFinanceAccount == "" && len(resolvedInstallments) > 0 {
		firstAcc, err := h.repo.FindReceivablesAccountByID(ctx, resolvedInstallments[0].ReceivablesAccountID)
		if err == nil && firstAcc != nil && firstAcc.CheckingAccountAlias != "" {
			targetOpenFinanceAccount = firstAcc.CheckingAccountAlias
		}
	}
	if targetOpenFinanceAccount == "" {
		return fmt.Errorf("settle_matched_receivable: targetOpenFinanceAccount is required and no default checking account alias is set")
	}

	// 2. Post atomic N:1 transfer: N sources (@XPTO/receivables) -> 1 destination (@account/openfinance)
	settlePayload := &domain.MultiSourceTransactionPayload{
		ChartOfAccountsGroupName: "OPEN_FINANCE_SYNC",
		Description: fmt.Sprintf("Open Finance Deposit %s (Conciliated with %d Receivables)",
			input.BankTransactionID, len(sourceLegs)),
		Code:                    input.BankTransactionID,
		TotalAmount:             totalAmount,
		AssetCode:               "BRL",
		Sources:                 sourceLegs,
		DestinationAccountAlias: targetOpenFinanceAccount,
		Metadata: map[string]any{
			"matcher_run_id":          input.MatcherRunID,
			"match_id":                input.MatchID,
			"bank_tx_id":              input.BankTransactionID,
			"receivables_count":       len(sourceLegs),
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

	return nil
}
