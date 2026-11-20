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

func TestIngestAcquirerSaleHandler_Execute(t *testing.T) {
	t.Run("Success 2 Installments D+0 Ingestion with Pending Status", func(t *testing.T) {
		repo := new(MockRepository)
		midazClient := new(MockMidazClient)

		conn := &domain.AcquirerConnection{
			ID:                   "conn_100",
			ReceivablesAccountID: "acc_100",
			AcquirerType:         domain.AcquirerSumUp,
		}
		acc := &domain.ReceivablesAccount{
			ID:           "acc_100",
			AccountAlias: "@merchant_b/receivables",
		}

		repo.On("FindAcquirerConnectionByID", mock.Anything, "conn_100").Return(conn, nil)
		repo.On("FindReceivablesAccountByID", mock.Anything, "acc_100").Return(acc, nil)

		saleDate := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
		sale := &domain.Sale{
			ConnectionID:      "conn_100",
			Acquirer:          domain.AcquirerSumUp,
			NSU:               "NSU_TEST_001",
			TID:               "TID_TEST_001",
			SaleDate:          saleDate,
			GrossAmount:       decimal.NewFromFloat(200.00),
			TotalMDRFee:       decimal.NewFromFloat(6.00),
			TotalNetAmount:    decimal.NewFromFloat(194.00),
			InstallmentsCount: 2,
			Installments: []domain.Installment{
				{
					ID:                     "inst_1",
					InstallmentNumber:      1,
					GrossAmount:            decimal.NewFromFloat(100.00),
					MDRFee:                 decimal.NewFromFloat(3.00),
					NetAmount:              decimal.NewFromFloat(97.00),
					ExpectedSettlementDate: saleDate.AddDate(0, 0, 30),
				},
				{
					ID:                     "inst_2",
					InstallmentNumber:      2,
					GrossAmount:            decimal.NewFromFloat(100.00),
					MDRFee:                 decimal.NewFromFloat(3.00),
					NetAmount:              decimal.NewFromFloat(97.00),
					ExpectedSettlementDate: saleDate.AddDate(0, 0, 60),
				},
			},
		}

		repo.On("SaveSaleWithInstallments", mock.Anything, sale).Return(nil)

		// Expect Midaz pending transaction for installment 1
		midazClient.On("PostTransaction", mock.Anything, "org_1", "led_1", mock.MatchedBy(func(tx *domain.MidazTransactionPayload) bool {
			return tx.Code == "REC-SUMUP-NSU_TEST_001-1" &&
				tx.SourceAccountAlias == "@external/acquirer_sumup" &&
				tx.DestinationAccountAlias == "@merchant_b/receivables" &&
				tx.Amount.Equal(decimal.NewFromFloat(100.00)) &&
				tx.Pending == true
		})).Return("midaz_tx_101", nil)

		repo.On("UpdateInstallmentReceivableTxID", mock.Anything, "inst_1", "midaz_tx_101").Return(nil)

		// Expect Midaz pending transaction for installment 2
		midazClient.On("PostTransaction", mock.Anything, "org_1", "led_1", mock.MatchedBy(func(tx *domain.MidazTransactionPayload) bool {
			return tx.Code == "REC-SUMUP-NSU_TEST_001-2" &&
				tx.SourceAccountAlias == "@external/acquirer_sumup" &&
				tx.DestinationAccountAlias == "@merchant_b/receivables" &&
				tx.Amount.Equal(decimal.NewFromFloat(100.00)) &&
				tx.Pending == true
		})).Return("midaz_tx_102", nil)

		repo.On("UpdateInstallmentReceivableTxID", mock.Anything, "inst_2", "midaz_tx_102").Return(nil)

		handler := NewIngestAcquirerSaleHandler(repo, midazClient)
		err := handler.Execute(context.Background(), IngestAcquirerSaleInput{
			OrganizationID: "org_1",
			LedgerID:       "led_1",
			Sale:           sale,
		})

		require.NoError(t, err)
		repo.AssertExpectations(t)
		midazClient.AssertExpectations(t)
	})

	t.Run("Validation Errors", func(t *testing.T) {
		handler := NewIngestAcquirerSaleHandler(nil, nil)

		err := handler.Execute(context.Background(), IngestAcquirerSaleInput{Sale: nil})
		assert.ErrorIs(t, err, ErrNilSalePayload)

		err = handler.Execute(context.Background(), IngestAcquirerSaleInput{
			Sale: &domain.Sale{ConnectionID: ""},
		})
		assert.ErrorIs(t, err, ErrEmptyConnectionID)

		err = handler.Execute(context.Background(), IngestAcquirerSaleInput{
			Sale: &domain.Sale{
				ConnectionID: "c1",
				GrossAmount:  decimal.Zero,
			},
		})
		assert.ErrorIs(t, err, ErrZeroGrossAmount)
	})
}
