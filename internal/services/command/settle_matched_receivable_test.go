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

type MockOpenFinanceNotifier struct {
	mock.Mock
}

func (m *MockOpenFinanceNotifier) MarkTransactionProcessed(ctx context.Context, bankTxID, midazTxCode string) error {
	args := m.Called(ctx, bankTxID, midazTxCode)
	return args.Error(0)
}

func TestSettleMatchedReceivableHandler_Execute(t *testing.T) {
	t.Run("Success D+X Multi-Source Settlement Transfer with Deductions and Open Finance Notification", func(t *testing.T) {
		repo := new(MockRepository)
		midazClient := new(MockMidazClient)
		notifier := new(MockOpenFinanceNotifier)

		pendingTx1 := "TX-PENDING-SUMUP-001"
		pendingTx2 := "TX-PENDING-PAGBANK-002"

		inst1 := &domain.Installment{
			ID:                   "inst_1",
			ReceivablesAccountID: "acc_rec_1",
			Acquirer:             domain.AcquirerSumUp,
			NSU:                  "SUMUP_NSU_100",
			InstallmentNumber:    1,
			NetAmount:            decimal.NewFromFloat(200.00),
			Status:               domain.InstallmentStatusPendingSettlement,
			MidazReceivableTxID:  &pendingTx1,
		}

		inst2 := &domain.Installment{
			ID:                   "inst_2",
			ReceivablesAccountID: "acc_rec_1",
			Acquirer:             domain.AcquirerPagBank,
			NSU:                  "PAGBANK_NSU_200",
			InstallmentNumber:    1,
			NetAmount:            decimal.NewFromFloat(100.00),
			Status:               domain.InstallmentStatusPendingSettlement,
			MidazReceivableTxID:  &pendingTx2,
		}

		acc := &domain.ReceivablesAccount{
			ID:                   "acc_rec_1",
			AccountAlias:         "@merchant_z/receivables",
			FeeAccountAlias:      "@merchant_z/receivables/fees",
			CheckingAccountAlias: "@merchant_z/bank_sicredi",
		}

		repo.On("FindInstallmentByReceivableCode", mock.Anything, domain.AcquirerSumUp, "SUMUP_NSU_100", 1).
			Return(inst1, nil)
		repo.On("FindInstallmentByReceivableCode", mock.Anything, domain.AcquirerPagBank, "PAGBANK_NSU_200", 1).
			Return(inst2, nil)

		repo.On("FindReceivablesAccountByID", mock.Anything, "acc_rec_1").
			Return(acc, nil).Twice()

		// Commit Pending Transactions in Midaz
		midazClient.On("CommitTransaction", mock.Anything, "org_1", "led_1", pendingTx1).
			Return(nil).Once()
		midazClient.On("CommitTransaction", mock.Anything, "org_1", "led_1", pendingTx2).
			Return(nil).Once()

		// Multi-source transfer with Deductions to Midaz
		midazClient.On("PostMultiSourceTransaction", mock.Anything, "org_1", "led_1", mock.MatchedBy(func(tx *domain.MultiSourceTransactionPayload) bool {
			return tx.Code == "OFP-SICREDI-BANKTX-9988" &&
				tx.DestinationAccountAlias == "@merchant_z/bank_sicredi" &&
				tx.TotalAmount.Equal(decimal.NewFromFloat(300.00)) &&
				len(tx.Sources) == 2 &&
				len(tx.Deductions) == 1 &&
				tx.Deductions[0].AccountAlias == "@merchant_z/receivables/fees" &&
				tx.Deductions[0].Amount.Equal(decimal.NewFromFloat(49.90))
		})).Return("MIDAZ-MULTI-TX-777", nil).Once()

		settleDate := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
		repo.On("MarkInstallmentReconciled", mock.Anything, "inst_1", "MIDAZ-MULTI-TX-777", "run_101", settleDate).
			Return(nil).Once()
		repo.On("MarkInstallmentReconciled", mock.Anything, "inst_2", "MIDAZ-MULTI-TX-777", "run_101", settleDate).
			Return(nil).Once()

		// Expect Open Finance notification call
		notifier.On("MarkTransactionProcessed", mock.Anything, "OFP-SICREDI-BANKTX-9988", "MIDAZ-MULTI-TX-777").
			Return(nil).Once()

		handler := NewSettleMatchedReceivableHandler(repo, midazClient).WithOpenFinanceNotifier(notifier)
		err := handler.Execute(context.Background(), SettleMatchedReceivableInput{
			OrganizationID:           "org_1",
			LedgerID:                 "led_1",
			MatchID:                  "match_abc_123",
			MatcherRunID:             "run_101",
			BankTransactionID:        "OFP-SICREDI-BANKTX-9988",
			TargetOpenFinanceAccount: "@merchant_z/bank_sicredi",
			TotalDepositAmount:       decimal.NewFromFloat(250.10),
			SettlementDate:           settleDate,
			Items: []MatchedReceivableItem{
				{
					SourceARecordID: "REC-SUMUP-SUMUP_NSU_100-1",
					Amount:          decimal.NewFromFloat(200.00),
				},
				{
					SourceARecordID: "REC-PAGBANK-PAGBANK_NSU_200-1",
					Amount:          decimal.NewFromFloat(100.00),
				},
			},
			Deductions: []SettlementDeductionItem{
				{
					FeeType:     "MACHINE_RENTAL",
					ChargedBy:   "SUMUP",
					Description: "Monthly POS terminal rental",
					Amount:      decimal.NewFromFloat(49.90),
				},
			},
		})

		require.NoError(t, err)
		repo.AssertExpectations(t)
		midazClient.AssertExpectations(t)
		notifier.AssertExpectations(t)
	})

	t.Run("Idempotent When All Items Already Reconciled", func(t *testing.T) {
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
			OrganizationID:    "org_1",
			LedgerID:          "led_1",
			BankTransactionID: "OFP-SICREDI-ALREADY",
			Items: []MatchedReceivableItem{
				{
					SourceARecordID: "REC-SUMUP-NSU_DONE-1",
				},
			},
		})

		require.NoError(t, err)
		midazClient.AssertNotCalled(t, "CommitTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		midazClient.AssertNotCalled(t, "PostMultiSourceTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Fails When No Items Provided", func(t *testing.T) {
		repo := new(MockRepository)
		midazClient := new(MockMidazClient)

		handler := NewSettleMatchedReceivableHandler(repo, midazClient)
		err := handler.Execute(context.Background(), SettleMatchedReceivableInput{
			OrganizationID: "org_1",
			LedgerID:       "led_1",
			Items:          []MatchedReceivableItem{},
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "no matched items provided")
	})
}
