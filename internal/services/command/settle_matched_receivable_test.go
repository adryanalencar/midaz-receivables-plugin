package command

import (
	"context"
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSettleMatchedReceivableHandler_Execute(t *testing.T) {
	t.Run("Success D+X Settlement Extinguishment", func(t *testing.T) {
		repo := new(MockRepository)
		midazClient := new(MockMidazClient)

		inst := &domain.Installment{
			ID:                   "inst_999",
			ReceivablesAccountID: "acc_rec_1",
			Acquirer:             domain.AcquirerSumUp,
			NSU:                  "NSU_987654",
			InstallmentNumber:    1,
			NetAmount:            decimal.NewFromFloat(97.00),
			Status:               domain.InstallmentStatusPendingSettlement,
		}

		acc := &domain.ReceivablesAccount{
			ID:                   "acc_rec_1",
			AccountAlias:         "@merchant_z/receivables",
			CheckingAccountAlias: "@merchant_z/checking",
		}

		repo.On("FindInstallmentByReceivableCode", mock.Anything, domain.AcquirerSumUp, "NSU_987654", 1).
			Return(inst, nil)
		repo.On("FindReceivablesAccountByID", mock.Anything, "acc_rec_1").
			Return(acc, nil)

		// Expect Midaz Settlement Transaction
		midazClient.On("PostTransaction", mock.Anything, "org_1", "led_1", mock.MatchedBy(func(tx *domain.MidazTransactionPayload) bool {
			return tx.Code == "SETTLE-SUMUP-NSU_987654-1" &&
				tx.SourceAccountAlias == "@merchant_z/receivables" &&
				tx.DestinationAccountAlias == "@merchant_z/checking" &&
				tx.Amount.Equal(decimal.NewFromFloat(97.00)) &&
				tx.ChartOfAccountsGroupName == "RECEIVABLES_SETTLEMENT"
		})).Return("SETTLE-TX-ID-999", nil)

		settleDate := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
		repo.On("MarkInstallmentReconciled", mock.Anything, "inst_999", "SETTLE-TX-ID-999", "run_101", settleDate).
			Return(nil)

		handler := NewSettleMatchedReceivableHandler(repo, midazClient)
		err := handler.Execute(context.Background(), SettleMatchedReceivableInput{
			OrganizationID:  "org_1",
			LedgerID:        "led_1",
			MatchID:         "m_1",
			SourceARecordID: "REC-SUMUP-NSU_987654-1",
			SourceBRecordID: "OFP-ITAU-9988",
			SettledAmount:   decimal.NewFromFloat(97.00),
			SettlementDate:  settleDate,
			MatcherRunID:    "run_101",
		})

		require.NoError(t, err)
		repo.AssertExpectations(t)
		midazClient.AssertExpectations(t)
	})

	t.Run("Idempotent When Already Reconciled", func(t *testing.T) {
		repo := new(MockRepository)
		midazClient := new(MockMidazClient)

		inst := &domain.Installment{
			ID:                "inst_already_done",
			Acquirer:          domain.AcquirerSumUp,
			NSU:               "NSU_DONE",
			InstallmentNumber: 1,
			Status:            domain.InstallmentStatusReconciled,
		}

		repo.On("FindInstallmentByReceivableCode", mock.Anything, domain.AcquirerSumUp, "NSU_DONE", 1).
			Return(inst, nil)

		handler := NewSettleMatchedReceivableHandler(repo, midazClient)
		err := handler.Execute(context.Background(), SettleMatchedReceivableInput{
			SourceARecordID: "REC-SUMUP-NSU_DONE-1",
		})

		require.NoError(t, err)
		// Midaz PostTransaction must NOT be called
		midazClient.AssertNotCalled(t, "PostTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}
