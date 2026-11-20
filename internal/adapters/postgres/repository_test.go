package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateReceivablesAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	account := &domain.ReceivablesAccount{
		ID:                   "acc-001",
		OrganizationID:       "org-001",
		LedgerID:             "led-001",
		MidazAccountID:       "mid-acc-001",
		AccountAlias:         "@merchant_alpha/receivables",
		CheckingAccountAlias: "@merchant_alpha/checking",
		Name:                 "Receivables - Merchant Alpha",
		Status:               "ACTIVE",
	}

	mock.ExpectExec("INSERT INTO receivables_accounts").
		WithArgs(
			account.ID, account.OrganizationID, account.LedgerID, account.MidazAccountID, account.ParentAccountID,
			nil, account.AccountAlias, nil, account.CheckingAccountAlias,
			nil, account.Name, account.Status, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateReceivablesAccount(context.Background(), account)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindReceivablesAccountByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	now := time.Now().UTC()
	rows := sqlmock.NewRows([]string{
		"id", "organization_id", "ledger_id", "midaz_account_id", "parent_account_id",
		"mother_account_alias", "account_alias", "fee_account_alias", "checking_account_alias",
		"matcher_context_id", "name", "status", "created_at", "updated_at",
	}).AddRow(
		"acc-001", "org-001", "led-001", "mid-acc-001", nil,
		nil, "@merchant_alpha/receivables", nil, "@merchant_alpha/checking",
		nil, "Receivables - Alpha", "ACTIVE", now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM receivables_accounts WHERE id = \\$1").
		WithArgs("acc-001").
		WillReturnRows(rows)

	acc, err := repo.FindReceivablesAccountByID(context.Background(), "acc-001")
	assert.NoError(t, err)
	assert.NotNil(t, acc)
	assert.Equal(t, "acc-001", acc.ID)
	assert.Equal(t, "@merchant_alpha/receivables", acc.AccountAlias)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAcquirerConnection(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	conn := &domain.AcquirerConnection{
		ID:                   "conn-001",
		ReceivablesAccountID: "acc-001",
		AcquirerType:         domain.AcquirerSumUp,
		DisplayName:          "SumUp Main POS",
		CredentialsEncrypted: "encrypted_base64_payload",
		Status:               domain.ConnectionStatusActive,
		AutoSyncEnabled:      true,
		SyncIntervalMinutes:  60,
	}

	mock.ExpectExec("INSERT INTO acquirer_connections").
		WithArgs(
			conn.ID, conn.ReceivablesAccountID, conn.AcquirerType, conn.DisplayName,
			conn.CredentialsEncrypted, conn.Status, conn.AutoSyncEnabled, conn.SyncIntervalMinutes,
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateAcquirerConnection(context.Background(), conn)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindAcquirerConnectionByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	now := time.Now().UTC()
	rows := sqlmock.NewRows([]string{
		"id", "receivables_account_id", "acquirer_type", "display_name",
		"credentials_encrypted", "status", "auto_sync_enabled", "sync_interval_minutes",
		"last_sync_at", "last_sync_status", "last_sync_error",
		"created_at", "updated_at",
	}).AddRow(
		"conn-001", "acc-001", "SUMUP", "SumUp Main POS",
		"encrypted_base64_payload", "ACTIVE", true, 60,
		nil, nil, nil,
		now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM acquirer_connections WHERE id = \\$1").
		WithArgs("conn-001").
		WillReturnRows(rows)

	conn, err := repo.FindAcquirerConnectionByID(context.Background(), "conn-001")
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	assert.Equal(t, "conn-001", conn.ID)
	assert.Equal(t, domain.AcquirerSumUp, conn.AcquirerType)
	assert.Equal(t, "encrypted_base64_payload", conn.CredentialsEncrypted)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveSaleWithInstallments(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	gross := decimal.NewFromFloat(150.00)
	fee := decimal.NewFromFloat(4.50)
	net := decimal.NewFromFloat(145.50)

	sale := &domain.Sale{
		ID:                   "sale-001",
		ConnectionID:         "conn-001",
		ReceivablesAccountID: "acc-001",
		Acquirer:             domain.AcquirerSumUp,
		NSU:                  "NSU-987654",
		TID:                  "TID-1234",
		SaleDate:             time.Now().UTC(),
		GrossAmount:          gross,
		TotalMDRFee:          fee,
		TotalNetAmount:       net,
		InstallmentsCount:    1,
		PaymentMethod:        "CREDIT_CARD",
		Installments: []domain.Installment{
			{
				ID:                     "inst-001",
				InstallmentNumber:      1,
				GrossAmount:            gross,
				MDRFee:                 fee,
				NetAmount:              net,
				ExpectedSettlementDate: time.Now().AddDate(0, 1, 0).UTC(),
				Status:                 domain.InstallmentStatusPendingSettlement,
			},
		},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO acquirer_sales").
		WithArgs(
			sale.ID, sale.ConnectionID, sale.ReceivablesAccountID, sale.Acquirer,
			sale.NSU, sale.TID, sale.AuthorizationCode, sale.SaleDate,
			sale.GrossAmount, sale.TotalMDRFee, sale.TotalNetAmount,
			sale.InstallmentsCount, sale.PaymentMethod, sale.CardBrand, sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("sale-001"))

	mock.ExpectExec("INSERT INTO receivables_installments").
		WithArgs(
			sale.Installments[0].ID, "sale-001", sale.ConnectionID, sale.ReceivablesAccountID,
			sale.Acquirer, sale.NSU, 1,
			gross, fee, net,
			sale.Installments[0].ExpectedSettlementDate, domain.InstallmentStatusPendingSettlement,
			nil, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.SaveSaleWithInstallments(context.Background(), sale)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMarkInstallmentReconciled(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	settlementDate := time.Now().UTC()
	mock.ExpectExec("UPDATE receivables_installments SET status = 'RECONCILED'").
		WithArgs("inst-001", "midaz-settle-tx-001", "matcher-run-001", settlementDate).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.MarkInstallmentReconciled(context.Background(), "inst-001", "midaz-settle-tx-001", "matcher-run-001", settlementDate)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetReceivablesCalendar(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	now := time.Now().UTC()
	expectedDate := now.AddDate(0, 1, 0)
	status := domain.InstallmentStatusPendingSettlement

	rows := sqlmock.NewRows([]string{
		"id", "sale_id", "connection_id", "receivables_account_id",
		"acquirer", "nsu", "installment_number",
		"gross_amount", "mdr_fee", "net_amount",
		"expected_settlement_date", "actual_settlement_date", "status",
		"midaz_receivable_tx_id", "midaz_settlement_tx_id",
		"matcher_record_id", "matcher_run_id",
		"created_at", "updated_at",
	}).AddRow(
		"inst-001", "sale-001", "conn-001", "acc-001",
		"SUMUP", "NSU-987654", 1,
		decimal.NewFromFloat(100.00), decimal.NewFromFloat(3.00), decimal.NewFromFloat(97.00),
		expectedDate, nil, status,
		nil, nil, nil, nil,
		now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM receivables_installments WHERE receivables_account_id = \\$1").
		WithArgs("acc-001", 100, 0).
		WillReturnRows(rows)

	filter := domain.CalendarFilter{
		ReceivablesAccountID: "acc-001",
	}

	installments, err := repo.GetReceivablesCalendar(context.Background(), filter)
	assert.NoError(t, err)
	assert.Len(t, installments, 1)
	assert.Equal(t, "inst-001", installments[0].ID)
	assert.Equal(t, "SUMUP", string(installments[0].Acquirer))
	assert.Equal(t, "97", installments[0].NetAmount.String())
	assert.NoError(t, mock.ExpectationsWereMet())
}
