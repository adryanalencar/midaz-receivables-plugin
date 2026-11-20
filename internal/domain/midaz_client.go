package domain

import (
	"context"

	"github.com/shopspring/decimal"
)

// MidazTransactionPayload encapsulates the fields needed to post double-entry transactions to Midaz.
type MidazTransactionPayload struct {
	ChartOfAccountsGroupName string          `json:"chartOfAccountsGroupName"`
	Description              string          `json:"description"`
	Code                     string          `json:"code"`
	Amount                   decimal.Decimal `json:"amount"`
	AssetCode                string          `json:"assetCode"`
	SourceAccountAlias       string          `json:"sourceAccountAlias"`
	DestinationAccountAlias  string          `json:"destinationAccountAlias"`
	FeeAmount                decimal.Decimal `json:"feeAmount,omitempty"`
	FeeAccountAlias          string          `json:"feeAccountAlias,omitempty"`
	Metadata                 map[string]any  `json:"metadata"`
}

// IMidazLedgerClient defines the outbound port to Midaz Onboarding and Transaction APIs.
type IMidazLedgerClient interface {
	// EnsureClearingAccount verifies or creates the universal @external/acquirer_<name> account.
	EnsureClearingAccount(ctx context.Context, organizationID, ledgerID string, acquirer AcquirerType) (string, error)

	// CreateAccount provisions an internal account (receivables wallet or checking) in Midaz.
	CreateAccount(ctx context.Context, organizationID, ledgerID, alias, name, accountType string) (string, error)

	// CheckAccountExists verifies whether an account alias exists in Midaz.
	CheckAccountExists(ctx context.Context, organizationID, ledgerID, alias string) (bool, error)

	// PostTransaction executes a double-entry transaction in Midaz (D+0 or D+X).
	PostTransaction(ctx context.Context, organizationID, ledgerID string, tx *MidazTransactionPayload) (string, error)
}
