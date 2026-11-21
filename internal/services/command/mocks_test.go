package command

import (
	"context"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateReceivablesAccount(ctx context.Context, account *domain.ReceivablesAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockRepository) FindReceivablesAccountByID(ctx context.Context, id string) (*domain.ReceivablesAccount, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ReceivablesAccount), args.Error(1)
}

func (m *MockRepository) FindReceivablesAccountByAlias(ctx context.Context, alias string) (*domain.ReceivablesAccount, error) {
	args := m.Called(ctx, alias)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ReceivablesAccount), args.Error(1)
}

func (m *MockRepository) ListReceivablesAccounts(ctx context.Context) ([]domain.ReceivablesAccount, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.ReceivablesAccount), args.Error(1)
}

func (m *MockRepository) CreateAcquirerConnection(ctx context.Context, conn *domain.AcquirerConnection) error {
	args := m.Called(ctx, conn)
	return args.Error(0)
}

func (m *MockRepository) FindAcquirerConnectionByID(ctx context.Context, id string) (*domain.AcquirerConnection, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AcquirerConnection), args.Error(1)
}

func (m *MockRepository) ListConnectionsByAccountID(ctx context.Context, accountID string) ([]domain.AcquirerConnection, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).([]domain.AcquirerConnection), args.Error(1)
}

func (m *MockRepository) ListActiveConnectionsForSync(ctx context.Context) ([]domain.AcquirerConnection, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.AcquirerConnection), args.Error(1)
}

func (m *MockRepository) UpdateConnectionSyncState(ctx context.Context, id string, syncStatus string, syncError *string, lastSyncAt time.Time) error {
	args := m.Called(ctx, id, syncStatus, syncError, lastSyncAt)
	return args.Error(0)
}

func (m *MockRepository) UpdateConnectionStatus(ctx context.Context, id string, status domain.ConnectionStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockRepository) SaveSaleWithInstallments(ctx context.Context, sale *domain.Sale) error {
	args := m.Called(ctx, sale)
	return args.Error(0)
}

func (m *MockRepository) FindSaleByNSU(ctx context.Context, connectionID, nsu string) (*domain.Sale, error) {
	args := m.Called(ctx, connectionID, nsu)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Sale), args.Error(1)
}

func (m *MockRepository) FindInstallmentByReceivableCode(ctx context.Context, acquirer domain.AcquirerType, nsu string, installmentNumber int) (*domain.Installment, error) {
	args := m.Called(ctx, acquirer, nsu, installmentNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Installment), args.Error(1)
}

func (m *MockRepository) ListPendingInstallmentsForMatcher(ctx context.Context, windowStart, windowEnd time.Time, limit int) ([]domain.Installment, error) {
	args := m.Called(ctx, windowStart, windowEnd, limit)
	return args.Get(0).([]domain.Installment), args.Error(1)
}

func (m *MockRepository) MarkInstallmentUploadedToMatcher(ctx context.Context, installmentID, matcherRecordID string) error {
	args := m.Called(ctx, installmentID, matcherRecordID)
	return args.Error(0)
}

func (m *MockRepository) UpdateInstallmentReceivableTxID(ctx context.Context, installmentID, txID string) error {
	args := m.Called(ctx, installmentID, txID)
	return args.Error(0)
}

func (m *MockRepository) MarkInstallmentReconciled(ctx context.Context, installmentID, midazSettlementTxID, matcherRunID string, settlementDate time.Time) error {
	args := m.Called(ctx, installmentID, midazSettlementTxID, matcherRunID, settlementDate)
	return args.Error(0)
}

func (m *MockRepository) GetReceivablesCalendar(ctx context.Context, filter domain.CalendarFilter) ([]domain.Installment, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domain.Installment), args.Error(1)
}

type MockMatcherClient struct {
	mock.Mock
}

func (m *MockMatcherClient) EnsureContextAndSources(ctx context.Context) (string, string, string, error) {
	args := m.Called(ctx)
	return args.String(0), args.String(1), args.String(2), args.Error(3)
}

func (m *MockMatcherClient) UploadExpectedReceivables(ctx context.Context, contextID, sourceAID string, records []domain.MatcherSourceRecord) error {
	args := m.Called(ctx, contextID, sourceAID, records)
	return args.Error(0)
}

func (m *MockMatcherClient) TriggerReconciliationRun(ctx context.Context, contextID string) (string, error) {
	args := m.Called(ctx, contextID)
	return args.String(0), args.Error(1)
}

func (m *MockMatcherClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockMidazClient struct {
	mock.Mock
}

func (m *MockMidazClient) EnsureClearingAccount(ctx context.Context, organizationID, ledgerID string, acquirer domain.AcquirerType) (string, error) {
	args := m.Called(ctx, organizationID, ledgerID, acquirer)
	return args.String(0), args.Error(1)
}

func (m *MockMidazClient) EnsureAcquirerInternalAccount(ctx context.Context, organizationID, ledgerID string, acquirer domain.AcquirerType) (string, error) {
	args := m.Called(ctx, organizationID, ledgerID, acquirer)
	return args.String(0), args.Error(1)
}

func (m *MockMidazClient) CreateAccount(ctx context.Context, organizationID, ledgerID, alias, name, accountType string, parentAccountID *string) (string, error) {
	args := m.Called(ctx, organizationID, ledgerID, alias, name, accountType, parentAccountID)
	return args.String(0), args.Error(1)
}

func (m *MockMidazClient) CheckAccountExists(ctx context.Context, organizationID, ledgerID, alias string) (bool, error) {
	args := m.Called(ctx, organizationID, ledgerID, alias)
	return args.Bool(0), args.Error(1)
}

func (m *MockMidazClient) PostTransaction(ctx context.Context, organizationID, ledgerID string, tx *domain.MidazTransactionPayload) (string, error) {
	args := m.Called(ctx, organizationID, ledgerID, tx)
	return args.String(0), args.Error(1)
}

func (m *MockMidazClient) CommitTransaction(ctx context.Context, organizationID, ledgerID, transactionID string) error {
	args := m.Called(ctx, organizationID, ledgerID, transactionID)
	return args.Error(0)
}

func (m *MockMidazClient) PostMultiSourceTransaction(ctx context.Context, organizationID, ledgerID string, tx *domain.MultiSourceTransactionPayload) (string, error) {
	args := m.Called(ctx, organizationID, ledgerID, tx)
	return args.String(0), args.Error(1)
}

func (m *MockMidazClient) ListExternalAccounts(ctx context.Context, organizationID, ledgerID string) ([]domain.ExternalAccount, error) {
	args := m.Called(ctx, organizationID, ledgerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ExternalAccount), args.Error(1)
}


