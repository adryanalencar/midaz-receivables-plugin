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

	// 3. Post D+0 double-entry transaction in Midaz for each installment
	clearingAlias := fmt.Sprintf("@external/acquirer_%s", strings.ToLower(string(sale.Acquirer)))

	for i := range sale.Installments {
		inst := &sale.Installments[i]
		code := inst.DeterministicReceivableCode()

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
		}

		tx := &domain.MidazTransactionPayload{
			ChartOfAccountsGroupName: "ACQUIRER_RECEIVABLES",
			Description: fmt.Sprintf("%s Card Receivable - Installment %d/%d (NSU: %s)",
				sale.Acquirer, inst.InstallmentNumber, sale.InstallmentsCount, sale.NSU),
			Code:                    code,
			Amount:                  inst.GrossAmount,
			FeeAmount:               inst.MDRFee,
			FeeAccountAlias:         "@platform/mdr_revenue",
			SourceAccountAlias:      clearingAlias,
			DestinationAccountAlias: acc.AccountAlias,
			Metadata:                metadata,
		}

		if _, err := h.midazClient.PostTransaction(ctx, input.OrganizationID, input.LedgerID, tx); err != nil {
			return fmt.Errorf("ingest_acquirer_sale: failed posting D+0 transaction for installment %d: %w", inst.InstallmentNumber, err)
		}
	}

	return nil
}
