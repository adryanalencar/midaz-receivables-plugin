package domain

import (
	"context"
	"time"
)

// CalendarFilter specifies query options for retrieving installment schedules.
type CalendarFilter struct {
	ReceivablesAccountID string
	Acquirer             *AcquirerType
	Status               *InstallmentStatus
	StartDate            *time.Time
	EndDate              *time.Time
	Limit                int
	Offset               int
}

// IReceivablesRepository defines the database persistence operations.
type IReceivablesRepository interface {
	// ReceivablesAccount operations
	CreateReceivablesAccount(ctx context.Context, account *ReceivablesAccount) error
	FindReceivablesAccountByID(ctx context.Context, id string) (*ReceivablesAccount, error)
	FindReceivablesAccountByAlias(ctx context.Context, alias string) (*ReceivablesAccount, error)
	ListReceivablesAccounts(ctx context.Context) ([]ReceivablesAccount, error)

	// AcquirerConnection operations
	CreateAcquirerConnection(ctx context.Context, conn *AcquirerConnection) error
	FindAcquirerConnectionByID(ctx context.Context, id string) (*AcquirerConnection, error)
	ListConnectionsByAccountID(ctx context.Context, accountID string) ([]AcquirerConnection, error)
	ListActiveConnectionsForSync(ctx context.Context) ([]AcquirerConnection, error)
	UpdateConnectionSyncState(ctx context.Context, id string, syncStatus string, syncError *string, lastSyncAt time.Time) error
	UpdateConnectionStatus(ctx context.Context, id string, status ConnectionStatus) error

	// Sales & Installments operations
	SaveSaleWithInstallments(ctx context.Context, sale *Sale) error
	FindSaleByNSU(ctx context.Context, connectionID, nsu string) (*Sale, error)
	FindInstallmentByReceivableCode(ctx context.Context, acquirer AcquirerType, nsu string, installmentNumber int) (*Installment, error)
	ListPendingInstallmentsForMatcher(ctx context.Context, windowStart, windowEnd time.Time, limit int) ([]Installment, error)
	MarkInstallmentUploadedToMatcher(ctx context.Context, installmentID, matcherRecordID string) error
	UpdateInstallmentReceivableTxID(ctx context.Context, installmentID, txID string) error
	MarkInstallmentReconciled(ctx context.Context, installmentID, midazSettlementTxID, matcherRunID string, settlementDate time.Time) error
	GetReceivablesCalendar(ctx context.Context, filter CalendarFilter) ([]Installment, error)
}
