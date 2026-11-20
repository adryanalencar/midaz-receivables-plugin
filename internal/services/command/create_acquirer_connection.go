package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/providers"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/crypto"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrMissingReceivablesAccountID = errors.New("create_acquirer_connection: receivables_account_id is required")
	ErrUnsupportedAcquirerType     = errors.New("create_acquirer_connection: acquirer type is not supported")
)

type CreateAcquirerConnectionInput struct {
	ReceivablesAccountID string                      `json:"receivablesAccountId"`
	Acquirer             domain.AcquirerType         `json:"acquirer"`
	DisplayName          string                      `json:"displayName"`
	Credentials          domain.ConnectionCredentials `json:"credentials"`
	WebhookSecret        string                      `json:"webhookSecret,omitempty"`
	AutoSyncEnabled      bool                        `json:"autoSyncEnabled"`
	SyncIntervalMinutes  int                         `json:"syncIntervalMinutes"`
	Metadata             map[string]any              `json:"metadata,omitempty"`
}

type CreateAcquirerConnectionHandler struct {
	repo            domain.IReceivablesRepository
	providerFactory *providers.Factory
	encryptor       crypto.Encryptor
}

func NewCreateAcquirerConnectionHandler(
	repo domain.IReceivablesRepository,
	providerFactory *providers.Factory,
	encryptor crypto.Encryptor,
) *CreateAcquirerConnectionHandler {
	return &CreateAcquirerConnectionHandler{
		repo:            repo,
		providerFactory: providerFactory,
		encryptor:       encryptor,
	}
}

func (h *CreateAcquirerConnectionHandler) Execute(ctx context.Context, input CreateAcquirerConnectionInput) (*domain.AcquirerConnection, error) {
	if strings.TrimSpace(input.ReceivablesAccountID) == "" {
		return nil, ErrMissingReceivablesAccountID
	}

	// 1. Verify that the parent ReceivablesAccount exists
	acc, err := h.repo.FindReceivablesAccountByID(ctx, input.ReceivablesAccountID)
	if err != nil {
		return nil, fmt.Errorf("create_acquirer_connection: failed looking up parent account: %w", err)
	}
	if acc == nil {
		return nil, fmt.Errorf("create_acquirer_connection: receivables account %s not found", input.ReceivablesAccountID)
	}

	// 2. Resolve Acquirer Provider adapter
	provider, err := h.providerFactory.Get(input.Acquirer)
	if err != nil {
		return nil, fmt.Errorf("create_acquirer_connection: %w", err)
	}

	// 3. Validate credentials directly with the Acquirer
	if err := provider.ValidateCredentials(ctx, &input.Credentials); err != nil {
		return nil, fmt.Errorf("create_acquirer_connection: credential validation failed: %w", err)
	}

	// 4. Encrypt credentials at rest using AES-256-GCM AEAD
	credBytes, err := json.Marshal(input.Credentials)
	if err != nil {
		return nil, fmt.Errorf("create_acquirer_connection: failed to serialize credentials: %w", err)
	}

	encryptedCreds, err := h.encryptor.Encrypt(credBytes)
	if err != nil {
		return nil, fmt.Errorf("create_acquirer_connection: encryption failed: %w", err)
	}

	displayName := input.DisplayName
	if displayName == "" {
		displayName = string(input.Acquirer)
	}

	syncInterval := input.SyncIntervalMinutes
	if syncInterval <= 0 {
		syncInterval = 60
	}

	now := time.Now().UTC()
	conn := &domain.AcquirerConnection{
		ID:                   uuid.NewString(),
		ReceivablesAccountID: input.ReceivablesAccountID,
		AcquirerType:         input.Acquirer,
		DisplayName:          displayName,
		CredentialsEncrypted: encryptedCreds,
		Status:               domain.ConnectionStatusActive,
		AutoSyncEnabled:      input.AutoSyncEnabled,
		SyncIntervalMinutes:  syncInterval,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := h.repo.CreateAcquirerConnection(ctx, conn); err != nil {
		return nil, fmt.Errorf("create_acquirer_connection: failed to save connection: %w", err)
	}

	return conn, nil
}
