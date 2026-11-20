package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type AcquirerType string

const (
	AcquirerSumUp   AcquirerType = "SUMUP"
	AcquirerPagBank AcquirerType = "PAGBANK"
)

func (a AcquirerType) String() string {
	return string(a)
}

// UniversalFiatClearingAccount represents the external fiat origin for funding internal acquirer liquidity.
const UniversalFiatClearingAccount = "@external/BRL:universal"

// AcquirerInternalAccountAlias returns the internal liquidity account alias for an acquirer (e.g. "@acquirer/sumup").
func AcquirerInternalAccountAlias(acquirer AcquirerType) string {
	return fmt.Sprintf("@acquirer/%s", strings.ToLower(string(acquirer)))
}

// AcquirerProvider defines the contract required for any payment acquirer adapter.
type AcquirerProvider interface {
	Type() AcquirerType
	ValidateCredentials(ctx context.Context, creds *ConnectionCredentials) error
	FetchSales(ctx context.Context, creds *ConnectionCredentials, filter SalesFilter) ([]Sale, error)
	FetchPayoutSchedule(ctx context.Context, creds *ConnectionCredentials, filter PayoutFilter) ([]PayoutSchedule, error)
	ParseWebhook(ctx context.Context, creds *ConnectionCredentials, headers map[string]string, body []byte) (*WebhookEvent, error)
}

type SalesFilter struct {
	StartDate time.Time
	EndDate   time.Time
	Limit     int
	Cursor    string
}

type PayoutFilter struct {
	StartDate time.Time
	EndDate   time.Time
}

type PayoutSchedule struct {
	Acquirer               AcquirerType    `json:"acquirer"`
	PayoutID               string          `json:"payoutId"`
	ExpectedSettlementDate time.Time       `json:"expectedSettlementDate"`
	GrossAmount            decimal.Decimal `json:"grossAmount"`
	FeeAmount              decimal.Decimal `json:"feeAmount"`
	NetAmount              decimal.Decimal `json:"netAmount"`
	AssociatedNSUs         []string        `json:"associatedNsus"`
}

type WebhookEvent struct {
	Type     string          `json:"type"` // "SALE_CREATED", "PAYOUT_COMPLETED", "CHARGEBACK"
	RawEvent string          `json:"rawEvent"`
	Sale     *Sale           `json:"sale,omitempty"`
	Payout   *PayoutSchedule `json:"payout,omitempty"`
}
