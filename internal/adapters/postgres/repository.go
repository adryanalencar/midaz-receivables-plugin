package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrAccountNotFound     = errors.New("repository: receivables account not found")
	ErrConnectionNotFound  = errors.New("repository: acquirer connection not found")
	ErrSaleNotFound        = errors.New("repository: acquirer sale not found")
	ErrInstallmentNotFound = errors.New("repository: receivable installment not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateReceivablesAccount inserts a new receivables account record with subaccount hierarchy.
func (r *Repository) CreateReceivablesAccount(ctx context.Context, account *domain.ReceivablesAccount) error {
	if account.ID == "" {
		account.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	account.CreatedAt = now
	account.UpdatedAt = now

	var motherAliasVal, mdrAliasVal, feeAliasVal, matcherCtxVal *string
	if account.MotherAccountAlias != "" {
		motherAliasVal = &account.MotherAccountAlias
	}
	if account.MDRRevenueAccountAlias != "" {
		mdrAliasVal = &account.MDRRevenueAccountAlias
	}
	if account.FeeAccountAlias != "" {
		feeAliasVal = &account.FeeAccountAlias
	}
	if account.MatcherContextID != "" {
		matcherCtxVal = &account.MatcherContextID
	}

	query := `
		INSERT INTO receivables_accounts (
			id, organization_id, ledger_id, midaz_account_id, parent_account_id,
			mother_account_alias, account_alias, mdr_revenue_account_alias, fee_account_alias, checking_account_alias,
			matcher_context_id, name, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.ExecContext(ctx, query,
		account.ID, account.OrganizationID, account.LedgerID, account.MidazAccountID, account.ParentAccountID,
		motherAliasVal, account.AccountAlias, mdrAliasVal, feeAliasVal, account.CheckingAccountAlias,
		matcherCtxVal, account.Name, account.Status, account.CreatedAt, account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: failed to insert receivables account: %w", err)
	}

	return nil
}

// FindReceivablesAccountByID retrieves an account by primary key UUID.
func (r *Repository) FindReceivablesAccountByID(ctx context.Context, id string) (*domain.ReceivablesAccount, error) {
	query := `
		SELECT id, organization_id, ledger_id, midaz_account_id, parent_account_id,
		       mother_account_alias, account_alias, mdr_revenue_account_alias, fee_account_alias, checking_account_alias,
		       matcher_context_id, name, status, created_at, updated_at
		FROM receivables_accounts
		WHERE id = $1
	`

	var acc domain.ReceivablesAccount
	var parentID, motherAlias, mdrAlias, feeAlias, matcherCtx sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&acc.ID, &acc.OrganizationID, &acc.LedgerID, &acc.MidazAccountID, &parentID,
		&motherAlias, &acc.AccountAlias, &mdrAlias, &feeAlias, &acc.CheckingAccountAlias,
		&matcherCtx, &acc.Name, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("repository: failed to query receivables account by id: %w", err)
	}

	if parentID.Valid {
		acc.ParentAccountID = &parentID.String
	}
	if motherAlias.Valid {
		acc.MotherAccountAlias = motherAlias.String
	}
	if mdrAlias.Valid {
		acc.MDRRevenueAccountAlias = mdrAlias.String
	}
	if feeAlias.Valid {
		acc.FeeAccountAlias = feeAlias.String
	}
	if matcherCtx.Valid {
		acc.MatcherContextID = matcherCtx.String
	}

	return &acc, nil
}

// FindReceivablesAccountByAlias retrieves an account by its unique Midaz alias.
func (r *Repository) FindReceivablesAccountByAlias(ctx context.Context, alias string) (*domain.ReceivablesAccount, error) {
	query := `
		SELECT id, organization_id, ledger_id, midaz_account_id, parent_account_id,
		       mother_account_alias, account_alias, mdr_revenue_account_alias, fee_account_alias, checking_account_alias,
		       matcher_context_id, name, status, created_at, updated_at
		FROM receivables_accounts
		WHERE account_alias = $1
	`

	var acc domain.ReceivablesAccount
	var parentID, motherAlias, mdrAlias, feeAlias, matcherCtx sql.NullString

	err := r.db.QueryRowContext(ctx, query, alias).Scan(
		&acc.ID, &acc.OrganizationID, &acc.LedgerID, &acc.MidazAccountID, &parentID,
		&motherAlias, &acc.AccountAlias, &mdrAlias, &feeAlias, &acc.CheckingAccountAlias,
		&matcherCtx, &acc.Name, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("repository: failed to query receivables account by alias: %w", err)
	}

	if parentID.Valid {
		acc.ParentAccountID = &parentID.String
	}
	if motherAlias.Valid {
		acc.MotherAccountAlias = motherAlias.String
	}
	if mdrAlias.Valid {
		acc.MDRRevenueAccountAlias = mdrAlias.String
	}
	if feeAlias.Valid {
		acc.FeeAccountAlias = feeAlias.String
	}
	if matcherCtx.Valid {
		acc.MatcherContextID = matcherCtx.String
	}

	return &acc, nil
}

// ListReceivablesAccounts retrieves all accounts.
func (r *Repository) ListReceivablesAccounts(ctx context.Context) ([]domain.ReceivablesAccount, error) {
	query := `
		SELECT id, organization_id, ledger_id, midaz_account_id, parent_account_id,
		       mother_account_alias, account_alias, mdr_revenue_account_alias, fee_account_alias, checking_account_alias,
		       matcher_context_id, name, status, created_at, updated_at
		FROM receivables_accounts
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to list receivables accounts: %w", err)
	}
	defer rows.Close()

	var accounts []domain.ReceivablesAccount
	for rows.Next() {
		var acc domain.ReceivablesAccount
		var parentID, motherAlias, mdrAlias, feeAlias, matcherCtx sql.NullString

		if err := rows.Scan(
			&acc.ID, &acc.OrganizationID, &acc.LedgerID, &acc.MidazAccountID, &parentID,
			&motherAlias, &acc.AccountAlias, &mdrAlias, &feeAlias, &acc.CheckingAccountAlias,
			&matcherCtx, &acc.Name, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: failed to scan account row: %w", err)
		}

		if parentID.Valid {
			acc.ParentAccountID = &parentID.String
		}
		if motherAlias.Valid {
			acc.MotherAccountAlias = motherAlias.String
		}
		if mdrAlias.Valid {
			acc.MDRRevenueAccountAlias = mdrAlias.String
		}
		if feeAlias.Valid {
			acc.FeeAccountAlias = feeAlias.String
		}
		if matcherCtx.Valid {
			acc.MatcherContextID = matcherCtx.String
		}

		accounts = append(accounts, acc)
	}

	return accounts, rows.Err()
}

// CreateAcquirerConnection stores a new connection with encrypted credentials.
func (r *Repository) CreateAcquirerConnection(ctx context.Context, conn *domain.AcquirerConnection) error {
	if conn.ID == "" {
		conn.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	conn.CreatedAt = now
	conn.UpdatedAt = now

	query := `
		INSERT INTO acquirer_connections (
			id, receivables_account_id, acquirer_type, display_name,
			credentials_encrypted, status, auto_sync_enabled, sync_interval_minutes,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		conn.ID, conn.ReceivablesAccountID, conn.AcquirerType, conn.DisplayName,
		conn.CredentialsEncrypted, conn.Status, conn.AutoSyncEnabled, conn.SyncIntervalMinutes,
		conn.CreatedAt, conn.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: failed to insert acquirer connection: %w", err)
	}

	return nil
}

// FindAcquirerConnectionByID retrieves a connection including encrypted credentials.
func (r *Repository) FindAcquirerConnectionByID(ctx context.Context, id string) (*domain.AcquirerConnection, error) {
	query := `
		SELECT id, receivables_account_id, acquirer_type, display_name,
		       credentials_encrypted, status, auto_sync_enabled, sync_interval_minutes,
		       last_sync_at, last_sync_status, last_sync_error,
		       created_at, updated_at
		FROM acquirer_connections
		WHERE id = $1
	`

	var conn domain.AcquirerConnection
	var lastSyncAt sql.NullTime
	var lastSyncStatus, lastSyncError sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conn.ID, &conn.ReceivablesAccountID, &conn.AcquirerType, &conn.DisplayName,
		&conn.CredentialsEncrypted, &conn.Status, &conn.AutoSyncEnabled, &conn.SyncIntervalMinutes,
		&lastSyncAt, &lastSyncStatus, &lastSyncError,
		&conn.CreatedAt, &conn.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrConnectionNotFound
		}
		return nil, fmt.Errorf("repository: failed to query acquirer connection by id: %w", err)
	}

	if lastSyncAt.Valid {
		conn.LastSyncAt = &lastSyncAt.Time
	}
	if lastSyncStatus.Valid {
		conn.LastSyncStatus = lastSyncStatus.String
	}
	if lastSyncError.Valid {
		conn.LastSyncError = &lastSyncError.String
	}

	return &conn, nil
}

// ListConnectionsByAccountID retrieves all connections bound to a receivables account.
func (r *Repository) ListConnectionsByAccountID(ctx context.Context, accountID string) ([]domain.AcquirerConnection, error) {
	query := `
		SELECT id, receivables_account_id, acquirer_type, display_name,
		       credentials_encrypted, status, auto_sync_enabled, sync_interval_minutes,
		       last_sync_at, last_sync_status, last_sync_error,
		       created_at, updated_at
		FROM acquirer_connections
		WHERE receivables_account_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to list connections by account: %w", err)
	}
	defer rows.Close()

	var connections []domain.AcquirerConnection
	for rows.Next() {
		var conn domain.AcquirerConnection
		var lastSyncAt sql.NullTime
		var lastSyncStatus, lastSyncError sql.NullString

		if err := rows.Scan(
			&conn.ID, &conn.ReceivablesAccountID, &conn.AcquirerType, &conn.DisplayName,
			&conn.CredentialsEncrypted, &conn.Status, &conn.AutoSyncEnabled, &conn.SyncIntervalMinutes,
			&lastSyncAt, &lastSyncStatus, &lastSyncError,
			&conn.CreatedAt, &conn.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: failed to scan connection row: %w", err)
		}

		if lastSyncAt.Valid {
			conn.LastSyncAt = &lastSyncAt.Time
		}
		if lastSyncStatus.Valid {
			conn.LastSyncStatus = lastSyncStatus.String
		}
		if lastSyncError.Valid {
			conn.LastSyncError = &lastSyncError.String
		}

		connections = append(connections, conn)
	}

	return connections, rows.Err()
}

// ListActiveConnectionsForSync retrieves all active connections with auto-sync enabled.
func (r *Repository) ListActiveConnectionsForSync(ctx context.Context) ([]domain.AcquirerConnection, error) {
	query := `
		SELECT id, receivables_account_id, acquirer_type, display_name,
		       credentials_encrypted, status, auto_sync_enabled, sync_interval_minutes,
		       last_sync_at, last_sync_status, last_sync_error,
		       created_at, updated_at
		FROM acquirer_connections
		WHERE status = 'ACTIVE' AND auto_sync_enabled = true
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to list active connections: %w", err)
	}
	defer rows.Close()

	var connections []domain.AcquirerConnection
	for rows.Next() {
		var conn domain.AcquirerConnection
		var lastSyncAt sql.NullTime
		var lastSyncStatus, lastSyncError sql.NullString

		if err := rows.Scan(
			&conn.ID, &conn.ReceivablesAccountID, &conn.AcquirerType, &conn.DisplayName,
			&conn.CredentialsEncrypted, &conn.Status, &conn.AutoSyncEnabled, &conn.SyncIntervalMinutes,
			&lastSyncAt, &lastSyncStatus, &lastSyncError,
			&conn.CreatedAt, &conn.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: failed to scan active connection: %w", err)
		}

		if lastSyncAt.Valid {
			conn.LastSyncAt = &lastSyncAt.Time
		}
		if lastSyncStatus.Valid {
			conn.LastSyncStatus = lastSyncStatus.String
		}
		if lastSyncError.Valid {
			conn.LastSyncError = &lastSyncError.String
		}

		connections = append(connections, conn)
	}

	return connections, rows.Err()
}

// UpdateConnectionSyncState updates the last sync outcome.
func (r *Repository) UpdateConnectionSyncState(ctx context.Context, id string, syncStatus string, syncError *string, lastSyncAt time.Time) error {
	query := `
		UPDATE acquirer_connections
		SET last_sync_status = $2,
		    last_sync_error = $3,
		    last_sync_at = $4,
		    updated_at = NOW()
		WHERE id = $1
	`

	res, err := r.db.ExecContext(ctx, query, id, syncStatus, syncError, lastSyncAt)
	if err != nil {
		return fmt.Errorf("repository: failed to update connection sync state: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return ErrConnectionNotFound
	}

	return nil
}

// UpdateConnectionStatus updates the status enum.
func (r *Repository) UpdateConnectionStatus(ctx context.Context, id string, status domain.ConnectionStatus) error {
	query := `
		UPDATE acquirer_connections
		SET status = $2,
		    updated_at = NOW()
		WHERE id = $1
	`

	res, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("repository: failed to update connection status: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return ErrConnectionNotFound
	}

	return nil
}

// SaveSaleWithInstallments atomically records a sale and its corresponding installment schedules.
func (r *Repository) SaveSaleWithInstallments(ctx context.Context, sale *domain.Sale) error {
	if sale.ID == "" {
		sale.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	sale.CreatedAt = now

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repository: failed to begin sale transaction: %w", err)
	}
	defer tx.Rollback()

	saleQuery := `
		INSERT INTO acquirer_sales (
			id, connection_id, receivables_account_id, acquirer,
			nsu, tid, authorization_code, sale_date,
			gross_amount, total_mdr_fee, total_net_amount,
			installments_count, payment_method, card_brand, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (connection_id, nsu) DO UPDATE SET
			gross_amount = EXCLUDED.gross_amount,
			total_mdr_fee = EXCLUDED.total_mdr_fee,
			total_net_amount = EXCLUDED.total_net_amount
		RETURNING id
	`

	var saleID string
	err = tx.QueryRowContext(ctx, saleQuery,
		sale.ID, sale.ConnectionID, sale.ReceivablesAccountID, sale.Acquirer,
		sale.NSU, sale.TID, sale.AuthorizationCode, sale.SaleDate,
		sale.GrossAmount, sale.TotalMDRFee, sale.TotalNetAmount,
		sale.InstallmentsCount, sale.PaymentMethod, sale.CardBrand, sale.CreatedAt,
	).Scan(&saleID)
	if err != nil {
		return fmt.Errorf("repository: failed to insert/update sale: %w", err)
	}
	sale.ID = saleID

	instQuery := `
		INSERT INTO receivables_installments (
			id, sale_id, connection_id, receivables_account_id,
			acquirer, nsu, installment_number,
			gross_amount, mdr_fee, net_amount,
			expected_settlement_date, status,
			midaz_receivable_tx_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (connection_id, nsu, installment_number) DO UPDATE SET
			gross_amount = EXCLUDED.gross_amount,
			mdr_fee = EXCLUDED.mdr_fee,
			net_amount = EXCLUDED.net_amount,
			expected_settlement_date = EXCLUDED.expected_settlement_date,
			midaz_receivable_tx_id = COALESCE(receivables_installments.midaz_receivable_tx_id, EXCLUDED.midaz_receivable_tx_id),
			updated_at = NOW()
	`

	for idx := range sale.Installments {
		inst := &sale.Installments[idx]
		if inst.ID == "" {
			inst.ID = uuid.NewString()
		}
		inst.SaleID = sale.ID
		inst.ConnectionID = sale.ConnectionID
		inst.ReceivablesAccountID = sale.ReceivablesAccountID
		inst.Acquirer = sale.Acquirer
		inst.NSU = sale.NSU
		inst.CreatedAt = now
		inst.UpdatedAt = now
		if inst.Status == "" {
			inst.Status = domain.InstallmentStatusPendingSettlement
		}

		_, err = tx.ExecContext(ctx, instQuery,
			inst.ID, inst.SaleID, inst.ConnectionID, inst.ReceivablesAccountID,
			inst.Acquirer, inst.NSU, inst.InstallmentNumber,
			inst.GrossAmount, inst.MDRFee, inst.NetAmount,
			inst.ExpectedSettlementDate, inst.Status,
			inst.MidazReceivableTxID, inst.CreatedAt, inst.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("repository: failed to insert installment %d: %w", inst.InstallmentNumber, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repository: failed to commit sale transaction: %w", err)
	}

	return nil
}

// FindSaleByNSU finds a sale by connection and unique NSU.
func (r *Repository) FindSaleByNSU(ctx context.Context, connectionID, nsu string) (*domain.Sale, error) {
	query := `
		SELECT id, connection_id, receivables_account_id, acquirer,
		       nsu, tid, authorization_code, sale_date,
		       gross_amount, total_mdr_fee, total_net_amount,
		       installments_count, payment_method, card_brand, created_at
		FROM acquirer_sales
		WHERE connection_id = $1 AND nsu = $2
	`

	var sale domain.Sale
	var tid, authCode, cardBrand sql.NullString

	err := r.db.QueryRowContext(ctx, query, connectionID, nsu).Scan(
		&sale.ID, &sale.ConnectionID, &sale.ReceivablesAccountID, &sale.Acquirer,
		&sale.NSU, &tid, &authCode, &sale.SaleDate,
		&sale.GrossAmount, &sale.TotalMDRFee, &sale.TotalNetAmount,
		&sale.InstallmentsCount, &sale.PaymentMethod, &cardBrand, &sale.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSaleNotFound
		}
		return nil, fmt.Errorf("repository: failed to query sale: %w", err)
	}

	if tid.Valid {
		sale.TID = tid.String
	}
	if authCode.Valid {
		sale.AuthorizationCode = authCode.String
	}
	if cardBrand.Valid {
		sale.CardBrand = cardBrand.String
	}

	return &sale, nil
}

// FindInstallmentByReceivableCode queries an installment by acquirer, NSU, and installment index.
func (r *Repository) FindInstallmentByReceivableCode(ctx context.Context, acquirer domain.AcquirerType, nsu string, installmentNumber int) (*domain.Installment, error) {
	query := `
		SELECT id, sale_id, connection_id, receivables_account_id,
		       acquirer, nsu, installment_number,
		       gross_amount, mdr_fee, net_amount,
		       expected_settlement_date, actual_settlement_date, status,
		       midaz_receivable_tx_id, midaz_settlement_tx_id,
		       matcher_record_id, matcher_run_id,
		       created_at, updated_at
		FROM receivables_installments
		WHERE acquirer = $1 AND nsu = $2 AND installment_number = $3
	`

	var inst domain.Installment
	var actualSettlementDate sql.NullTime
	var midazRecTxID, midazSettleTxID, matcherRecordID, matcherRunID sql.NullString

	err := r.db.QueryRowContext(ctx, query, acquirer, nsu, installmentNumber).Scan(
		&inst.ID, &inst.SaleID, &inst.ConnectionID, &inst.ReceivablesAccountID,
		&inst.Acquirer, &inst.NSU, &inst.InstallmentNumber,
		&inst.GrossAmount, &inst.MDRFee, &inst.NetAmount,
		&inst.ExpectedSettlementDate, &actualSettlementDate, &inst.Status,
		&midazRecTxID, &midazSettleTxID,
		&matcherRecordID, &matcherRunID,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInstallmentNotFound
		}
		return nil, fmt.Errorf("repository: failed to query installment by code: %w", err)
	}

	if actualSettlementDate.Valid {
		inst.ActualSettlementDate = &actualSettlementDate.Time
	}
	if midazRecTxID.Valid {
		inst.MidazReceivableTxID = &midazRecTxID.String
	}
	if midazSettleTxID.Valid {
		inst.MidazSettlementTxID = &midazSettleTxID.String
	}
	if matcherRecordID.Valid {
		inst.MatcherRecordID = &matcherRecordID.String
	}
	if matcherRunID.Valid {
		inst.MatcherRunID = &matcherRunID.String
	}

	return &inst, nil
}

// ListPendingInstallmentsForMatcher lists pending installments within a settlement window.
func (r *Repository) ListPendingInstallmentsForMatcher(ctx context.Context, windowStart, windowEnd time.Time, limit int) ([]domain.Installment, error) {
	query := `
		SELECT id, sale_id, connection_id, receivables_account_id,
		       acquirer, nsu, installment_number,
		       gross_amount, mdr_fee, net_amount,
		       expected_settlement_date, actual_settlement_date, status,
		       midaz_receivable_tx_id, midaz_settlement_tx_id,
		       matcher_record_id, matcher_run_id,
		       created_at, updated_at
		FROM receivables_installments
		WHERE status = 'PENDING_SETTLEMENT'
		  AND expected_settlement_date BETWEEN $1 AND $2
		ORDER BY expected_settlement_date ASC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, windowStart, windowEnd, limit)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to query pending installments for matcher: %w", err)
	}
	defer rows.Close()

	var installments []domain.Installment
	for rows.Next() {
		var inst domain.Installment
		var actualSettlementDate sql.NullTime
		var midazRecTxID, midazSettleTxID, matcherRecordID, matcherRunID sql.NullString

		if err := rows.Scan(
			&inst.ID, &inst.SaleID, &inst.ConnectionID, &inst.ReceivablesAccountID,
			&inst.Acquirer, &inst.NSU, &inst.InstallmentNumber,
			&inst.GrossAmount, &inst.MDRFee, &inst.NetAmount,
			&inst.ExpectedSettlementDate, &actualSettlementDate, &inst.Status,
			&midazRecTxID, &midazSettleTxID,
			&matcherRecordID, &matcherRunID,
			&inst.CreatedAt, &inst.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: failed to scan installment: %w", err)
		}

		if actualSettlementDate.Valid {
			inst.ActualSettlementDate = &actualSettlementDate.Time
		}
		if midazRecTxID.Valid {
			inst.MidazReceivableTxID = &midazRecTxID.String
		}
		if midazSettleTxID.Valid {
			inst.MidazSettlementTxID = &midazSettleTxID.String
		}
		if matcherRecordID.Valid {
			inst.MatcherRecordID = &matcherRecordID.String
		}
		if matcherRunID.Valid {
			inst.MatcherRunID = &matcherRunID.String
		}

		installments = append(installments, inst)
	}

	return installments, rows.Err()
}

// MarkInstallmentUploadedToMatcher tags the installment with the Matcher record reference.
func (r *Repository) MarkInstallmentUploadedToMatcher(ctx context.Context, installmentID, matcherRecordID string) error {
	query := `
		UPDATE receivables_installments
		SET matcher_record_id = $2,
		    updated_at = NOW()
		WHERE id = $1
	`

	res, err := r.db.ExecContext(ctx, query, installmentID, matcherRecordID)
	if err != nil {
		return fmt.Errorf("repository: failed to update matcher record id: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return ErrInstallmentNotFound
	}

	return nil
}

// UpdateInstallmentReceivableTxID links the initial D+0 Midaz transaction ID to the installment.
func (r *Repository) UpdateInstallmentReceivableTxID(ctx context.Context, installmentID, txID string) error {
	query := `
		UPDATE receivables_installments
		SET midaz_receivable_tx_id = $2,
		    updated_at = NOW()
		WHERE id = $1
	`

	res, err := r.db.ExecContext(ctx, query, installmentID, txID)
	if err != nil {
		return fmt.Errorf("repository: failed to update installment receivable tx id: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return ErrInstallmentNotFound
	}

	return nil
}

// MarkInstallmentReconciled flags the installment as reconciled and links the settlement transaction.
func (r *Repository) MarkInstallmentReconciled(ctx context.Context, installmentID, midazSettlementTxID, matcherRunID string, settlementDate time.Time) error {
	query := `
		UPDATE receivables_installments
		SET status = 'RECONCILED',
		    midaz_settlement_tx_id = $2,
		    matcher_run_id = $3,
		    actual_settlement_date = $4,
		    updated_at = NOW()
		WHERE id = $1
	`

	res, err := r.db.ExecContext(ctx, query, installmentID, midazSettlementTxID, matcherRunID, settlementDate)
	if err != nil {
		return fmt.Errorf("repository: failed to mark installment reconciled: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return ErrInstallmentNotFound
	}

	return nil
}

// GetReceivablesCalendar queries the installment agenda with dynamic filters.
func (r *Repository) GetReceivablesCalendar(ctx context.Context, filter domain.CalendarFilter) ([]domain.Installment, error) {
	var conditions []string
	var args []any
	argID := 1

	conditions = append(conditions, fmt.Sprintf("receivables_account_id = $%d", argID))
	args = append(args, filter.ReceivablesAccountID)
	argID++

	if filter.Acquirer != nil {
		conditions = append(conditions, fmt.Sprintf("acquirer = $%d", argID))
		args = append(args, *filter.Acquirer)
		argID++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argID))
		args = append(args, *filter.Status)
		argID++
	}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("expected_settlement_date >= $%d", argID))
		args = append(args, *filter.StartDate)
		argID++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("expected_settlement_date <= $%d", argID))
		args = append(args, *filter.EndDate)
		argID++
	}

	limit := 100
	if filter.Limit > 0 {
		limit = filter.Limit
	}
	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	query := fmt.Sprintf(`
		SELECT id, sale_id, connection_id, receivables_account_id,
		       acquirer, nsu, installment_number,
		       gross_amount, mdr_fee, net_amount,
		       expected_settlement_date, actual_settlement_date, status,
		       midaz_receivable_tx_id, midaz_settlement_tx_id,
		       matcher_record_id, matcher_run_id,
		       created_at, updated_at
		FROM receivables_installments
		WHERE %s
		ORDER BY expected_settlement_date ASC
		LIMIT $%d OFFSET $%d
	`, strings.Join(conditions, " AND "), argID, argID+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to query receivables calendar: %w", err)
	}
	defer rows.Close()

	var installments []domain.Installment
	for rows.Next() {
		var inst domain.Installment
		var actualSettlementDate sql.NullTime
		var midazRecTxID, midazSettleTxID, matcherRecordID, matcherRunID sql.NullString

		if err := rows.Scan(
			&inst.ID, &inst.SaleID, &inst.ConnectionID, &inst.ReceivablesAccountID,
			&inst.Acquirer, &inst.NSU, &inst.InstallmentNumber,
			&inst.GrossAmount, &inst.MDRFee, &inst.NetAmount,
			&inst.ExpectedSettlementDate, &actualSettlementDate, &inst.Status,
			&midazRecTxID, &midazSettleTxID,
			&matcherRecordID, &matcherRunID,
			&inst.CreatedAt, &inst.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: failed to scan calendar installment: %w", err)
		}

		if actualSettlementDate.Valid {
			inst.ActualSettlementDate = &actualSettlementDate.Time
		}
		if midazRecTxID.Valid {
			inst.MidazReceivableTxID = &midazRecTxID.String
		}
		if midazSettleTxID.Valid {
			inst.MidazSettlementTxID = &midazSettleTxID.String
		}
		if matcherRecordID.Valid {
			inst.MatcherRecordID = &matcherRecordID.String
		}
		if matcherRunID.Valid {
			inst.MatcherRunID = &matcherRunID.String
		}

		installments = append(installments, inst)
	}

	return installments, rows.Err()
}
