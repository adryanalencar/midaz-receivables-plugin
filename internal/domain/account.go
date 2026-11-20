package domain

import (
	"time"
)

// ReceivablesAccount represents a merchant's designated receivables wallet in Midaz.
type ReceivablesAccount struct {
	ID                   string    `json:"id"`
	OrganizationID       string    `json:"organizationId"`
	LedgerID             string    `json:"ledgerId"`
	MidazAccountID       string    `json:"midazAccountId"`
	AccountAlias         string    `json:"accountAlias"`
	CheckingAccountAlias string    `json:"checkingAccountAlias"`
	Name                 string    `json:"name"`
	Status               string    `json:"status"` // "ACTIVE", "INACTIVE"
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

// IsActive returns whether the receivables account is active.
func (a *ReceivablesAccount) IsActive() bool {
	return a.Status == "ACTIVE"
}
