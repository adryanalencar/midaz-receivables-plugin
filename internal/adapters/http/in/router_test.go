package in

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) CreateReceivablesAccount(ctx context.Context, account *domain.ReceivablesAccount) error {
	return nil
}
func (m *MockRepo) FindReceivablesAccountByID(ctx context.Context, id string) (*domain.ReceivablesAccount, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ReceivablesAccount), args.Error(1)
}
func (m *MockRepo) FindReceivablesAccountByAlias(ctx context.Context, alias string) (*domain.ReceivablesAccount, error) {
	return nil, nil
}
func (m *MockRepo) ListReceivablesAccounts(ctx context.Context) ([]domain.ReceivablesAccount, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.ReceivablesAccount), args.Error(1)
}
func (m *MockRepo) CreateAcquirerConnection(ctx context.Context, conn *domain.AcquirerConnection) error {
	return nil
}
func (m *MockRepo) FindAcquirerConnectionByID(ctx context.Context, id string) (*domain.AcquirerConnection, error) {
	return nil, nil
}
func (m *MockRepo) ListConnectionsByAccountID(ctx context.Context, accountID string) ([]domain.AcquirerConnection, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).([]domain.AcquirerConnection), args.Error(1)
}
func (m *MockRepo) ListActiveConnectionsForSync(ctx context.Context) ([]domain.AcquirerConnection, error) {
	return nil, nil
}
func (m *MockRepo) UpdateConnectionSyncState(ctx context.Context, id string, syncStatus string, syncError *string, lastSyncAt time.Time) error {
	return nil
}
func (m *MockRepo) UpdateConnectionStatus(ctx context.Context, id string, status domain.ConnectionStatus) error {
	return nil
}
func (m *MockRepo) SaveSaleWithInstallments(ctx context.Context, sale *domain.Sale) error {
	return nil
}
func (m *MockRepo) FindSaleByNSU(ctx context.Context, connectionID, nsu string) (*domain.Sale, error) {
	return nil, nil
}
func (m *MockRepo) FindInstallmentByReceivableCode(ctx context.Context, acquirer domain.AcquirerType, nsu string, installmentNumber int) (*domain.Installment, error) {
	return nil, nil
}
func (m *MockRepo) ListPendingInstallmentsForMatcher(ctx context.Context, windowStart, windowEnd time.Time, limit int) ([]domain.Installment, error) {
	return nil, nil
}
func (m *MockRepo) MarkInstallmentUploadedToMatcher(ctx context.Context, installmentID, matcherRecordID string) error {
	return nil
}
func (m *MockRepo) UpdateInstallmentReceivableTxID(ctx context.Context, installmentID, txID string) error {
	return nil
}
func (m *MockRepo) MarkInstallmentReconciled(ctx context.Context, installmentID, midazSettlementTxID, matcherRunID string, settlementDate time.Time) error {
	return nil
}
func (m *MockRepo) GetReceivablesCalendar(ctx context.Context, filter domain.CalendarFilter) ([]domain.Installment, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domain.Installment), args.Error(1)
}

type MockMidazClient struct {
	mock.Mock
}

func (m *MockMidazClient) EnsureClearingAccount(ctx context.Context, organizationID, ledgerID string, acquirer domain.AcquirerType) (string, error) {
	return "", nil
}
func (m *MockMidazClient) EnsureAcquirerInternalAccount(ctx context.Context, organizationID, ledgerID string, acquirer domain.AcquirerType) (string, error) {
	return "", nil
}
func (m *MockMidazClient) CreateAccount(ctx context.Context, organizationID, ledgerID, alias, name, accountType string, parentAccountID *string) (string, error) {
	return "", nil
}
func (m *MockMidazClient) CheckAccountExists(ctx context.Context, organizationID, ledgerID, alias string) (bool, error) {
	return false, nil
}
func (m *MockMidazClient) PostTransaction(ctx context.Context, organizationID, ledgerID string, tx *domain.MidazTransactionPayload) (string, error) {
	return "", nil
}
func (m *MockMidazClient) CommitTransaction(ctx context.Context, organizationID, ledgerID, transactionID string) error {
	return nil
}
func (m *MockMidazClient) PostMultiSourceTransaction(ctx context.Context, organizationID, ledgerID string, tx *domain.MultiSourceTransactionPayload) (string, error) {
	return "", nil
}
func (m *MockMidazClient) ListExternalAccounts(ctx context.Context, organizationID, ledgerID string) ([]domain.ExternalAccount, error) {
	args := m.Called(ctx, organizationID, ledgerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ExternalAccount), args.Error(1)
}

func TestRouter_HealthAndManifest(t *testing.T) {
	cfg := &bootstrap.Config{
		PluginHost: "http://localhost:8082",
	}

	manifestHandler := NewManifestHandler(cfg)
	repo := new(MockRepo)
	accountHandler := NewAccountHandler(cfg, repo, nil, nil, nil)
	connectionHandler := NewConnectionHandler(cfg, repo, nil, nil, nil, nil)

	app := NewRouter(manifestHandler, accountHandler, connectionHandler, nil, nil)

	t.Run("GET /health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /ready", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /api/manifest", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/manifest", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var manifest map[string]any
		err = json.NewDecoder(resp.Body).Decode(&manifest)
		require.NoError(t, err)
		assert.Equal(t, "midaz-plugin-receivables", manifest["name"])
		assert.Equal(t, "reconciliation", manifest["category"])
	})
}

func TestRouter_AccountEndpoints(t *testing.T) {
	cfg := &bootstrap.Config{
		MidazOrganizationID: "org_1",
		MidazLedgerID:       "led_1",
	}

	repo := new(MockRepo)
	midazClient := new(MockMidazClient)
	manifestHandler := NewManifestHandler(cfg)
	accountHandler := NewAccountHandler(cfg, repo, nil, nil, midazClient)
	connectionHandler := NewConnectionHandler(cfg, repo, nil, nil, nil, nil)

	app := NewRouter(manifestHandler, accountHandler, connectionHandler, nil, nil)

	t.Run("GET /v1/external-accounts", func(t *testing.T) {
		midazClient.On("ListExternalAccounts", mock.Anything, "org_1", "led_1").
			Return([]domain.ExternalAccount{
				{ID: "ext_1", Alias: "@external/BRL:universal", Name: "Universal Fiat Clearing", AssetCode: "BRL", Type: "external"},
				{ID: "ext_2", Alias: "@external/BRL:custom", Name: "Custom Clearing", AssetCode: "BRL", Type: "external"},
			}, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/external-accounts", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res struct {
			Items []domain.ExternalAccount `json:"items"`
			Count int                      `json:"count"`
		}
		err = json.NewDecoder(resp.Body).Decode(&res)
		require.NoError(t, err)
		assert.Len(t, res.Items, 2)
		assert.Equal(t, 2, res.Count)
		assert.Equal(t, "@external/BRL:universal", res.Items[0].Alias)
	})

	t.Run("GET /v1/accounts/receivables - List", func(t *testing.T) {
		repo.On("ListReceivablesAccounts", mock.Anything).
			Return([]domain.ReceivablesAccount{
				{ID: "acc_1", AccountAlias: "@m1/receivables", Name: "Merchant 1"},
			}, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/accounts/receivables", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /v1/accounts/receivables/:id - Found", func(t *testing.T) {
		repo.On("FindReceivablesAccountByID", mock.Anything, "acc_1").
			Return(&domain.ReceivablesAccount{ID: "acc_1", AccountAlias: "@m1/receivables"}, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/accounts/receivables/acc_1", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /v1/accounts/receivables/:id - Not Found", func(t *testing.T) {
		repo.On("FindReceivablesAccountByID", mock.Anything, "missing").
			Return(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/accounts/receivables/missing", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestRouter_MatcherWebhookEndpoint(t *testing.T) {
	cfg := &bootstrap.Config{
		MidazOrganizationID: "org_1",
		MidazLedgerID:       "led_1",
	}

	repo := new(MockRepo)
	manifestHandler := NewManifestHandler(cfg)
	accountHandler := NewAccountHandler(cfg, repo, nil, nil, nil)
	connectionHandler := NewConnectionHandler(cfg, repo, nil, nil, nil, nil)

	matcherHandler := NewMatcherWebhookHandler(cfg, nil)
	app := NewRouter(manifestHandler, accountHandler, connectionHandler, matcherHandler, nil)

	payload := domain.MatcherWebhookPayload{
		Event:        "RECONCILIATION_RUN_COMPLETED",
		RunID:        "run_test_01",
		MatchesCount: 0,
		Matches:      []domain.SettlementMatch{},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/matcher", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
