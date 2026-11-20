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

// Execute ensures the universal external clearing account and internal acquirer liquidity accounts exist in Midaz.
func (h *EnsureAcquirerAccountsHandler) Execute(ctx context.Context, organizationID, ledgerID string) error {
	acquirers := []domain.AcquirerType{
		domain.AcquirerSumUp,
		domain.AcquirerPagBank,
	}

	// 1. Ensure universal external fiat account: @external/BRL:universal
	if _, err := h.midazClient.EnsureClearingAccount(ctx, organizationID, ledgerID, domain.AcquirerSumUp); err != nil {
		return fmt.Errorf("ensure_acquirer_accounts: failed ensuring universal fiat clearing account: %w", err)
	}

	// 2. Ensure internal liquidity accounts for each acquirer: @acquirer/<name>
	for _, acquirer := range acquirers {
		if _, err := h.midazClient.EnsureAcquirerInternalAccount(ctx, organizationID, ledgerID, acquirer); err != nil {
			return fmt.Errorf("ensure_acquirer_accounts: failed ensuring internal account for %s: %w", acquirer, err)
		}
	}

	return nil
}
