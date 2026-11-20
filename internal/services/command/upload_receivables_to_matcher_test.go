package command

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

func TestUploadReceivablesToMatcherHandler_Execute(t *testing.T) {
	t.Run("Success Uploading Pending Installments", func(t *testing.T) {
		repo := new(MockRepository)
		matcherClient := new(MockMatcherClient)

		matcherClient.On("EnsureContextAndSources", mock.Anything).
			Return("ctx_rec_1", "src_agenda", "src_bank", nil)

		now := time.Now().UTC()
		pendingInsts := []domain.Installment{
			{
				ID:                     "inst_1",
				ReceivablesAccountID:   "acc_1",
				ConnectionID:           "conn_1",
				Acquirer:               domain.AcquirerSumUp,
				NSU:                    "NSU_99",
				InstallmentNumber:      1,
				NetAmount:              decimal.NewFromFloat(97.00),
				ExpectedSettlementDate: now.AddDate(0, 0, 5),
				Status:                 domain.InstallmentStatusPendingSettlement,
			},
		}

		repo.On("ListPendingInstallmentsForMatcher", mock.Anything, mock.Anything, mock.Anything, 500).
			Return(pendingInsts, nil)

		matcherClient.On("UploadExpectedReceivables", mock.Anything, "ctx_rec_1", "src_agenda", mock.MatchedBy(func(recs []domain.MatcherSourceRecord) bool {
			return len(recs) == 1 &&
				recs[0].ID == "REC-SUMUP-NSU_99-1" &&
				recs[0].Amount.Equal(decimal.NewFromFloat(97.00))
		})).Return(nil)

		repo.On("MarkInstallmentUploadedToMatcher", mock.Anything, "inst_1", "REC-SUMUP-NSU_99-1").
			Return(nil)

		handler := NewUploadReceivablesToMatcherHandler(repo, matcherClient)
		uploadedCount, err := handler.Execute(context.Background(), UploadReceivablesToMatcherInput{
			WindowDays: 7,
			Limit:      500,
		})

		require.NoError(t, err)
		assert.Equal(t, 1, uploadedCount)
		repo.AssertExpectations(t)
		matcherClient.AssertExpectations(t)
	})

	t.Run("No Pending Installments", func(t *testing.T) {
		repo := new(MockRepository)
		matcherClient := new(MockMatcherClient)

		matcherClient.On("EnsureContextAndSources", mock.Anything).
			Return("ctx_rec_1", "src_agenda", "src_bank", nil)

		repo.On("ListPendingInstallmentsForMatcher", mock.Anything, mock.Anything, mock.Anything, 500).
			Return([]domain.Installment{}, nil)

		handler := NewUploadReceivablesToMatcherHandler(repo, matcherClient)
		uploadedCount, err := handler.Execute(context.Background(), UploadReceivablesToMatcherInput{})

		require.NoError(t, err)
		assert.Equal(t, 0, uploadedCount)
	})
}
