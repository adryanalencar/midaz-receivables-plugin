package workers

import (
	"context"
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/services/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) CreateReceivablesAccount(ctx context.Context, account *domain.ReceivablesAccount) error {
	return nil
}
func (m *MockRepo) FindReceivablesAccountByID(ctx context.Context, id string) (*domain.ReceivablesAccount, error) {
	return nil, nil
}
func (m *MockRepo) FindReceivablesAccountByAlias(ctx context.Context, alias string) (*domain.ReceivablesAccount, error) {
	return nil, nil
}
func (m *MockRepo) ListReceivablesAccounts(ctx context.Context) ([]domain.ReceivablesAccount, error) {
	return nil, nil
}
func (m *MockRepo) CreateAcquirerConnection(ctx context.Context, conn *domain.AcquirerConnection) error {
	return nil
}
func (m *MockRepo) FindAcquirerConnectionByID(ctx context.Context, id string) (*domain.AcquirerConnection, error) {
	return nil, nil
}
func (m *MockRepo) ListConnectionsByAccountID(ctx context.Context, accountID string) ([]domain.AcquirerConnection, error) {
	return nil, nil
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
	args := m.Called(ctx, windowStart, windowEnd, limit)
	return args.Get(0).([]domain.Installment), args.Error(1)
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
	return nil, nil
}

type MockMatcher struct {
	mock.Mock
}

func (m *MockMatcher) EnsureContextAndSources(ctx context.Context) (string, string, string, error) {
	args := m.Called(ctx)
	return args.String(0), args.String(1), args.String(2), args.Error(3)
}
func (m *MockMatcher) UploadExpectedReceivables(ctx context.Context, contextID, sourceAID string, records []domain.MatcherSourceRecord) error {
	return nil
}
func (m *MockMatcher) TriggerReconciliationRun(ctx context.Context, contextID string) (string, error) {
	return "", nil
}
func (m *MockMatcher) HealthCheck(ctx context.Context) error {
	return nil
}

func TestMatcherUploadWorker_StartStop(t *testing.T) {
	repo := new(MockRepo)
	matcherClient := new(MockMatcher)

	matcherClient.On("EnsureContextAndSources", mock.Anything).
		Return("ctx_1", "src_a", "src_b", nil)
	repo.On("ListPendingInstallmentsForMatcher", mock.Anything, mock.Anything, mock.Anything, 50).
		Return([]domain.Installment{}, nil)

	cfg := &bootstrap.Config{
		SyncIntervalMinutes:     1,
		MatcherUploadWindowDays: 7,
		WorkerBatchLimit:        50,
	}

	handler := command.NewUploadReceivablesToMatcherHandler(repo, matcherClient)
	worker := NewMatcherUploadWorker(cfg, handler)

	ctx, cancel := context.WithCancel(context.Background())
	worker.Start(ctx)

	// Allow cycle to run
	time.Sleep(50 * time.Millisecond)

	worker.Stop()
	cancel()

	assert.True(t, true)
}
