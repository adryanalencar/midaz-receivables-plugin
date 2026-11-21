package aditum

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
)

// EdiRecordCV represents a parsed Sales Voucher (Comprovante de Venda / Recebível) record from EDI Layout 3.2
type EdiRecordCV struct {
	MerchantDocument       string          `json:"merchantDocument"`
	TransactionNSU         string          `json:"transactionNsu"`
	TransactionDate        time.Time       `json:"transactionDate"`
	SettlementDate         time.Time       `json:"settlementDate"`
	GrossAmount            decimal.Decimal `json:"grossAmount"`
	FeeAmount              decimal.Decimal `json:"feeAmount"`
	NetAmount              decimal.Decimal `json:"netAmount"`
	MaskedCardNumber       string          `json:"maskedCardNumber"`
	InstallmentNumber      int             `json:"installmentNumber"`
	InstallmentCount       int             `json:"installmentCount"`
	InstallmentNSU         string          `json:"installmentNsu"`
	InstallmentGrossAmount decimal.Decimal `json:"installmentGrossAmount"`
	InstallmentFeeAmount   decimal.Decimal `json:"installmentFeeAmount"`
	InstallmentNetAmount   decimal.Decimal `json:"installmentNetAmount"`
	BankCode               string          `json:"bankCode"`
	BranchCode             string          `json:"branchCode"`
	AccountNumber          string          `json:"accountNumber"`
	AuthorizationCode      string          `json:"authorizationCode"`
	BrandCode              string          `json:"brandCode"`
}

// EdiFile32 represents a parsed EDI 3.2 settlement file
type EdiFile32 struct {
	GenerationDate time.Time
	MovementID     string
	SalesRecords   []EdiRecordCV
}

func parseFixedField(line string, start, end int) string {
	// Fixed-width 1-based indexing [start, end]
	if len(line) < end {
		if len(line) >= start {
			return strings.TrimSpace(line[start-1:])
		}
		return ""
	}
	return strings.TrimSpace(line[start-1 : end])
}

func parseFixedDate(val string) time.Time {
	val = strings.TrimSpace(val)
	if len(val) == 8 {
		if t, err := time.Parse("20060102", val); err == nil {
			return t
		}
	}
	return time.Time{}
}

func parseFixedAmount(val string) decimal.Decimal {
	val = strings.TrimSpace(val)
	if val == "" {
		return decimal.Zero
	}
	cents, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return decimal.Zero
	}
	return decimal.NewFromInt(cents).Div(decimal.NewFromInt(100)).Round(2)
}

func parseFixedInt(val string) int {
	val = strings.TrimSpace(val)
	if val == "" {
		return 0
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return n
}

// ParseEdiLayout32 parses an EDI text stream into an EdiFile32 structure
func ParseEdiLayout32(r io.Reader) (*EdiFile32, error) {
	scanner := bufio.NewScanner(r)
	edi := &EdiFile32{
		SalesRecords: make([]EdiRecordCV, 0),
	}

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 {
			continue
		}

		recordType := line[0:2]
		switch recordType {
		case "A0": // Header de Arquivo
			dateStr := parseFixedField(line, 9, 16)
			edi.GenerationDate = parseFixedDate(dateStr)
			edi.MovementID = parseFixedField(line, 23, 28)

		case "CV": // Comprovante de Venda / Liquidação
			cv := EdiRecordCV{
				MerchantDocument:       parseFixedField(line, 3, 17),
				TransactionNSU:         parseFixedField(line, 18, 29),
				TransactionDate:        parseFixedDate(parseFixedField(line, 30, 37)),
				SettlementDate:         parseFixedDate(parseFixedField(line, 45, 52)),
				GrossAmount:            parseFixedAmount(parseFixedField(line, 55, 65)),
				FeeAmount:              parseFixedAmount(parseFixedField(line, 66, 76)),
				NetAmount:              parseFixedAmount(parseFixedField(line, 77, 87)),
				MaskedCardNumber:       parseFixedField(line, 88, 106),
				InstallmentNumber:      parseFixedInt(parseFixedField(line, 107, 108)),
				InstallmentCount:       parseFixedInt(parseFixedField(line, 109, 110)),
				InstallmentNSU:         parseFixedField(line, 111, 122),
				InstallmentGrossAmount: parseFixedAmount(parseFixedField(line, 123, 133)),
				InstallmentFeeAmount:   parseFixedAmount(parseFixedField(line, 134, 144)),
				InstallmentNetAmount:   parseFixedAmount(parseFixedField(line, 145, 155)),
				BankCode:               parseFixedField(line, 156, 158),
				BranchCode:             parseFixedField(line, 159, 164),
				AccountNumber:          parseFixedField(line, 165, 175),
				AuthorizationCode:      parseFixedField(line, 176, 187),
				BrandCode:              parseFixedField(line, 188, 190),
			}
			edi.SalesRecords = append(edi.SalesRecords, cv)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("aditum: error reading edi stream: %w", err)
	}

	return edi, nil
}

// ConvertEdiToPayoutSchedules transforms EDI 3.2 records grouped by settlement date into domain.PayoutSchedule entries
func ConvertEdiToPayoutSchedules(edi *EdiFile32) []domain.PayoutSchedule {
	if edi == nil || len(edi.SalesRecords) == 0 {
		return nil
	}

	type dateGroup struct {
		gross decimal.Decimal
		fee   decimal.Decimal
		net   decimal.Decimal
		nsus  []string
		date  time.Time
	}

	groups := make(map[string]*dateGroup)
	for _, rec := range edi.SalesRecords {
		if rec.SettlementDate.IsZero() {
			continue
		}
		dateKey := rec.SettlementDate.Format("2006-01-02")
		g, exists := groups[dateKey]
		if !exists {
			g = &dateGroup{
				date: rec.SettlementDate,
				nsus: make([]string, 0),
			}
			groups[dateKey] = g
		}

		instGross := rec.InstallmentGrossAmount
		if instGross.IsZero() {
			instGross = rec.GrossAmount
		}
		instFee := rec.InstallmentFeeAmount
		if instFee.IsZero() {
			instFee = rec.FeeAmount
		}
		instNet := rec.InstallmentNetAmount
		if instNet.IsZero() {
			instNet = rec.NetAmount
		}

		g.gross = g.gross.Add(instGross)
		g.fee = g.fee.Add(instFee)
		g.net = g.net.Add(instNet)
		if rec.TransactionNSU != "" {
			g.nsus = append(g.nsus, rec.TransactionNSU)
		}
	}

	payouts := make([]domain.PayoutSchedule, 0, len(groups))
	for dateKey, g := range groups {
		payouts = append(payouts, domain.PayoutSchedule{
			Acquirer:               domain.AcquirerAditum,
			PayoutID:               fmt.Sprintf("EDI-%s-%s", edi.MovementID, dateKey),
			ExpectedSettlementDate: g.date,
			GrossAmount:            g.gross,
			FeeAmount:              g.fee,
			NetAmount:              g.net,
			AssociatedNSUs:         g.nsus,
		})
	}

	return payouts
}
