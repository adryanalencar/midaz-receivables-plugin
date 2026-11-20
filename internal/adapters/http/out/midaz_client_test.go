package out

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMidazClient_CheckAccountExists(t *testing.T) {
	t.Run("Account Exists", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/organizations/org_1/ledgers/led_1/accounts", r.URL.Path)
			assert.Equal(t, "@my_alias", r.URL.Query().Get("alias"))
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		cfg := &bootstrap.Config{
			MidazOnboardingURL: ts.URL,
		}
		client := NewMidazClient(cfg, ts.Client())

		exists, err := client.CheckAccountExists(context.Background(), "org_1", "led_1", "@my_alias")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Account Not Found", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		cfg := &bootstrap.Config{
			MidazOnboardingURL: ts.URL,
		}
		client := NewMidazClient(cfg, ts.Client())

		exists, err := client.CheckAccountExists(context.Background(), "org_1", "led_1", "@missing")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestMidazClient_EnsureClearingAccount(t *testing.T) {
	t.Run("Universal External Clearing Already exists", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				return
			}
			t.Fatalf("unexpected call: %s %s", r.Method, r.URL.Path)
		}))
		defer ts.Close()

		cfg := &bootstrap.Config{
			MidazOnboardingURL: ts.URL,
		}
		client := NewMidazClient(cfg, ts.Client())

		alias, err := client.EnsureClearingAccount(context.Background(), "org_1", "led_1", domain.AcquirerSumUp)
		require.NoError(t, err)
		assert.Equal(t, domain.UniversalFiatClearingAccount, alias)
	})

	t.Run("Acquirer Internal Liquidity Account Creation", func(t *testing.T) {
		step := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && step == 0 {
				step++
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
				return
			}
			t.Fatalf("unexpected call: %s %s", r.Method, r.URL.Path)
		}))
		defer ts.Close()

		cfg := &bootstrap.Config{
			MidazOnboardingURL: ts.URL,
		}
		client := NewMidazClient(cfg, ts.Client())

		alias, err := client.EnsureAcquirerInternalAccount(context.Background(), "org_1", "led_1", domain.AcquirerSumUp)
		require.NoError(t, err)
		assert.Equal(t, "@acquirer/sumup", alias)
	})
}

func TestMidazClient_PostTransaction(t *testing.T) {
	t.Run("Single-Leg Transaction Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/organizations/org_1/ledgers/led_1/transactions/json", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusCreated)
		}))
		defer ts.Close()

		cfg := &bootstrap.Config{
			MidazTransactionURL: ts.URL,
		}
		client := NewMidazClient(cfg, ts.Client())

		tx := &domain.MidazTransactionPayload{
			ChartOfAccountsGroupName: "RECEIVABLES_SETTLEMENT",
			Description:              "Settlement test",
			Code:                     "SETTLE-123",
			Amount:                   decimal.NewFromFloat(100.00),
			AssetCode:                "BRL",
			SourceAccountAlias:       "@merchant/receivables",
			DestinationAccountAlias:  "@merchant/checking",
		}

		code, err := client.PostTransaction(context.Background(), "org_1", "led_1", tx)
		require.NoError(t, err)
		assert.Equal(t, "SETTLE-123", code)
	})

	t.Run("Multi-Leg Transaction With Fee", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		cfg := &bootstrap.Config{
			MidazTransactionURL: ts.URL,
		}
		client := NewMidazClient(cfg, ts.Client())

		tx := &domain.MidazTransactionPayload{
			ChartOfAccountsGroupName: "ACQUIRER_RECEIVABLES",
			Description:              "Receivable D+0",
			Code:                     "REC-SUMUP-101-1",
			Amount:                   decimal.NewFromFloat(100.00),
			FeeAmount:                decimal.NewFromFloat(3.00),
			FeeAccountAlias:          "@merchant/receivables/mdr_revenue",
			SourceAccountAlias:       "@external/acquirer_sumup",
			DestinationAccountAlias:  "@merchant/receivables",
		}

		code, err := client.PostTransaction(context.Background(), "org_1", "led_1", tx)
		require.NoError(t, err)
		assert.Equal(t, "REC-SUMUP-101-1", code)
	})

	t.Run("Conflict Idempotency (Already Created)", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusConflict)
		}))
		defer ts.Close()

		cfg := &bootstrap.Config{
			MidazTransactionURL: ts.URL,
		}
		client := NewMidazClient(cfg, ts.Client())

		tx := &domain.MidazTransactionPayload{
			Code:   "REC-ALREADY-EXISTS",
			Amount: decimal.NewFromFloat(50.00),
		}

		code, err := client.PostTransaction(context.Background(), "org_1", "led_1", tx)
		require.NoError(t, err)
		assert.Equal(t, "REC-ALREADY-EXISTS", code)
	})
}
