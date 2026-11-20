package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// SettlementMatch represents a confirmed match between an expected receivable and a bank statement credit from Matcher 6.
type SettlementMatch struct {
	MatchID          string          `json:"match_id"`
	RuleID           string          `json:"rule_id"`
	SourceARecordID  string          `json:"source_a_record_id"` // Matches DeterministicReceivableCode e.g. "REC-SUMUP-987654321-1"
	SourceBRecordID  string          `json:"source_b_record_id"` // Matches Open Finance transaction ID e.g. "OFP-ITAU-998877"
	SettlementAmount decimal.Decimal `json:"amount"`
	SettlementDate   time.Time       `json:"settlement_date"`
}

// MatcherWebhookPayload represents the inbound event sent by Matcher 6 when a reconciliation run finishes.
type MatcherWebhookPayload struct {
	Event         string            `json:"event"`
	ContextID     string            `json:"context_id"`
	RunID         string            `json:"run_id"`
	MatchesCount  int               `json:"matches_count"`
	Matches       []SettlementMatch `json:"matches"`
}

// SettlementExecutionResult records the outcome of posting the settlement transfer in Midaz.
type SettlementExecutionResult struct {
	InstallmentID       string `json:"installmentId"`
	MidazSettlementTxID string `json:"midazSettlementTxId"`
	Status              string `json:"status"` // "SUCCESS", "FAILED", "ALREADY_SETTLED"
	ErrorMessage        string `json:"errorMessage,omitempty"`
}
