package sumup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSumUpAdapter_ValidateCredentials(t *testing.T) {
	t.Run("Valid Credentials", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v0.1/me", r.URL.Path)
			assert.Equal(t, "Bearer sup_test_key", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"default_merchant_profile": "merchant_123"}`))
		}))
		defer ts.Close()

		adapter := NewAdapter(ts.URL, ts.Client())
		err := adapter.ValidateCredentials(context.Background(), &domain.ConnectionCredentials{
			APIKey: "sup_test_key",
		})
		assert.NoError(t, err)
	})

	t.Run("Missing APIKey", func(t *testing.T) {
		adapter := NewAdapter("", nil)
		err := adapter.ValidateCredentials(context.Background(), &domain.ConnectionCredentials{})
		assert.ErrorIs(t, err, ErrMissingAPIKey)
	})

	t.Run("Unauthorized Credentials", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer ts.Close()

		adapter := NewAdapter(ts.URL, ts.Client())
		err := adapter.ValidateCredentials(context.Background(), &domain.ConnectionCredentials{
			APIKey: "invalid_key",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials (HTTP 401)")
	})
}

func TestSumUpAdapter_FetchSales(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v0.1/me/transactions/history", r.URL.Path)
		assert.Equal(t, "Bearer sup_key", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "tx_101",
					"transaction_code": "NSU_101",
					"amount": 200.0,
					"fee_amount": 4.0,
					"net_amount": 196.0,
					"installments_count": 1,
					"timestamp": "2026-06-01T10:00:00Z",
					"status": "SUCCESSFUL",
					"card": { "brand": "VISA", "type": "CREDIT" }
				},
				{
					"id": "tx_failed",
					"status": "FAILED",
					"amount": 50.0
				}
			]
		}`))
	}))
	defer ts.Close()

	adapter := NewAdapter(ts.URL, ts.Client())
	sales, err := adapter.FetchSales(context.Background(), &domain.ConnectionCredentials{APIKey: "sup_key"}, domain.SalesFilter{Limit: 10})
	require.NoError(t, err)
	require.Len(t, sales, 1) // Failed transaction ignored

	assert.Equal(t, "NSU_101", sales[0].NSU)
	assert.Equal(t, "tx_101", sales[0].TID)
}

func TestSumUpAdapter_FetchPayoutSchedule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v0.1/me/payouts", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"id": "payout_001",
				"date": "2026-06-05",
				"amount": 500.0,
				"fee": 12.50,
				"net_amount": 487.50,
				"currency": "BRL",
				"status": "SUCCESSFUL",
				"transaction_ids": ["tx_1", "tx_2"]
			}
		]`))
	}))
	defer ts.Close()

	adapter := NewAdapter(ts.URL, ts.Client())
	payouts, err := adapter.FetchPayoutSchedule(context.Background(), &domain.ConnectionCredentials{APIKey: "sup_key"}, domain.PayoutFilter{})
	require.NoError(t, err)
	require.Len(t, payouts, 1)

	assert.Equal(t, "payout_001", payouts[0].PayoutID)
	assert.Len(t, payouts[0].AssociatedNSUs, 2)
}

func TestSumUpAdapter_ParseWebhook(t *testing.T) {
	adapter := NewAdapter("", nil)
	payload := []byte(`{
		"id": "tx_hook_1",
		"transaction_code": "NSU_HOOK_1",
		"amount": 80.0,
		"fee_amount": 2.0,
		"net_amount": 78.0,
		"installments_count": 1,
		"timestamp": "2026-06-01T15:00:00Z",
		"status": "SUCCESSFUL"
	}`)

	event, err := adapter.ParseWebhook(context.Background(), &domain.ConnectionCredentials{}, nil, payload)
	require.NoError(t, err)
	assert.Equal(t, "SALE_CREATED", event.Type)
	require.NotNil(t, event.Sale)
	assert.Equal(t, "NSU_HOOK_1", event.Sale.NSU)
}
