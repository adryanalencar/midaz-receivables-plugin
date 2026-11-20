package domain

import (
	"context"

	"github.com/shopspring/decimal"
)

// MatcherSourceRecord represents an individual item uploaded to Matcher 6.
type MatcherSourceRecord struct {
	ID           string          `json:"id"`
	Date         string          `json:"expected_date"` // YYYY-MM-DD
	Amount       decimal.Decimal `json:"net_amount"`
	MerchantID   string          `json:"merchant_id"`
	Acquirer     string          `json:"acquirer"`
	ConnectionID string          `json:"connection_id"`
}

// IMatcherClient defines the outbound port to Matcher 6 (:4018).
type IMatcherClient interface {
	// EnsureContextAndSources verifies or creates the reconciliation context and sources in Matcher 6.
	EnsureContextAndSources(ctx context.Context) (contextID string, sourceAID string, sourceBID string, error error)

	// UploadExpectedReceivables uploads a batch of pending installments (Source A) to Matcher 6.
	UploadExpectedReceivables(ctx context.Context, contextID, sourceAID string, records []MatcherSourceRecord) error

	// TriggerReconciliationRun instructs Matcher 6 to execute a reconciliation pass.
	TriggerReconciliationRun(ctx context.Context, contextID string) (runID string, error error)

	// HealthCheck asserts connectivity to Matcher 6 /health.
	HealthCheck(ctx context.Context) error
}
