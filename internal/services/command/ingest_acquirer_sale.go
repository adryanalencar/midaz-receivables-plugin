package command

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
)

var (
	ErrNilSalePayload    = errors.New("ingest_acquirer_sale: nil sale payload")
	ErrEmptyConnectionID = errors.New("ingest_acquirer_sale: connection_id is required")
	ErrZeroGrossAmount   = errors.New("ingest_acquirer_sale: gross amount must be greater than zero")
	ErrZeroInstallments  = errors.New("ingest_acquirer_sale: installments count must be at least 1")
)

type IngestAcquirerSaleInput struct {
	OrganizationID string       `json:"organizationId"`
	LedgerID       string       `json:"ledgerId"`
	Sale           *domain.Sale `json:"sale"`
}

type IngestAcquirerSaleHandler struct {
	repo        domain.IReceivablesRepository
	midazClient domain.IMidazLedgerClient
}

func NewIngestAcquirerSaleHandler(repo domain.IReceivablesRepository, midazClient domain.IMidazLedgerClient) *IngestAcquirerSaleHandler {
	return &IngestAcquirerSaleHandler{
		repo:        repo,
		midazClient: midazClient,
	}
}

func (h *IngestAcquirerSaleHandler) Execute(ctx context.Context, input IngestAcquirerSaleInput) error {
	sale := input.Sale
	if sale == nil {
		return ErrNilSalePayload
	}
	if strings.TrimSpace(sale.ConnectionID) == "" {
		return ErrEmptyConnectionID
	}
	if sale.GrossAmount.IsZero() || sale.GrossAmount.IsNegative() {
		return ErrZeroGrossAmount
	}
	if sale.InstallmentsCount <= 0 || len(sale.Installments) == 0 {
		return ErrZeroInstallments
	}

	// 1. Look up parent connection and receivables account to obtain Midaz target alias
	conn, err := h.repo.FindAcquirerConnectionByID(ctx, sale.ConnectionID)
	if err != nil {
		return fmt.Errorf("ingest_acquirer_sale: failed looking up connection: %w", err)
	}
	if conn == nil {
		return fmt.Errorf("ingest_acquirer_sale: connection %s not found", sale.ConnectionID)
	}

	acc, err := h.repo.FindReceivablesAccountByID(ctx, conn.ReceivablesAccountID)
	if err != nil {
		return fmt.Errorf("ingest_acquirer_sale: failed looking up account: %w", err)
	}
	if acc == nil {
		return fmt.Errorf("ingest_acquirer_sale: parent receivables account %s not found", conn.ReceivablesAccountID)
	}

	sale.ReceivablesAccountID = acc.ID
	for i := range sale.Installments {
		sale.Installments[i].ReceivablesAccountID = acc.ID
		sale.Installments[i].ConnectionID = conn.ID
		sale.Installments[i].Acquirer = sale.Acquirer
		sale.Installments[i].NSU = sale.NSU
	}

	// 2. Persist sale and installments to local database
	if err := h.repo.SaveSaleWithInstallments(ctx, sale); err != nil {
		return fmt.Errorf("ingest_acquirer_sale: failed saving sale and installments: %w", err)
	}

	// 3. Post D+0 two-step double-entry in Midaz for each installment:
	//    Step A: Fund the internal acquirer account from universal external fiat (immediate)
	//    Step B: Allocate from internal acquirer account to receivables account as PENDING (on-hold)
	universalClearingAlias := acc.ExternalAccountAlias
	if universalClearingAlias == "" {
		universalClearingAlias = domain.UniversalFiatClearingAccount
	}
	acquirerInternalAlias := domain.AcquirerInternalAccountAlias(sale.Acquirer)

	for i := range sale.Installments {
		inst := &sale.Installments[i]
		code := inst.DeterministicReceivableCode()
		fundingCode := fmt.Sprintf("FUND-%s-%s-%d", sale.Acquirer, sale.NSU, inst.InstallmentNumber)

		// Step A: Immediate Funding to Internal Acquirer Account
		fundTx := &domain.MidazTransactionPayload{
			ChartOfAccountsGroupName: "ACQUIRER_FUNDING",
			Description: fmt.Sprintf("Funding %s Acquirer Liquidity for Sale NSU %s (Installment %d/%d)",
				sale.Acquirer, sale.NSU, inst.InstallmentNumber, sale.InstallmentsCount),
			Code:                    fundingCode,
			Amount:                  inst.GrossAmount,
			SourceAccountAlias:      universalClearingAlias,
			DestinationAccountAlias: acquirerInternalAlias,
			Pending:                 false,
			Metadata: map[string]any{
				"acquirer":           string(sale.Acquirer),
				"nsu":                sale.NSU,
				"installment_number": inst.InstallmentNumber,
			},
		}

		if _, err := h.midazClient.PostTransaction(ctx, input.OrganizationID, input.LedgerID, fundTx); err != nil {
			return fmt.Errorf("ingest_acquirer_sale: failed posting funding transaction for installment %d: %w", inst.InstallmentNumber, err)
		}

		// Step B: Post Pending Transaction from internal acquirer account (puts funds on-hold in @acquirer/<name>)
		mdrAccountAlias := acc.MDRRevenueAccountAlias
		if mdrAccountAlias == "" {
			mdrAccountAlias = acc.AccountAlias + "/mdr_revenue"
		}

		metadata := map[string]any{
			"acquirer":                 string(sale.Acquirer),
			"connection_id":            conn.ID,
			"receivables_account_id":   acc.ID,
			"nsu":                      sale.NSU,
			"tid":                      sale.TID,
			"installment_number":       inst.InstallmentNumber,
			"total_installments":       sale.InstallmentsCount,
			"gross_amount":             inst.GrossAmount.StringFixed(2),
			"mdr_fee":                  inst.MDRFee.StringFixed(2),
			"net_amount":               inst.NetAmount.StringFixed(2),
			"expected_settlement_date": inst.ExpectedSettlementDate.Format("2006-01-02"),
			"reason":                   "mdr_discount",
			"charged_by":               string(sale.Acquirer),
		}

		tx := &domain.MidazTransactionPayload{
			ChartOfAccountsGroupName: "ACQUIRER_RECEIVABLES",
			Description: fmt.Sprintf("%s Card Receivable - Installment %d/%d (NSU: %s)",
				sale.Acquirer, inst.InstallmentNumber, sale.InstallmentsCount, sale.NSU),
			Code:                    code,
			Amount:                  inst.GrossAmount,
			FeeAmount:               inst.MDRFee,
			FeeAccountAlias:         mdrAccountAlias,
			SourceAccountAlias:      acquirerInternalAlias,
			DestinationAccountAlias: acc.AccountAlias,
			Pending:                 true,
			Metadata:                metadata,
		}

		txID, err := h.midazClient.PostTransaction(ctx, input.OrganizationID, input.LedgerID, tx)
		if err != nil {
			return fmt.Errorf("ingest_acquirer_sale: failed posting D+0 pending transaction for installment %d: %w", inst.InstallmentNumber, err)
		}

		inst.MidazReceivableTxID = &txID
		_ = h.repo.UpdateInstallmentReceivableTxID(ctx, inst.ID, txID)
	}

	return nil
}
