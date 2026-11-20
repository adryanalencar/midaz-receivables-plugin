package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
)

type SettleMatchedReceivableInput struct {
	OrganizationID  string          `json:"organizationId"`
	LedgerID        string          `json:"ledgerId"`
	MatchID         string          `json:"matchId"`
	SourceARecordID string          `json:"sourceARecordId"` // e.g. "REC-SUMUP-987654321-1"
	SourceBRecordID string          `json:"sourceBRecordId"` // e.g. "OFP-ITAU-998877"
	SettledAmount   decimal.Decimal `json:"settledAmount"`
	SettlementDate  time.Time       `json:"settlementDate"`
	MatcherRunID    string          `json:"matcherRunId"`
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
	// 1. Parse deterministic Source A receivable code
	acquirer, nsu, installmentNum, err := domain.ParseDeterministicReceivableCode(input.SourceARecordID)
	if err != nil {
		return fmt.Errorf("settle_matched_receivable: failed parsing source_a code %s: %w", input.SourceARecordID, err)
	}

	// 2. Look up the specific installment line in repository
	inst, err := h.repo.FindInstallmentByReceivableCode(ctx, acquirer, nsu, installmentNum)
	if err != nil {
		return fmt.Errorf("settle_matched_receivable: database query failed: %w", err)
	}
	if inst == nil {
		return fmt.Errorf("settle_matched_receivable: installment not found for code %s", input.SourceARecordID)
	}

	// 3. Check for idempotency: already reconciled
	if inst.Status == domain.InstallmentStatusReconciled {
		return nil
	}

	// 4. Look up parent receivables account
	acc, err := h.repo.FindReceivablesAccountByID(ctx, inst.ReceivablesAccountID)
	if err != nil {
		return fmt.Errorf("settle_matched_receivable: account lookup failed: %w", err)
	}
	if acc == nil {
		return fmt.Errorf("settle_matched_receivable: parent receivables account %s not found", inst.ReceivablesAccountID)
	}

	// 5. Determine checking account destination
	checkingAlias := acc.CheckingAccountAlias
	if checkingAlias == "" {
		checkingAlias = strings.Replace(acc.AccountAlias, "/receivables", "/checking", 1)
	}

	settleAmount := input.SettledAmount
	if settleAmount.IsZero() {
		settleAmount = inst.NetAmount
	}

	settlementDate := input.SettlementDate
	if settlementDate.IsZero() {
		settlementDate = time.Now().UTC()
	}

	// 6. Post Midaz D+X Settlement Transaction (Extinguishes claim: @receivables -> @checking)
	code := inst.DeterministicSettlementCode()
	metadata := map[string]any{
		"matcher_run_id":     input.MatcherRunID,
		"match_id":           input.MatchID,
		"source_a_record_id": input.SourceARecordID,
		"source_b_record_id": input.SourceBRecordID,
		"acquirer":           string(inst.Acquirer),
		"nsu":                inst.NSU,
		"installment_number": inst.InstallmentNumber,
		"settlement_date":    settlementDate.Format("2006-01-02"),
	}

	tx := &domain.MidazTransactionPayload{
		ChartOfAccountsGroupName: "RECEIVABLES_SETTLEMENT",
		Description: fmt.Sprintf("%s Settlement Extinguishment - Matcher 6 Confirmed (NSU: %s)",
			inst.Acquirer, inst.NSU),
		Code:                    code,
		Amount:                  settleAmount,
		AssetCode:               "BRL",
		SourceAccountAlias:      acc.AccountAlias,
		DestinationAccountAlias: checkingAlias,
		Metadata:                metadata,
	}

	settleTxID, err := h.midazClient.PostTransaction(ctx, input.OrganizationID, input.LedgerID, tx)
	if err != nil {
		return fmt.Errorf("settle_matched_receivable: failed posting settlement to Midaz: %w", err)
	}

	// 7. Mark installment as RECONCILED in local database
	if err := h.repo.MarkInstallmentReconciled(ctx, inst.ID, settleTxID, input.MatcherRunID, settlementDate); err != nil {
		return fmt.Errorf("settle_matched_receivable: failed updating installment status: %w", err)
	}

	return nil
}
