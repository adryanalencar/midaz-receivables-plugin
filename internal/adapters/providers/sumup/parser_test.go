package sumup

import (
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeSale_SingleInstallment(t *testing.T) {
	tx := &SumUpTransaction{
		ID:                "tx_sumup_001",
		TransactionCode:   "NSU_SUMUP_001",
		Amount:            150.00,
		FeeAmount:         3.50,
		NetAmount:         146.50,
		InstallmentsCount: 1,
		Timestamp:         "2026-03-15T14:30:00Z",
		Status:            "SUCCESSFUL",
		PaymentType:       "ECOM",
		Card: &SumUpCard{
			Brand: "MASTERCARD",
			Type:  "CREDIT",
		},
	}

	sale, err := NormalizeSale(tx, "conn_123", "rec_acc_456")
	require.NoError(t, err)
	require.NotNil(t, sale)

	assert.Equal(t, "conn_123", sale.ConnectionID)
	assert.Equal(t, "rec_acc_456", sale.ReceivablesAccountID)
	assert.Equal(t, domain.AcquirerSumUp, sale.Acquirer)
	assert.Equal(t, "NSU_SUMUP_001", sale.NSU)
	assert.Equal(t, "tx_sumup_001", sale.TID)
	assert.True(t, sale.GrossAmount.Equal(decimal.NewFromFloat(150.00)))
	assert.True(t, sale.TotalMDRFee.Equal(decimal.NewFromFloat(3.50)))
	assert.True(t, sale.TotalNetAmount.Equal(decimal.NewFromFloat(146.50)))
	assert.Equal(t, 1, sale.InstallmentsCount)
	assert.Equal(t, "CREDIT_CARD", sale.PaymentMethod)
	assert.Equal(t, "MASTERCARD", sale.CardBrand)

	require.Len(t, sale.Installments, 1)
	inst := sale.Installments[0]
	assert.Equal(t, 1, inst.InstallmentNumber)
	assert.True(t, inst.GrossAmount.Equal(decimal.NewFromFloat(150.00)))
	assert.True(t, inst.MDRFee.Equal(decimal.NewFromFloat(3.50)))
	assert.True(t, inst.NetAmount.Equal(decimal.NewFromFloat(146.50)))
	assert.Equal(t, domain.InstallmentStatusPendingSettlement, inst.Status)
}

func TestNormalizeSale_OddCentsDistribution(t *testing.T) {
	// Gross: 100.00, Fee: 3.01, Net: 96.99 over 3 installments
	tx := &SumUpTransaction{
		ID:                "tx_odd_002",
		TransactionCode:   "NSU_ODD_002",
		Amount:            100.00,
		FeeAmount:         3.01,
		NetAmount:         96.99,
		InstallmentsCount: 3,
		Timestamp:         "2026-04-01T10:00:00Z",
		Status:            "SUCCESSFUL",
		Card: &SumUpCard{
			Brand: "VISA",
			Type:  "CREDIT",
		},
	}

	sale, err := NormalizeSale(tx, "conn_1", "acc_1")
	require.NoError(t, err)
	require.Len(t, sale.Installments, 3)

	// Sum of installments must exactly match total
	var sumGross, sumFee, sumNet decimal.Decimal
	for _, inst := range sale.Installments {
		sumGross = sumGross.Add(inst.GrossAmount)
		sumFee = sumFee.Add(inst.MDRFee)
		sumNet = sumNet.Add(inst.NetAmount)
	}

	assert.True(t, sumGross.Equal(decimal.NewFromFloat(100.00)), "Gross sum must be 100.00, got %s", sumGross)
	assert.True(t, sumFee.Equal(decimal.NewFromFloat(3.01)), "Fee sum must be 3.01, got %s", sumFee)
	assert.True(t, sumNet.Equal(decimal.NewFromFloat(96.99)), "Net sum must be 96.99, got %s", sumNet)

	// Remainder cent must be on installment 1
	// 100.00 / 3 = 33.33 each, remainder 0.01 -> inst 1 is 33.34
	assert.True(t, sale.Installments[0].GrossAmount.Equal(decimal.RequireFromString("33.34")))
	assert.True(t, sale.Installments[1].GrossAmount.Equal(decimal.RequireFromString("33.33")))
	assert.True(t, sale.Installments[2].GrossAmount.Equal(decimal.RequireFromString("33.33")))

	// Invariant: Gross - MDRFee == NetAmount for each installment
	for i, inst := range sale.Installments {
		diff := inst.GrossAmount.Sub(inst.MDRFee)
		assert.True(t, diff.Equal(inst.NetAmount), "Installment %d gross - fee (%s) must equal net (%s)", i+1, diff, inst.NetAmount)
	}
}

func TestNormalizeSale_DebitCardDPlus1(t *testing.T) {
	saleTime := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	tx := &SumUpTransaction{
		ID:                "tx_debit_003",
		Amount:            50.00,
		FeeAmount:         1.00,
		NetAmount:         49.00,
		InstallmentsCount: 1,
		Timestamp:         saleTime.Format(time.RFC3339),
		Status:            "SUCCESSFUL",
		Card: &SumUpCard{
			Brand: "ELO",
			Type:  "DEBIT",
		},
	}

	sale, err := NormalizeSale(tx, "conn_1", "acc_1")
	require.NoError(t, err)
	assert.Equal(t, "DEBIT_CARD", sale.PaymentMethod)

	require.Len(t, sale.Installments, 1)
	expectedSettlement := saleTime.AddDate(0, 0, 1)
	assert.Equal(t, expectedSettlement, sale.Installments[0].ExpectedSettlementDate)
}

func TestNormalizeSale_NilTransaction(t *testing.T) {
	sale, err := NormalizeSale(nil, "conn", "acc")
	assert.Error(t, err)
	assert.Nil(t, sale)
}
