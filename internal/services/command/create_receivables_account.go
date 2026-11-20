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
	ParentAccountID             *string        `json:"parentAccountId,omitempty"`    // Mother Account ID in Midaz (CNPJ Master)
	MotherAccountAlias          string         `json:"motherAccountAlias,omitempty"` // e.g. "@merchant_alpha"
	Alias                       string         `json:"alias"`                        // e.g. "@merchant_alpha/receivables"
	Name                        string         `json:"name"`
	CNPJ                        string         `json:"cnpj"`
	DefaultCheckingAccountAlias string         `json:"defaultCheckingAccountAlias"`
	Metadata                    map[string]any `json:"metadata"`
}

type CreateReceivablesAccountHandler struct {
	repo          domain.IReceivablesRepository
	midazClient   domain.IMidazLedgerClient
	matcherClient domain.IMatcherClient
}

func NewCreateReceivablesAccountHandler(
	repo domain.IReceivablesRepository,
	midazClient domain.IMidazLedgerClient,
	matcherClient domain.IMatcherClient,
) *CreateReceivablesAccountHandler {
	return &CreateReceivablesAccountHandler{
		repo:          repo,
		midazClient:   midazClient,
		matcherClient: matcherClient,
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

	feeAlias := alias + "/fees"

	checkingAlias := input.DefaultCheckingAccountAlias
	if checkingAlias != "" && !strings.HasPrefix(checkingAlias, "@") {
		checkingAlias = "@" + checkingAlias
	}

	// 1. Provision Receivables Subaccount in Midaz with native parentAccountId
	midazAccountID, err := h.midazClient.CreateAccount(ctx, input.OrganizationID, input.LedgerID, alias, input.Name, "creditCard", input.ParentAccountID)
	if err != nil {
		return nil, fmt.Errorf("create_receivables_account: failed to provision receivables account in Midaz: %w", err)
	}

	// 2. Provision Tariff/Fee Subaccount in Midaz (@<merchant>/receivables/fees)
	_, _ = h.midazClient.CreateAccount(ctx, input.OrganizationID, input.LedgerID, feeAlias, input.Name+" - Tariffs & Deductions", "deposit", input.ParentAccountID)

	// 3. Ensure dedicated tenant Context in Matcher 6 if matcher client is configured
	matcherContextID := ""
	if h.matcherClient != nil {
		ctxID, _, _, err := h.matcherClient.EnsureContextAndSources(ctx)
		if err == nil {
			matcherContextID = ctxID
		}
	}

	now := time.Now().UTC()
	account := &domain.ReceivablesAccount{
		ID:                   uuid.NewString(),
		OrganizationID:       input.OrganizationID,
		LedgerID:             input.LedgerID,
		MidazAccountID:       midazAccountID,
		ParentAccountID:      input.ParentAccountID,
		MotherAccountAlias:   input.MotherAccountAlias,
		AccountAlias:         alias,
		FeeAccountAlias:      feeAlias,
		CheckingAccountAlias: checkingAlias,
		MatcherContextID:     matcherContextID,
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
