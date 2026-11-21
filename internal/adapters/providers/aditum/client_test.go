package aditum

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAditumAdapter_ValidateCredentials(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/Charges", r.URL.Path)
			assert.Equal(t, "1", r.URL.Query().Get("pageSize"))
			assert.Equal(t, "merc-test-1", r.Header.Get("merchantId"))
			assert.Equal(t, "Bearer tok-test-1", r.Header.Get("Authorization"))

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"charges": [], "totalItems": 0}`))
		}))
		defer ts.Close()

		adapter := NewAdapter(ts.URL, ts.URL, ts.Client())
		assert.Equal(t, domain.AcquirerAditum, adapter.Type())

		creds := &domain.ConnectionCredentials{
			MerchantCode: "merc-test-1",
			Token:        "tok-test-1",
		}

		err := adapter.ValidateCredentials(context.Background(), creds)
		require.NoError(t, err)
	})

	t.Run("missing credentials", func(t *testing.T) {
		adapter := NewAdapter("", "", nil)
		err := adapter.ValidateCredentials(context.Background(), &domain.ConnectionCredentials{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "merchant_code")
	})

	t.Run("unauthorized", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer ts.Close()

		adapter := NewAdapter(ts.URL, ts.URL, ts.Client())
		creds := &domain.ConnectionCredentials{
			MerchantCode: "merc-test-1",
			Token:        "invalid-tok",
		}

		err := adapter.ValidateCredentials(context.Background(), creds)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials (HTTP 401)")
	})
}

func TestAditumAdapter_FetchSales(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/Charges", r.URL.Path)
		assert.Equal(t, "50", r.URL.Query().Get("pageSize"))
		assert.Equal(t, "merc-test-1", r.Header.Get("merchantId"))

		resp := AditumChargesListResponse{
			CurrentPage: 1,
			PageSize:    50,
			TotalItems:  1,
			Charges: []AditumChargeItem{
				{
					ID:           "charge-001",
					NSU:          "nsu-111222",
					ChargeStatus: "Authorized",
					DataCobranca: "2026-03-20T12:00:00",
					Transactions: []AditumTransactionItem{
						{
							TransactionID:     "tx-001",
							Amount:            15000, // R$ 150.00
							FeeAmount:         450,   // R$ 4.50
							NetAmount:         14550, // R$ 145.50
							IsCapture:         true,
							PaymentType:       "credit",
							InstallmentNumber: 1,
						},
					},
				},
				{
					ID:           "charge-canceled",
					NSU:          "nsu-canceled",
					ChargeStatus: "Canceled",
					Transactions: []AditumTransactionItem{
						{
							TransactionID: "tx-canceled",
							Amount:        5000,
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	adapter := NewAdapter(ts.URL, ts.URL, ts.Client())
	creds := &domain.ConnectionCredentials{
		MerchantCode: "merc-test-1",
		Token:        "tok-test-1",
	}

	sales, err := adapter.FetchSales(context.Background(), creds, domain.SalesFilter{
		StartDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	// Canceled charge should be excluded
	require.Len(t, sales, 1)

	assert.Equal(t, domain.AcquirerAditum, sales[0].Acquirer)
	assert.Equal(t, "nsu-111222", sales[0].NSU)
	assert.True(t, decimal.NewFromFloat(150.00).Equal(sales[0].GrossAmount))
	assert.True(t, decimal.NewFromFloat(4.50).Equal(sales[0].TotalMDRFee))
	assert.True(t, decimal.NewFromFloat(145.50).Equal(sales[0].TotalNetAmount))
}

func TestAditumAdapter_FetchPayoutSchedule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AditumChargesListResponse{
			Charges: []AditumChargeItem{
				{
					ID:           "chg-payout-1",
					NSU:          "nsu-payout-1",
					ChargeStatus: "Authorized",
					DataCobranca: "2026-03-01T12:00:00",
					Transactions: []AditumTransactionItem{
						{
							TransactionID:     "tx-p-1",
							Amount:            10000,
							FeeAmount:         300,
							NetAmount:         9700,
							IsCapture:         true,
							PaymentType:       "debit",
							InstallmentNumber: 1,
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	adapter := NewAdapter(ts.URL, ts.URL, ts.Client())
	creds := &domain.ConnectionCredentials{
		MerchantCode: "merc-test-1",
		Token:        "tok-test-1",
	}

	payouts, err := adapter.FetchPayoutSchedule(context.Background(), creds, domain.PayoutFilter{})
	require.NoError(t, err)
	require.Len(t, payouts, 1)

	assert.Equal(t, domain.AcquirerAditum, payouts[0].Acquirer)
	assert.True(t, decimal.NewFromInt(100).Equal(payouts[0].GrossAmount))
	assert.True(t, decimal.NewFromInt(3).Equal(payouts[0].FeeAmount))
	assert.True(t, decimal.NewFromInt(97).Equal(payouts[0].NetAmount))
	assert.Contains(t, payouts[0].AssociatedNSUs, "nsu-payout-1")
}

func TestAditumAdapter_ParseWebhook(t *testing.T) {
	adapter := NewAdapter("", "", nil)
	creds := &domain.ConnectionCredentials{
		MerchantCode: "merc-aditum-1",
		Token:        "secret-token-123",
	}

	webhookBody := []byte(`{
		"ChargeId": "chg-hook-99",
		"Nsu": "nsu-hook-99",
		"Transactions": [
			{
				"TransactionId": "tx-hook-99",
				"Amount": 10000,
				"FeeAmount": 200,
				"NetAmount": 9800,
				"IsCapture": true,
				"PaymentType": "credit",
				"InstallmentNumber": 1
			}
		]
	}`)

	t.Run("valid auth header", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer secret-token-123",
		}

		event, err := adapter.ParseWebhook(context.Background(), creds, headers, webhookBody)
		require.NoError(t, err)
		require.NotNil(t, event)
		assert.Equal(t, "SALE_CREATED", event.Type)
		assert.Equal(t, "nsu-hook-99", event.Sale.NSU)
		assert.Equal(t, domain.AcquirerAditum, event.Sale.Acquirer)
	})

	t.Run("invalid auth header", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer wrong-token",
		}

		event, err := adapter.ParseWebhook(context.Background(), creds, headers, webhookBody)
		require.Error(t, err)
		assert.Nil(t, event)
		assert.Equal(t, ErrInvalidWebhookAuth, err)
	})
}
