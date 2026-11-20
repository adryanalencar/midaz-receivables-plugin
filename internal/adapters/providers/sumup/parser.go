package sumup

import (
	"fmt"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
)

type SumUpHistoryResponse struct {
	Items []SumUpTransaction `json:"items"`
}

type SumUpTransaction struct {
	ID                string             `json:"id"`
	TransactionCode   string             `json:"transaction_code"`
	Amount            float64            `json:"amount"`
	Currency          string             `json:"currency"`
	Timestamp         string             `json:"timestamp"`
	Status            string             `json:"status"` // "SUCCESSFUL", "CANCELLED", "FAILED"
	PaymentType       string             `json:"payment_type"`
	InstallmentsCount int                `json:"installments_count"`
	Card              *SumUpCard         `json:"card,omitempty"`
	FeeAmount         float64            `json:"fee_amount"`
	NetAmount         float64            `json:"net_amount"`
	PayoutPlan        string             `json:"payout_plan"`
	Events            []SumUpEventDetail `json:"events,omitempty"`
}

type SumUpCard struct {
	Type        string `json:"type"` // "CREDIT", "DEBIT"
	Brand       string `json:"brand"`
	Last4Digits string `json:"last_4_digits"`
}

type SumUpEventDetail struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Status    string  `json:"status"`
	Amount    float64 `json:"amount"`
	Timestamp string  `json:"timestamp"`
}

type SumUpPayoutItem struct {
	ID             string   `json:"id"`
	Date           string   `json:"date"` // YYYY-MM-DD
	Amount         float64  `json:"amount"`
	Fee            float64  `json:"fee"`
	NetAmount      float64  `json:"net_amount"`
	Currency       string   `json:"currency"`
	Status         string   `json:"status"`
	TransactionIDs []string `json:"transaction_ids"`
}

// NormalizeSale converts a raw SumUp transaction into the standardized domain Sale entity.
func NormalizeSale(tx *SumUpTransaction, connectionID, receivablesAccountID string) (*domain.Sale, error) {
	if tx == nil {
		return nil, fmt.Errorf("sumup: nil transaction payload")
	}

	saleTime, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		// Fallback parse without fractional seconds
		saleTime, err = time.Parse("2006-01-02T15:04:05Z", tx.Timestamp)
		if err != nil {
			saleTime = time.Now().UTC()
		}
	}

	gross := decimal.NewFromFloat(tx.Amount).Round(2)
	fee := decimal.NewFromFloat(tx.FeeAmount).Round(2)
	net := decimal.NewFromFloat(tx.NetAmount).Round(2)

	// Invariant: Gross - Fee == Net
	if !gross.Sub(fee).Equal(net) {
		net = gross.Sub(fee)
	}

	installmentsCount := tx.InstallmentsCount
	if installmentsCount <= 0 {
		installmentsCount = 1
	}

	cardBrand := ""
	paymentMethod := "CREDIT_CARD"
	if tx.Card != nil {
		cardBrand = tx.Card.Brand
		if tx.Card.Type == "DEBIT" {
			paymentMethod = "DEBIT_CARD"
		}
	}

	nsu := tx.TransactionCode
	if nsu == "" {
		nsu = tx.ID
	}

	sale := &domain.Sale{
		ConnectionID:         connectionID,
		ReceivablesAccountID: receivablesAccountID,
		Acquirer:             domain.AcquirerSumUp,
		NSU:                  nsu,
		TID:                  tx.ID,
		SaleDate:             saleTime,
		GrossAmount:          gross,
		TotalMDRFee:          fee,
		TotalNetAmount:       net,
		InstallmentsCount:    installmentsCount,
		PaymentMethod:        paymentMethod,
		CardBrand:            cardBrand,
		Installments:         make([]domain.Installment, installmentsCount),
		CreatedAt:            time.Now().UTC(),
	}

	// Calculate installment distribution
	countDec := decimal.NewFromInt(int64(installmentsCount))
	instGrossBase := gross.DivRound(countDec, 2)
	instFeeBase := fee.DivRound(countDec, 2)
	instNetBase := net.DivRound(countDec, 2)

	// Remainder cents allocated to installment 1
	grossRemainder := gross.Sub(instGrossBase.Mul(countDec))
	feeRemainder := fee.Sub(instFeeBase.Mul(countDec))
	netRemainder := net.Sub(instNetBase.Mul(countDec))

	for i := 0; i < installmentsCount; i++ {
		instNum := i + 1
		g := instGrossBase
		f := instFeeBase
		n := instNetBase

		if instNum == 1 {
			g = g.Add(grossRemainder)
			f = f.Add(feeRemainder)
			n = n.Add(netRemainder)
		}

		// Expected settlement: 30 days per installment
		settlementDays := 30 * instNum
		if paymentMethod == "DEBIT_CARD" {
			settlementDays = 1 // Debit card settles in D+1
		}
		expectedSettlement := saleTime.AddDate(0, 0, settlementDays)

		sale.Installments[i] = domain.Installment{
			ConnectionID:           connectionID,
			ReceivablesAccountID:   receivablesAccountID,
			Acquirer:               domain.AcquirerSumUp,
			NSU:                    nsu,
			InstallmentNumber:      instNum,
			GrossAmount:            g,
			MDRFee:                 f,
			NetAmount:              n,
			ExpectedSettlementDate: expectedSettlement,
			Status:                 domain.InstallmentStatusPendingSettlement,
		}
	}

	return sale, nil
}
