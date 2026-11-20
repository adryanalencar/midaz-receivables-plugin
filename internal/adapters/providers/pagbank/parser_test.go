package pagbank

import (
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPagBank_NormalizeSale_SingleInstallment(t *testing.T) {
	charge := &PagBankCharge{
		ID:          "CHAR_PAG_001",
		ReferenceID: "REF_001",
		Status:      "PAID",
		PaidAt:      "2026-04-10T11:00:00Z",
		Amount: PagBankAmount{
			Value:    25000, // R$ 250.00
			Currency: "BRL",
		},
		PaymentMethod: PagBankPaymentMethod{
			Type:         "CREDIT_CARD",
			Installments: 1,
			Card: &PagBankCard{
				Brand:      "MASTERCARD",
				LastDigits: "9988",
			},
		},
		PaymentResponse: PagBankPaymentResponse{
			Code:      "20000",
			Message:   "SUCESSO",
			Reference: "NSU_PAG_001",
		},
		Fees: &PagBankFees{
			Seller: 625, // R$ 6.25 (2.5%)
		},
	}

	sale, err := NormalizeSale(charge, "conn_pag_1", "acc_pag_1")
	require.NoError(t, err)
	require.NotNil(t, sale)

	assert.Equal(t, "conn_pag_1", sale.ConnectionID)
	assert.Equal(t, "acc_pag_1", sale.ReceivablesAccountID)
	assert.Equal(t, domain.AcquirerPagBank, sale.Acquirer)
	assert.Equal(t, "NSU_PAG_001", sale.NSU)
	assert.Equal(t, "CHAR_PAG_001", sale.TID)
	assert.True(t, sale.GrossAmount.Equal(decimal.NewFromFloat(250.00)))
	assert.True(t, sale.TotalMDRFee.Equal(decimal.NewFromFloat(6.25)))
	assert.True(t, sale.TotalNetAmount.Equal(decimal.NewFromFloat(243.75)))
	assert.Equal(t, 1, sale.InstallmentsCount)
	assert.Equal(t, "CREDIT_CARD", sale.PaymentMethod)
	assert.Equal(t, "MASTERCARD", sale.CardBrand)

	require.Len(t, sale.Installments, 1)
	assert.True(t, sale.Installments[0].GrossAmount.Equal(decimal.NewFromFloat(250.00)))
	assert.True(t, sale.Installments[0].MDRFee.Equal(decimal.NewFromFloat(6.25)))
	assert.True(t, sale.Installments[0].NetAmount.Equal(decimal.NewFromFloat(243.75)))
}

func TestPagBank_NormalizeSale_OddCentsMultiInstallment(t *testing.T) {
	// R$ 100.00 in 3 installments, fee R$ 3.01
	charge := &PagBankCharge{
		ID:          "CHAR_ODD_002",
		ReferenceID: "REF_ODD_002",
		Status:      "PAID",
		PaidAt:      "2026-05-01T10:00:00Z",
		Amount: PagBankAmount{
			Value: 10000, // 100.00
		},
		PaymentMethod: PagBankPaymentMethod{
			Type:         "CREDIT_CARD",
			Installments: 3,
		},
		PaymentResponse: PagBankPaymentResponse{
			Reference: "NSU_ODD_PAG",
		},
		Fees: &PagBankFees{
			Seller: 301, // 3.01
		},
	}

	sale, err := NormalizeSale(charge, "conn_1", "acc_1")
	require.NoError(t, err)
	require.Len(t, sale.Installments, 3)

	var sumGross, sumFee, sumNet decimal.Decimal
	for _, inst := range sale.Installments {
		sumGross = sumGross.Add(inst.GrossAmount)
		sumFee = sumFee.Add(inst.MDRFee)
		sumNet = sumNet.Add(inst.NetAmount)
	}

	assert.True(t, sumGross.Equal(decimal.NewFromFloat(100.00)))
	assert.True(t, sumFee.Equal(decimal.NewFromFloat(3.01)))
	assert.True(t, sumNet.Equal(decimal.NewFromFloat(96.99)))

	// Remainder cent on installment 1
	assert.True(t, sale.Installments[0].GrossAmount.Equal(decimal.RequireFromString("33.34")))
	assert.True(t, sale.Installments[1].GrossAmount.Equal(decimal.RequireFromString("33.33")))
	assert.True(t, sale.Installments[2].GrossAmount.Equal(decimal.RequireFromString("33.33")))
}

func TestPagBank_NormalizeSale_DebitCardDPlus1(t *testing.T) {
	saleTime := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	charge := &PagBankCharge{
		ID:        "CHAR_DEBIT_001",
		Status:    "PAID",
		PaidAt:    saleTime.Format(time.RFC3339),
		Amount:    PagBankAmount{Value: 5000},
		PaymentMethod: PagBankPaymentMethod{
			Type:         "DEBIT_CARD",
			Installments: 1,
		},
	}

	sale, err := NormalizeSale(charge, "conn", "acc")
	require.NoError(t, err)
	assert.Equal(t, "DEBIT_CARD", sale.PaymentMethod)
	assert.Equal(t, saleTime.AddDate(0, 0, 1), sale.Installments[0].ExpectedSettlementDate)
}

func TestPagBank_NormalizeSale_Nil(t *testing.T) {
	sale, err := NormalizeSale(nil, "conn", "acc")
	assert.Error(t, err)
	assert.Nil(t, sale)
}
