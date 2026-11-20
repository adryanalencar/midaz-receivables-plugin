package pagbank

import (
	"fmt"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
)

type PagBankOrdersResponse struct {
	Orders []PagBankOrder `json:"orders"`
}

type PagBankChargesResponse struct {
	Charges []PagBankCharge `json:"charges"`
}

type PagBankOrder struct {
	ID          string          `json:"id"`
	ReferenceID string          `json:"reference_id"`
	CreatedAt   string          `json:"created_at"`
	Charges     []PagBankCharge `json:"charges"`
}

type PagBankCharge struct {
	ID              string                 `json:"id"`
	ReferenceID     string                 `json:"reference_id"`
	Status          string                 `json:"status"` // "PAID", "AUTHORIZED", "DECLINED", "CANCELED"
	CreatedAt       string                 `json:"created_at"`
	PaidAt          string                 `json:"paid_at"`
	Amount          PagBankAmount          `json:"amount"`
	PaymentMethod   PagBankPaymentMethod   `json:"payment_method"`
	PaymentResponse PagBankPaymentResponse `json:"payment_response"`
	Fees            *PagBankFees           `json:"fees,omitempty"`
}

type PagBankAmount struct {
	Value    int64          `json:"value"` // In cents
	Currency string         `json:"currency"`
	Summary  PagBankSummary `json:"summary"`
}

type PagBankSummary struct {
	Total    int64 `json:"total"`
	Paid     int64 `json:"paid"`
	Refunded int64 `json:"refunded"`
}

type PagBankPaymentMethod struct {
	Type         string       `json:"type"` // "CREDIT_CARD", "DEBIT_CARD", "PIX", "BOLETO"
	Installments int          `json:"installments"`
	Card         *PagBankCard `json:"card,omitempty"`
}

type PagBankCard struct {
	Brand      string `json:"brand"`
	LastDigits string `json:"last_digits"`
}

type PagBankPaymentResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Reference string `json:"reference"` // Acquirer NSU
}

type PagBankFees struct {
	Seller int64 `json:"seller"` // MDR fee in cents
}

type PagBankPayoutItem struct {
	ID            string   `json:"id"`
	ScheduledDate string   `json:"scheduled_date"` // YYYY-MM-DD
	Amount        int64    `json:"amount"`         // Gross in cents
	Fee           int64    `json:"fee"`            // Fee in cents
	NetAmount     int64    `json:"net_amount"`     // Net in cents
	Status        string   `json:"status"`
	ChargeIDs     []string `json:"charge_ids"`
}

// NormalizeSale transforms a PagBank charge into the domain Sale model.
func NormalizeSale(charge *PagBankCharge, connectionID, receivablesAccountID string) (*domain.Sale, error) {
	if charge == nil {
		return nil, fmt.Errorf("pagbank: nil charge payload")
	}

	dateStr := charge.PaidAt
	if dateStr == "" {
		dateStr = charge.CreatedAt
	}

	saleTime, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		saleTime, err = time.Parse("2006-01-02T15:04:05Z07:00", dateStr)
		if err != nil {
			saleTime = time.Now().UTC()
		}
	}

	hundred := decimal.NewFromInt(100)
	gross := decimal.NewFromInt(charge.Amount.Value).DivRound(hundred, 2)

	var fee decimal.Decimal
	if charge.Fees != nil && charge.Fees.Seller > 0 {
		fee = decimal.NewFromInt(charge.Fees.Seller).DivRound(hundred, 2)
	} else {
		// Default estimation if not provided in payload (e.g. 2.99%)
		fee = gross.Mul(decimal.RequireFromString("0.0299")).Round(2)
	}

	net := gross.Sub(fee)

	installmentsCount := charge.PaymentMethod.Installments
	if installmentsCount <= 0 {
		installmentsCount = 1
	}

	cardBrand := ""
	if charge.PaymentMethod.Card != nil {
		cardBrand = charge.PaymentMethod.Card.Brand
	}

	paymentMethod := charge.PaymentMethod.Type
	if paymentMethod == "" {
		paymentMethod = "CREDIT_CARD"
	}

	nsu := charge.PaymentResponse.Reference
	if nsu == "" {
		nsu = charge.ReferenceID
		if nsu == "" {
			nsu = charge.ID
		}
	}

	sale := &domain.Sale{
		ConnectionID:         connectionID,
		ReceivablesAccountID: receivablesAccountID,
		Acquirer:             domain.AcquirerPagBank,
		NSU:                  nsu,
		TID:                  charge.ID,
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

		settlementDays := 30 * instNum
		if paymentMethod == "DEBIT_CARD" || paymentMethod == "PIX" {
			settlementDays = 1
		}
		expectedSettlement := saleTime.AddDate(0, 0, settlementDays)

		sale.Installments[i] = domain.Installment{
			ConnectionID:           connectionID,
			ReceivablesAccountID:   receivablesAccountID,
			Acquirer:               domain.AcquirerPagBank,
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
