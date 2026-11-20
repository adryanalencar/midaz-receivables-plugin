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

func TestMatcherClient_HealthCheck(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/version", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"version": "1.0.0"}`))
	}))
	defer ts.Close()

	cfg := &bootstrap.Config{MatcherURL: ts.URL}
	client := NewMatcherClient(cfg, ts.Client())

	err := client.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestMatcherClient_EnsureContextAndSources(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/contexts":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"items": []}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/contexts":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id": "ctx_123", "slug": "acquirers-settlement"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/contexts/ctx_123/sources":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"items": []}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/contexts/ctx_123/sources":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id": "src_auto", "name": "auto"}`))
		default:
			t.Fatalf("unexpected call: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	cfg := &bootstrap.Config{
		MatcherURL:         ts.URL,
		MatcherContextSlug: "acquirers-settlement",
	}
	client := NewMatcherClient(cfg, ts.Client())

	ctxID, srcA, srcB, err := client.EnsureContextAndSources(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "ctx_123", ctxID)
	assert.NotEmpty(t, srcA)
	assert.NotEmpty(t, srcB)
}

func TestMatcherClient_UploadExpectedReceivables(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/imports/contexts/ctx_1/sources/src_1/upload", r.URL.Path)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	cfg := &bootstrap.Config{MatcherURL: ts.URL}
	client := NewMatcherClient(cfg, ts.Client())

	records := []domain.MatcherSourceRecord{
		{
			ID:         "REC-SUMUP-001-1",
			Date:       "2026-08-01",
			Amount:     decimal.NewFromFloat(97.00),
			MerchantID: "m_1",
			Acquirer:   "SUMUP",
		},
	}

	err := client.UploadExpectedReceivables(context.Background(), "ctx_1", "src_1", records)
	assert.NoError(t, err)
}

func TestMatcherClient_TriggerReconciliationRun(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/contexts/ctx_1/reconciliations/runs", r.URL.Path)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"runId": "run_999", "status": "PENDING"}`))
	}))
	defer ts.Close()

	cfg := &bootstrap.Config{MatcherURL: ts.URL}
	client := NewMatcherClient(cfg, ts.Client())

	runID, err := client.TriggerReconciliationRun(context.Background(), "ctx_1")
	require.NoError(t, err)
	assert.Equal(t, "run_999", runID)
}
