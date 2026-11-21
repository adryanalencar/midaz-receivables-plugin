package domain

import (
	"time"
)

// ReceivablesAccount represents a merchant's designated receivables wallet in Midaz,
// tied to a Mother Account (CNPJ profile) and linked to dedicated subaccounts.
type ReceivablesAccount struct {
	ID                   string    `json:"id"`
	OrganizationID       string    `json:"organizationId"`
	LedgerID             string    `json:"ledgerId"`
	MidazAccountID       string    `json:"midazAccountId"`
	ParentAccountID      *string   `json:"parentAccountId,omitempty"`      // UUID of the Mother Account in Midaz
	MotherAccountAlias   string    `json:"motherAccountAlias,omitempty"`   // e.g. "@merchant_alpha"
	AccountAlias           string    `json:"accountAlias"`                     // e.g. "@merchant_alpha/receivables"
	ExternalAccountAlias   string    `json:"externalAccountAlias,omitempty"`   // Chosen external funding origin: e.g. "@external/BRL:universal"
	MDRRevenueAccountAlias string    `json:"mdrRevenueAccountAlias,omitempty"` // Subaccount for MDR: "@merchant_alpha/receivables/mdr_revenue"
	FeeAccountAlias        string    `json:"feeAccountAlias,omitempty"`        // Subaccount for tariffs/fees: "@merchant_alpha/receivables/fees"
	CheckingAccountAlias   string    `json:"checkingAccountAlias"`             // Default checking/openfinance alias
	MatcherContextID     string    `json:"matcherContextId,omitempty"`     // Dedicated Context ID in Matcher 6 for this tenant
	Name                 string    `json:"name"`
	Status               string    `json:"status"`                         // "ACTIVE", "INACTIVE"
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

// IsActive returns whether the receivables account is active.
func (a *ReceivablesAccount) IsActive() bool {
	return a.Status == "ACTIVE"
}
