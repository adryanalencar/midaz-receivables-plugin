package pagbank

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPagBankAdapter_ValidateCredentials(t *testing.T) {
	t.Run("Valid Token", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/charges", r.URL.Path)
			assert.Equal(t, "Bearer pbg_test_token", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"charges": []}`))
		}))
		defer ts.Close()

		adapter := NewAdapter(ts.URL, ts.Client())
		err := adapter.ValidateCredentials(context.Background(), &domain.ConnectionCredentials{
			Token: "pbg_test_token",
		})
		assert.NoError(t, err)
	})

	t.Run("Missing Token", func(t *testing.T) {
		adapter := NewAdapter("", nil)
		err := adapter.ValidateCredentials(context.Background(), &domain.ConnectionCredentials{})
		assert.ErrorIs(t, err, ErrMissingToken)
	})

	t.Run("Unauthorized Token", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer ts.Close()

		adapter := NewAdapter(ts.URL, ts.Client())
		err := adapter.ValidateCredentials(context.Background(), &domain.ConnectionCredentials{
			Token: "invalid_token",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials (HTTP 401)")
	})
}

func TestPagBankAdapter_FetchSales(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/charges", r.URL.Path)
		assert.Equal(t, "Bearer pbg_token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"charges": [
				{
					"id": "CHAR_201",
					"reference_id": "REF_201",
					"status": "PAID",
					"paid_at": "2026-06-01T12:00:00Z",
					"amount": { "value": 30000, "currency": "BRL" },
					"payment_method": { "type": "CREDIT_CARD", "installments": 1 },
					"payment_response": { "reference": "NSU_201" },
					"fees": { "seller": 600 }
				},
				{
					"id": "CHAR_CANCELLED",
					"status": "CANCELED"
				}
			]
		}`))
	}))
	defer ts.Close()

	adapter := NewAdapter(ts.URL, ts.Client())
	sales, err := adapter.FetchSales(context.Background(), &domain.ConnectionCredentials{Token: "pbg_token"}, domain.SalesFilter{Limit: 10})
	require.NoError(t, err)
	require.Len(t, sales, 1)

	assert.Equal(t, "NSU_201", sales[0].NSU)
	assert.Equal(t, "CHAR_201", sales[0].TID)
}

func TestPagBankAdapter_FetchPayoutSchedule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/transfers", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"id": "trans_001",
				"scheduled_date": "2026-06-10",
				"amount": 100000,
				"fee": 2500,
				"net_amount": 97500,
				"status": "PAID",
				"charge_ids": ["CHAR_1", "CHAR_2"]
			}
		]`))
	}))
	defer ts.Close()

	adapter := NewAdapter(ts.URL, ts.Client())
	payouts, err := adapter.FetchPayoutSchedule(context.Background(), &domain.ConnectionCredentials{Token: "pbg_token"}, domain.PayoutFilter{})
	require.NoError(t, err)
	require.Len(t, payouts, 1)

	assert.Equal(t, "trans_001", payouts[0].PayoutID)
	assert.Len(t, payouts[0].AssociatedNSUs, 2)
}

func TestPagBankAdapter_ParseWebhook(t *testing.T) {
	adapter := NewAdapter("", nil)
	payload := []byte(`{
		"id": "CHAR_HOOK_1",
		"reference_id": "REF_HOOK_1",
		"status": "PAID",
		"paid_at": "2026-06-01T15:00:00Z",
		"amount": { "value": 15000, "currency": "BRL" },
		"payment_method": { "type": "CREDIT_CARD", "installments": 1 },
		"payment_response": { "reference": "NSU_HOOK_PAG" }
	}`)

	event, err := adapter.ParseWebhook(context.Background(), &domain.ConnectionCredentials{}, nil, payload)
	require.NoError(t, err)
	assert.Equal(t, "SALE_CREATED", event.Type)
	require.NotNil(t, event.Sale)
	assert.Equal(t, "NSU_HOOK_PAG", event.Sale.NSU)
}
