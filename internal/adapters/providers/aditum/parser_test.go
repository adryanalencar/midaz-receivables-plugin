package aditum

import (
	"testing"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeSale_SingleInstallment(t *testing.T) {
	charge := &AditumChargeItem{
		ID:           "chg-123",
		NSU:          "100200300",
		ChargeStatus: "Authorized",
		MerchantID:   "merc-999",
		DataCobranca: "2026-03-20T14:30:00",
		Transactions: []AditumTransactionItem{
			{
				TransactionID:     "tx-456",
				AuthorizationCode: "AUTH123",
				Amount:            10000, // R$ 100.00
				FeeAmount:         300,   // R$ 3.00
				NetAmount:         9700,  // R$ 97.00
				IsCapture:         true,
				PaymentType:       "credit",
				InstallmentNumber: 1,
				Card: &AditumCardItem{
					Brand:     "Mastercard",
					BrandName: "MASTERCARD",
				},
			},
		},
	}

	sale, err := NormalizeSale(charge, "conn-1", "acc-1")
	require.NoError(t, err)
	require.NotNil(t, sale)

	assert.Equal(t, domain.AcquirerAditum, sale.Acquirer)
	assert.Equal(t, "100200300", sale.NSU)
	assert.Equal(t, "tx-456", sale.TID)
	assert.Equal(t, "AUTH123", sale.AuthorizationCode)
	assert.Equal(t, "CREDIT_CARD", sale.PaymentMethod)
	assert.Equal(t, "MASTERCARD", sale.CardBrand)
	assert.True(t, decimal.NewFromInt(100).Equal(sale.GrossAmount))
	assert.True(t, decimal.NewFromInt(3).Equal(sale.TotalMDRFee))
	assert.True(t, decimal.NewFromInt(97).Equal(sale.TotalNetAmount))
	assert.Equal(t, 1, sale.InstallmentsCount)
	assert.Len(t, sale.Installments, 1)

	inst := sale.Installments[0]
	assert.Equal(t, 1, inst.InstallmentNumber)
	assert.True(t, decimal.NewFromInt(100).Equal(inst.GrossAmount))
	assert.True(t, decimal.NewFromInt(3).Equal(inst.MDRFee))
	assert.True(t, decimal.NewFromInt(97).Equal(inst.NetAmount))
	assert.Equal(t, domain.InstallmentStatusPendingSettlement, inst.Status)
}

func TestNormalizeSale_MultiInstallment(t *testing.T) {
	charge := &AditumChargeItem{
		ID:           "chg-999",
		NSU:          "987654321",
		DataCobranca: "2026-03-01T10:00:00Z",
		Transactions: []AditumTransactionItem{
			{
				TransactionID:     "tx-multi",
				Amount:            10000, // R$ 100.00
				FeeAmount:         600,   // R$ 6.00
				NetAmount:         9400,  // R$ 94.00
				IsCapture:         true,
				PaymentType:       "credit",
				InstallmentNumber: 3,
			},
		},
	}

	sale, err := NormalizeSale(charge, "conn-2", "acc-2")
	require.NoError(t, err)
	require.NotNil(t, sale)

	assert.Equal(t, 3, sale.InstallmentsCount)
	assert.Len(t, sale.Installments, 3)

	var totalGross decimal.Decimal
	var totalFee decimal.Decimal
	var totalNet decimal.Decimal
	for _, inst := range sale.Installments {
		totalGross = totalGross.Add(inst.GrossAmount)
		totalFee = totalFee.Add(inst.MDRFee)
		totalNet = totalNet.Add(inst.NetAmount)
	}

	assert.True(t, sale.GrossAmount.Equal(totalGross))
	assert.True(t, sale.TotalMDRFee.Equal(totalFee))
	assert.True(t, sale.TotalNetAmount.Equal(totalNet))
}

func TestNormalizeSale_DebitCard(t *testing.T) {
	charge := &AditumChargeItem{
		ID:  "chg-debit",
		NSU: "555444333",
		Transactions: []AditumTransactionItem{
			{
				TransactionID:     "tx-deb",
				Amount:            5000, // R$ 50.00
				FeeAmount:         100,  // R$ 1.00
				NetAmount:         4900, // R$ 49.00
				IsCapture:         true,
				PaymentType:       "debit",
				InstallmentNumber: 1,
			},
		},
	}

	sale, err := NormalizeSale(charge, "conn-3", "acc-3")
	require.NoError(t, err)
	require.NotNil(t, sale)

	assert.Equal(t, "DEBIT_CARD", sale.PaymentMethod)
	require.Len(t, sale.Installments, 1)

	expectedDate := sale.SaleDate.AddDate(0, 0, 1).Format("2006-01-02")
	assert.Equal(t, expectedDate, sale.Installments[0].ExpectedSettlementDate.Format("2006-01-02"))
}

func TestNormalizeWebhook(t *testing.T) {
	webhookJSON := []byte(`{
		"ChargeId": "charge-webhook-1",
		"Nsu": "nsu-webhook-100",
		"ChargeStatus": "Authorized",
		"MerchantId": "merchant-aditum-1",
		"CreationDate": "2026-03-21T18:00:00",
		"Transactions": [
			{
				"TransactionId": "tid-webhook-1",
				"AuthorizationCode": "AUTH999",
				"Amount": 25050,
				"PaymentType": "credit",
				"InstallmentNumber": 1,
				"IsCapture": true,
				"Card": {
					"BrandName": "VISA"
				}
			}
		]
	}`)

	sale, err := NormalizeWebhook(webhookJSON, "conn-hook", "acc-hook")
	require.NoError(t, err)
	require.NotNil(t, sale)

	assert.Equal(t, "nsu-webhook-100", sale.NSU)
	assert.Equal(t, "tid-webhook-1", sale.TID)
	assert.Equal(t, "VISA", sale.CardBrand)
	assert.True(t, decimal.NewFromFloat(250.50).Equal(sale.GrossAmount))
}

func TestParseCentsToDecimal(t *testing.T) {
	dec := ParseCentsToDecimal(1999)
	assert.True(t, decimal.NewFromFloat(19.99).Equal(dec))

	decStr := ParseStringCentsToDecimal("00000025000")
	assert.True(t, decimal.NewFromFloat(250.00).Equal(decStr))
}
