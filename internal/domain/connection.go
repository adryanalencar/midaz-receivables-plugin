package domain

import (
	"time"
)

type ConnectionStatus string

const (
	ConnectionStatusActive   ConnectionStatus = "ACTIVE"
	ConnectionStatusInactive ConnectionStatus = "INACTIVE"
	ConnectionStatusExpired  ConnectionStatus = "EXPIRED"
	ConnectionStatusError    ConnectionStatus = "ERROR"
)

// AcquirerConnection stores encrypted credentials and sync state for a specific acquirer.
type AcquirerConnection struct {
	ID                   string           `json:"id"`
	ReceivablesAccountID string           `json:"receivablesAccountId"`
	AcquirerType         AcquirerType     `json:"acquirerType"` // "SUMUP", "PAGBANK"
	DisplayName          string           `json:"displayName"`
	CredentialsEncrypted string           `json:"-"` // AES-256-GCM base64 ciphertext (never marshaled to JSON)
	Status               ConnectionStatus `json:"status"`
	AutoSyncEnabled      bool             `json:"autoSyncEnabled"`
	SyncIntervalMinutes  int              `json:"syncIntervalMinutes"`
	LastSyncAt           *time.Time       `json:"lastSyncAt,omitempty"`
	LastSyncStatus       string           `json:"lastSyncStatus,omitempty"`
	LastSyncError        *string          `json:"lastSyncError,omitempty"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
}

// ConnectionCredentials represents the in-memory decrypted credentials structure.
type ConnectionCredentials struct {
	APIKey               string         `json:"api_key,omitempty"`
	MerchantCode         string         `json:"merchant_code,omitempty"`
	ClientID             string         `json:"client_id,omitempty"`
	ClientSecret         string         `json:"client_secret,omitempty"`
	Token                string         `json:"token,omitempty"`
	WebhookSecret        string         `json:"webhook_secret,omitempty"`
	AdditionalProperties map[string]any `json:"additional_properties,omitempty"`
}

// IsActive returns whether the acquirer connection is operational.
func (c *AcquirerConnection) IsActive() bool {
	return c.Status == ConnectionStatusActive
}
