package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type InstallmentStatus string

const (
	InstallmentStatusPendingSettlement InstallmentStatus = "PENDING_SETTLEMENT"
	InstallmentStatusReconciled        InstallmentStatus = "RECONCILED"
	InstallmentStatusCancelled         InstallmentStatus = "CANCELLED"
	InstallmentStatusChargeback        InstallmentStatus = "CHARGEBACK"
	InstallmentStatusOverdue           InstallmentStatus = "OVERDUE"
)

// Sale represents a confirmed card sale captured by an acquirer.
type Sale struct {
	ID                   string          `json:"id"`
	ConnectionID         string          `json:"connectionId"`
	ReceivablesAccountID string          `json:"receivablesAccountId"`
	Acquirer             AcquirerType    `json:"acquirer"`
	NSU                  string          `json:"nsu"`
	TID                  string          `json:"tid,omitempty"`
	AuthorizationCode    string          `json:"authorizationCode,omitempty"`
	SaleDate             time.Time       `json:"saleDate"`
	GrossAmount          decimal.Decimal `json:"grossAmount"`
	TotalMDRFee          decimal.Decimal `json:"totalMdrFee"`
	TotalNetAmount       decimal.Decimal `json:"totalNetAmount"`
	InstallmentsCount    int             `json:"installmentsCount"`
	PaymentMethod        string          `json:"paymentMethod"` // "CREDIT_CARD", "DEBIT_CARD", "PIX"
	CardBrand            string          `json:"cardBrand,omitempty"`
	Installments         []Installment   `json:"installments"`
	CreatedAt            time.Time       `json:"createdAt"`
}

// Installment represents a single installment line belonging to a Sale.
type Installment struct {
	ID                     string            `json:"id"`
	SaleID                 string            `json:"saleId"`
	ConnectionID           string            `json:"connectionId"`
	ReceivablesAccountID   string            `json:"receivablesAccountId"`
	Acquirer               AcquirerType      `json:"acquirer"`
	NSU                    string            `json:"nsu"`
	InstallmentNumber      int               `json:"installmentNumber"`
	GrossAmount            decimal.Decimal   `json:"grossAmount"`
	MDRFee                 decimal.Decimal   `json:"mdrFee"`
	NetAmount              decimal.Decimal   `json:"netAmount"`
	ExpectedSettlementDate time.Time         `json:"expectedSettlementDate"`
	ActualSettlementDate   *time.Time        `json:"actualSettlementDate,omitempty"`
	Status                 InstallmentStatus `json:"status"`
	MidazReceivableTxID    *string           `json:"midazReceivableTxId,omitempty"`
	MidazSettlementTxID    *string           `json:"midazSettlementTxId,omitempty"`
	MatcherRecordID        *string           `json:"matcherRecordId,omitempty"`
	MatcherRunID           *string           `json:"matcherRunId,omitempty"`
	CreatedAt              time.Time         `json:"createdAt"`
	UpdatedAt              time.Time         `json:"updatedAt"`
}

// DeterministicReceivableCode generates the idempotent Midaz transaction code for D+0 agenda booking.
func (i *Installment) DeterministicReceivableCode() string {
	return "REC-" + string(i.Acquirer) + "-" + i.NSU + "-" + decimal.NewFromInt(int64(i.InstallmentNumber)).String()
}

// DeterministicSettlementCode generates the idempotent Midaz transaction code for D+X settlement extinguishment.
func (i *Installment) DeterministicSettlementCode() string {
	return "SETTLE-" + string(i.Acquirer) + "-" + i.NSU + "-" + decimal.NewFromInt(int64(i.InstallmentNumber)).String()
}

// ParseDeterministicReceivableCode extracts acquirer, nsu, and installment number from a REC- transaction code.
func ParseDeterministicReceivableCode(code string) (AcquirerType, string, int, error) {
	// Expected format: REC-<ACQUIRER>-<NSU>-<INSTALLMENT_NUMBER>
	if !strings.HasPrefix(code, "REC-") {
		return "", "", 0, fmt.Errorf("invalid receivable code prefix: %s", code)
	}

	parts := strings.Split(code, "-")
	if len(parts) < 4 {
		return "", "", 0, fmt.Errorf("invalid receivable code structure: %s", code)
	}

	acquirer := AcquirerType(parts[1])
	installmentNumStr := parts[len(parts)-1]
	installmentNum, err := strconv.Atoi(installmentNumStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid installment number in code %s: %w", code, err)
	}

	// NSU can contain dashes, so join middle parts
	nsu := strings.Join(parts[2:len(parts)-1], "-")

	return acquirer, nsu, installmentNum, nil
}
