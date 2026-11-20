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
	Pending                  bool            `json:"pending"`
	Metadata                 map[string]any  `json:"metadata"`
}

// TransactionSourceLeg represents an individual origin leg in an N:1 multi-source transaction.
type TransactionSourceLeg struct {
	AccountAlias string          `json:"accountAlias"`
	Amount       decimal.Decimal `json:"amount"`
	Description  string          `json:"description"`
	Metadata     map[string]any  `json:"metadata,omitempty"`
}

// MultiSourceTransactionPayload encapsulates an N:1 transfer from multiple receivables sources to one destination.
type MultiSourceTransactionPayload struct {
	ChartOfAccountsGroupName string                 `json:"chartOfAccountsGroupName"`
	Description              string                 `json:"description"`
	Code                     string                 `json:"code"`
	TotalAmount              decimal.Decimal        `json:"totalAmount"`
	AssetCode                string                 `json:"assetCode"`
	Sources                  []TransactionSourceLeg `json:"sources"`
	DestinationAccountAlias  string                 `json:"destinationAccountAlias"`
	Metadata                 map[string]any         `json:"metadata"`
}

// IMidazLedgerClient defines the outbound port to Midaz Onboarding and Transaction APIs.
type IMidazLedgerClient interface {
	// EnsureClearingAccount verifies or creates the universal @external/acquirer_<name> account.
	EnsureClearingAccount(ctx context.Context, organizationID, ledgerID string, acquirer AcquirerType) (string, error)

	// CreateAccount provisions an internal account (receivables wallet or openfinance mirrored account) in Midaz.
	CreateAccount(ctx context.Context, organizationID, ledgerID, alias, name, accountType string) (string, error)

	// CheckAccountExists verifies whether an account alias exists in Midaz.
	CheckAccountExists(ctx context.Context, organizationID, ledgerID, alias string) (bool, error)

	// PostTransaction executes a double-entry transaction in Midaz (can be pending or committed).
	PostTransaction(ctx context.Context, organizationID, ledgerID string, tx *MidazTransactionPayload) (string, error)

	// CommitTransaction transitions a pending transaction to committed, realizing the credit in the account.
	CommitTransaction(ctx context.Context, organizationID, ledgerID, transactionID string) error

	// PostMultiSourceTransaction executes an atomic N:1 transaction transferring from N receivables sources to 1 destination.
	PostMultiSourceTransaction(ctx context.Context, organizationID, ledgerID string, tx *MultiSourceTransactionPayload) (string, error)
}
