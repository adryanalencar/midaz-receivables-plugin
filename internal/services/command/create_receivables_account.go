package command

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrEmptyAccountAlias = errors.New("create_receivables_account: alias is required")
	ErrEmptyAccountName  = errors.New("create_receivables_account: name is required")
)

type CreateReceivablesAccountInput struct {
	OrganizationID              string         `json:"organizationId"`
	LedgerID                    string         `json:"ledgerId"`
	Alias                       string         `json:"alias"`
	Name                        string         `json:"name"`
	CNPJ                        string         `json:"cnpj"`
	DefaultCheckingAccountAlias string         `json:"defaultCheckingAccountAlias"`
	Metadata                    map[string]any `json:"metadata"`
}

type CreateReceivablesAccountHandler struct {
	repo        domain.IReceivablesRepository
	midazClient domain.IMidazLedgerClient
}

func NewCreateReceivablesAccountHandler(repo domain.IReceivablesRepository, midazClient domain.IMidazLedgerClient) *CreateReceivablesAccountHandler {
	return &CreateReceivablesAccountHandler{
		repo:        repo,
		midazClient: midazClient,
	}
}

func (h *CreateReceivablesAccountHandler) Execute(ctx context.Context, input CreateReceivablesAccountInput) (*domain.ReceivablesAccount, error) {
	if strings.TrimSpace(input.Alias) == "" {
		return nil, ErrEmptyAccountAlias
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, ErrEmptyAccountName
	}

	alias := input.Alias
	if !strings.HasPrefix(alias, "@") {
		alias = "@" + alias
	}
	if !strings.HasSuffix(alias, "/receivables") {
		alias = alias + "/receivables"
	}

	checkingAlias := input.DefaultCheckingAccountAlias
	if checkingAlias != "" && !strings.HasPrefix(checkingAlias, "@") {
		checkingAlias = "@" + checkingAlias
	}

	// Provision account in Midaz
	midazAlias, err := h.midazClient.CreateAccount(ctx, input.OrganizationID, input.LedgerID, alias, input.Name, "creditCard")
	if err != nil {
		return nil, fmt.Errorf("create_receivables_account: failed to provision account in Midaz: %w", err)
	}

	now := time.Now().UTC()
	account := &domain.ReceivablesAccount{
		ID:                   uuid.NewString(),
		OrganizationID:       input.OrganizationID,
		LedgerID:             input.LedgerID,
		MidazAccountID:       midazAlias,
		AccountAlias:         alias,
		CheckingAccountAlias: checkingAlias,
		Name:                 input.Name,
		Status:               "ACTIVE",
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := h.repo.CreateReceivablesAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("create_receivables_account: failed to persist in repository: %w", err)
	}

	return account, nil
}
