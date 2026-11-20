package command

import (
	"context"
	"fmt"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
)

type EnsureAcquirerAccountsHandler struct {
	midazClient domain.IMidazLedgerClient
}

func NewEnsureAcquirerAccountsHandler(midazClient domain.IMidazLedgerClient) *EnsureAcquirerAccountsHandler {
	return &EnsureAcquirerAccountsHandler{
		midazClient: midazClient,
	}
}

// Execute ensures that the universal clearing accounts for SumUp and PagBank exist in Midaz.
func (h *EnsureAcquirerAccountsHandler) Execute(ctx context.Context, organizationID, ledgerID string) error {
	acquirers := []domain.AcquirerType{
		domain.AcquirerSumUp,
		domain.AcquirerPagBank,
	}

	for _, acquirer := range acquirers {
		if _, err := h.midazClient.EnsureClearingAccount(ctx, organizationID, ledgerID, acquirer); err != nil {
			return fmt.Errorf("ensure_acquirer_accounts: failed for %s: %w", acquirer, err)
		}
	}

	return nil
}
