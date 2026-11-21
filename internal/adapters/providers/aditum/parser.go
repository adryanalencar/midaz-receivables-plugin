package aditum

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
)

// AditumChargesListResponse represents the JSON response from Aditum Portal GET /v1/Charges
type AditumChargesListResponse struct {
	Charges     []AditumChargeItem `json:"charges"`
	CurrentPage int                `json:"currentPage"`
	PageSize    int                `json:"pageSize"`
	TotalItems  int                `json:"totalItems"`
	TotalPages  int                `json:"totalPages"`
}

// AditumChargeItem represents an individual charge in Aditum
type AditumChargeItem struct {
	ID                  string                  `json:"id"`
	ChargeID            string                  `json:"chargeId"`
	NSU                 string                  `json:"nsu"`
	ChargeStatus        string                  `json:"chargeStatus"`
	MerchantID          string                  `json:"merchantId"`
	MerchantFantasyName string                  `json:"merchantFantasyName"`
	CNPJ                string                  `json:"cnpj"`
	DataCobranca        string                  `json:"dataCobranca"`
	CreationDate        string                  `json:"creationDate"`
	CreationDateTime    string                  `json:"creationDateTime"`
	Transactions        []AditumTransactionItem `json:"transactions"`
}

// AditumTransactionItem represents a transaction attempt or capture within a charge
type AditumTransactionItem struct {
	TransactionID             string             `json:"transactionId"`
	AuthorizationCode         string             `json:"authorizationCode"`
	AuthorizationResponseCode string             `json:"authorizationResponseCode"`
	Amount                    int64              `json:"amount"` // in cents
	OriginalAmount            int64              `json:"originalAmount"`
	FeeAmount                 int64              `json:"feeAmount"` // in cents (optional)
	NetAmount                 int64              `json:"netAmount"` // in cents (optional)
	IsCapture                 bool               `json:"isCapture"`
	IsCanceled                bool               `json:"isCanceled"`
	IsApproved                bool               `json:"isApproved"`
	TransactionStatus         string             `json:"transactionStatus"`
	CancelStatus              string             `json:"cancelStatus"`
	PaymentType               string             `json:"paymentType"` // "credit", "creditcard", "debit", "debitcard", "pix"
	InstallmentNumber         int                `json:"installmentNumber"`
	AcquirerID                string             `json:"acquirerId"`
	AcquirerName              string             `json:"acquirerName"`
	CreationDateTime          string             `json:"creationDateTime"`
	CaptureDateTime           string             `json:"captureDateTime"`
	Card                      *AditumCardItem    `json:"card"`
	Terminal                  *AditumTerminalItem `json:"terminal"`
}

// AditumCardItem contains card token/details
type AditumCardItem struct {
	ID             string `json:"id"`
	Brand          string `json:"brand"`
	BrandName      string `json:"brandName"`
	CardNumber     string `json:"cardNumber"`
	CardholderName string `json:"cardholderName"`
}

// AditumTerminalItem contains terminal POS details
type AditumTerminalItem struct {
	SerialNumber string `json:"serialNumber"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Type         string `json:"type"`
}

// AditumWebhookPayload represents the structure sent by Aditum Webhooks (supports PascalCase and camelCase)
type AditumWebhookPayload struct {
	ChargeID         string                  `json:"ChargeId"`
	ChargeIDAlt      string                  `json:"chargeId"`
	ID               string                  `json:"id"`
	NSU              string                  `json:"Nsu"`
	NSUAlt           string                  `json:"nsu"`
	ChargeStatus     string                  `json:"ChargeStatus"`
	ChargeStatusAlt  string                  `json:"chargeStatus"`
	MerchantID       string                  `json:"MerchantId"`
	MerchantIDAlt    string                  `json:"merchantId"`
	Amount           int64                   `json:"Amount"`
	AmountAlt        int64                   `json:"amount"`
	CreationDate     string                  `json:"CreationDate"`
	CreationDateTime string                  `json:"CreationDateTime"`
	Transactions     []AditumTransactionItem `json:"Transactions"`
	TransactionsAlt  []AditumTransactionItem `json:"transactions"`
}

func parseAditumTimestamp(ts string) time.Time {
	if ts == "" {
		return time.Now().UTC()
	}

	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range formats {
		if t, err := time.Parse(layout, ts); err == nil {
			return t.UTC()
		}
	}

	return time.Now().UTC()
}

// NormalizeSale converts an AditumChargeItem into the domain Sale entity
func NormalizeSale(item *AditumChargeItem, connectionID, receivablesAccountID string) (*domain.Sale, error) {
	if item == nil {
		return nil, fmt.Errorf("aditum: nil charge item")
	}

	chargeID := item.ID
	if chargeID == "" {
		chargeID = item.ChargeID
	}

	nsu := item.NSU

	// Pick first captured/approved transaction or first transaction
	var tx *AditumTransactionItem
	for i := range item.Transactions {
		t := &item.Transactions[i]
		if t.IsCapture || strings.EqualFold(t.TransactionStatus, "Authorized") || strings.EqualFold(t.TransactionStatus, "Paid") {
			tx = t
			break
		}
	}
	if tx == nil && len(item.Transactions) > 0 {
		tx = &item.Transactions[0]
	}

	var gross decimal.Decimal
	var fee decimal.Decimal
	var net decimal.Decimal
	var tid string
	var authCode string
	var paymentMethod = "CREDIT_CARD"
	var cardBrand = ""
	var installmentsCount = 1

	if tx != nil {
		if nsu == "" {
			nsu = tx.TransactionID
		}
		tid = tx.TransactionID
		authCode = tx.AuthorizationCode

		if tx.Amount > 0 {
			gross = decimal.NewFromInt(tx.Amount).Div(decimal.NewFromInt(100)).Round(2)
		} else if tx.OriginalAmount > 0 {
			gross = decimal.NewFromInt(tx.OriginalAmount).Div(decimal.NewFromInt(100)).Round(2)
		}

		if tx.FeeAmount > 0 {
			fee = decimal.NewFromInt(tx.FeeAmount).Div(decimal.NewFromInt(100)).Round(2)
		}
		if tx.NetAmount > 0 {
			net = decimal.NewFromInt(tx.NetAmount).Div(decimal.NewFromInt(100)).Round(2)
		} else {
			net = gross.Sub(fee)
		}

		if tx.InstallmentNumber > 1 {
			installmentsCount = tx.InstallmentNumber
		}

		pt := strings.ToLower(tx.PaymentType)
		if strings.Contains(pt, "debit") {
			paymentMethod = "DEBIT_CARD"
		} else if strings.Contains(pt, "pix") {
			paymentMethod = "PIX"
		} else {
			paymentMethod = "CREDIT_CARD"
		}

		if tx.Card != nil {
			cardBrand = tx.Card.BrandName
			if cardBrand == "" {
				cardBrand = tx.Card.Brand
			}
		}
	}

	if gross.IsZero() {
		return nil, fmt.Errorf("aditum: charge has zero or invalid gross amount")
	}

	// Invariant: Gross - Fee == Net
	if !gross.Sub(fee).Equal(net) {
		net = gross.Sub(fee)
	}

	dateStr := item.DataCobranca
	if dateStr == "" {
		dateStr = item.CreationDateTime
	}
	if dateStr == "" {
		dateStr = item.CreationDate
	}
	if dateStr == "" && tx != nil {
		dateStr = tx.CaptureDateTime
		if dateStr == "" {
			dateStr = tx.CreationDateTime
		}
	}

	saleDate := parseAditumTimestamp(dateStr)

	sale := &domain.Sale{
		ConnectionID:         connectionID,
		ReceivablesAccountID: receivablesAccountID,
		Acquirer:             domain.AcquirerAditum,
		NSU:                  nsu,
		TID:                  tid,
		AuthorizationCode:    authCode,
		SaleDate:             saleDate,
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
		expectedSettlement := saleDate.AddDate(0, 0, settlementDays)

		sale.Installments[i] = domain.Installment{
			ConnectionID:           connectionID,
			ReceivablesAccountID:   receivablesAccountID,
			Acquirer:               domain.AcquirerAditum,
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

// NormalizeWebhook parses raw JSON webhook bytes into a domain.Sale
func NormalizeWebhook(body []byte, connectionID, receivablesAccountID string) (*domain.Sale, error) {
	var payload AditumWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("aditum: invalid webhook json: %w", err)
	}

	chargeID := payload.ChargeID
	if chargeID == "" {
		chargeID = payload.ChargeIDAlt
	}
	if chargeID == "" {
		chargeID = payload.ID
	}

	nsu := payload.NSU
	if nsu == "" {
		nsu = payload.NSUAlt
	}

	txs := payload.Transactions
	if len(txs) == 0 {
		txs = payload.TransactionsAlt
	}

	item := &AditumChargeItem{
		ID:               chargeID,
		ChargeID:         chargeID,
		NSU:              nsu,
		ChargeStatus:     payload.ChargeStatus,
		MerchantID:       payload.MerchantID,
		CreationDate:     payload.CreationDate,
		CreationDateTime: payload.CreationDateTime,
		Transactions:     txs,
	}

	if item.ChargeStatus == "" {
		item.ChargeStatus = payload.ChargeStatusAlt
	}
	if item.MerchantID == "" {
		item.MerchantID = payload.MerchantIDAlt
	}

	// If no transactions in webhook array but top-level amount exists, synthesize transaction
	if len(item.Transactions) == 0 && (payload.Amount > 0 || payload.AmountAlt > 0) {
		amt := payload.Amount
		if amt == 0 {
			amt = payload.AmountAlt
		}
		item.Transactions = []AditumTransactionItem{
			{
				TransactionID:     chargeID,
				Amount:            amt,
				IsCapture:         true,
				IsApproved:        true,
				PaymentType:       "credit",
				InstallmentNumber: 1,
			},
		}
	}

	return NormalizeSale(item, connectionID, receivablesAccountID)
}

// ParseCentsToDecimal converts integer cents (e.g. 15000) to decimal 150.00
func ParseCentsToDecimal(cents int64) decimal.Decimal {
	return decimal.NewFromInt(cents).Div(decimal.NewFromInt(100)).Round(2)
}

// ParseStringCentsToDecimal parses a string representing cents (e.g. "00000015000") to decimal 150.00
func ParseStringCentsToDecimal(centsStr string) decimal.Decimal {
	clean := strings.TrimSpace(centsStr)
	if clean == "" {
		return decimal.Zero
	}
	val, err := strconv.ParseInt(clean, 10, 64)
	if err != nil {
		return decimal.Zero
	}
	return ParseCentsToDecimal(val)
}
