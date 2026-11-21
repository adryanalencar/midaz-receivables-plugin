package aditum

import (
	"strings"
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEdiLayout32_AndConversion(t *testing.T) {
	// Sample A0 header line (length >= 78)
	// A0 000000 20260320 120000 000001 ADITUM ADMINISTRADORA         0001 000000001 1 000001
	a0Line := "A0" + "000000" + "20260320" + "120000" + "000001" + strings.Repeat(" ", 30) + "0001" + "000000001" + "1" + "000001"

	// Sample CV line (length >= 506)
	// Pos 1-2: CV
	// Pos 3-17: Documento (15 chars) -> 12345678000195
	// Pos 18-29: NSU (12 chars) -> 000012345678
	// Pos 30-37: Data Transação (8 chars) -> 20260320
	// Pos 38-43: Hora Transação (6 chars) -> 143000
	// Pos 44: Tipo lancamento (1 char) -> 1
	// Pos 45-52: Data Liquidacao (8 chars) -> 20260420
	// Pos 53: Tipo produto (1 char) -> 1
	// Pos 54: Meio captura (1 char) -> 1
	// Pos 55-65: Valor Bruto (11 chars) -> 00000010000 (R$ 100.00)
	// Pos 66-76: Valor Taxa (11 chars) -> 00000000300 (R$ 3.00)
	// Pos 77-87: Valor Liquido (11 chars) -> 00000009700 (R$ 97.00)
	// Pos 88-106: Cartao Mascarado (19 chars) -> 411111******1111
	// Pos 107-108: Num Parcela (2 chars) -> 01
	// Pos 109-110: Total Parcelas (2 chars) -> 01
	// Pos 111-122: NSU Parcela (12 chars) -> 000012345678
	// Pos 123-133: Valor Bruto Parcela (11 chars) -> 00000010000
	// Pos 134-144: Valor Taxa Parcela (11 chars) -> 00000000300
	// Pos 145-155: Valor Liquido Parcela (11 chars) -> 00000009700
	// Pos 156-158: Banco (3 chars) -> 001
	// Pos 159-164: Agencia (6 chars) -> 001234
	// Pos 165-175: Conta (11 chars) -> 00012345678
	// Pos 176-187: Cod Autorizacao (12 chars) -> AUTH123456
	// Pos 188-190: Bandeira (3 chars) -> 001
	cvBuilder := strings.Builder{}
	cvBuilder.WriteString("CV")
	cvBuilder.WriteString(fmtFixed("12345678000195", 15))
	cvBuilder.WriteString(fmtFixed("000012345678", 12))
	cvBuilder.WriteString("20260320")
	cvBuilder.WriteString("143000")
	cvBuilder.WriteString("1")
	cvBuilder.WriteString("20260420")
	cvBuilder.WriteString("1")
	cvBuilder.WriteString("1")
	cvBuilder.WriteString("00000010000") // Gross: 100.00
	cvBuilder.WriteString("00000000300") // Fee: 3.00
	cvBuilder.WriteString("00000009700") // Net: 97.00
	cvBuilder.WriteString(fmtFixed("411111******1111", 19))
	cvBuilder.WriteString("01")
	cvBuilder.WriteString("01")
	cvBuilder.WriteString(fmtFixed("000012345678", 12))
	cvBuilder.WriteString("00000010000")
	cvBuilder.WriteString("00000000300")
	cvBuilder.WriteString("00000009700")
	cvBuilder.WriteString("001")
	cvBuilder.WriteString("001234")
	cvBuilder.WriteString("00012345678")
	cvBuilder.WriteString(fmtFixed("AUTH123456", 12))
	cvBuilder.WriteString("001")
	cvBuilder.WriteString(strings.Repeat(" ", 320)) // Pad to >= 506 chars

	ediContent := a0Line + "\n" + cvBuilder.String() + "\n"

	edi, err := ParseEdiLayout32(strings.NewReader(ediContent))
	require.NoError(t, err)
	require.NotNil(t, edi)

	assert.Equal(t, "000001", edi.MovementID)
	assert.Equal(t, time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC), edi.GenerationDate)
	require.Len(t, edi.SalesRecords, 1)

	rec := edi.SalesRecords[0]
	assert.Equal(t, "000012345678", rec.TransactionNSU)
	assert.Equal(t, time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC), rec.SettlementDate)
	assert.True(t, decimal.NewFromInt(100).Equal(rec.GrossAmount))
	assert.True(t, decimal.NewFromInt(3).Equal(rec.FeeAmount))
	assert.True(t, decimal.NewFromInt(97).Equal(rec.NetAmount))
	assert.Equal(t, 1, rec.InstallmentNumber)
	assert.Equal(t, "001", rec.BankCode)
	assert.Equal(t, "00012345678", rec.AccountNumber)

	payouts := ConvertEdiToPayoutSchedules(edi)
	require.Len(t, payouts, 1)
	p := payouts[0]
	assert.Equal(t, domain.AcquirerAditum, p.Acquirer)
	assert.Equal(t, time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC), p.ExpectedSettlementDate)
	assert.True(t, decimal.NewFromInt(100).Equal(p.GrossAmount))
	assert.True(t, decimal.NewFromInt(3).Equal(p.FeeAmount))
	assert.True(t, decimal.NewFromInt(97).Equal(p.NetAmount))
	assert.Contains(t, p.AssociatedNSUs, "000012345678")
}

func fmtFixed(val string, length int) string {
	if len(val) >= length {
		return val[:length]
	}
	return val + strings.Repeat(" ", length-len(val))
}
