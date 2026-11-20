package query

import (
	"context"
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
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

func TestGetCalendarHandler_Execute(t *testing.T) {
	repo := new(MockRepo)
	handler := NewGetCalendarHandler(repo)

	status := domain.InstallmentStatusPendingSettlement
	acquirer := domain.AcquirerSumUp
	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 9, 30, 23, 59, 59, 0, time.UTC)

	expected := []domain.Installment{
		{
			ID:                     "inst_cal_1",
			ReceivablesAccountID:   "acc_1",
			Acquirer:               domain.AcquirerSumUp,
			GrossAmount:            decimal.NewFromFloat(100.00),
			NetAmount:              decimal.NewFromFloat(97.00),
			ExpectedSettlementDate: startDate.AddDate(0, 0, 10),
			Status:                 domain.InstallmentStatusPendingSettlement,
		},
	}

	repo.On("GetReceivablesCalendar", mock.Anything, mock.MatchedBy(func(f domain.CalendarFilter) bool {
		return f.ReceivablesAccountID == "acc_1" &&
			*f.Acquirer == domain.AcquirerSumUp &&
			*f.Status == domain.InstallmentStatusPendingSettlement
	})).Return(expected, nil)

	installments, err := handler.Execute(context.Background(), GetCalendarInput{
		ReceivablesAccountID: "acc_1",
		Acquirer:             &acquirer,
		Status:               &status,
		StartDate:            &startDate,
		EndDate:              &endDate,
	})

	require.NoError(t, err)
	assert.Len(t, installments, 1)
	assert.Equal(t, "inst_cal_1", installments[0].ID)
	repo.AssertExpectations(t)
}
