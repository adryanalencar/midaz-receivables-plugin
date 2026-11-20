package out

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
)

// MidazClient handles outbound communication with Midaz Onboarding and Transaction APIs.
type MidazClient struct {
	cfg        *bootstrap.Config
	httpClient *http.Client
}

// NewMidazClient constructs a new Midaz API client.
func NewMidazClient(cfg *bootstrap.Config, httpClient *http.Client) *MidazClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &MidazClient{
		cfg:        cfg,
		httpClient: httpClient,
	}
}

// EnsureClearingAccount verifies or provisions the universal @external/acquirer_<name> account in Midaz.
func (c *MidazClient) EnsureClearingAccount(ctx context.Context, organizationID, ledgerID string, acquirer domain.AcquirerType) (string, error) {
	alias := fmt.Sprintf("@external/acquirer_%s", strings.ToLower(string(acquirer)))

	exists, err := c.CheckAccountExists(ctx, organizationID, ledgerID, alias)
	if err != nil {
		return "", fmt.Errorf("midaz: failed checking clearing account: %w", err)
	}
	if exists {
		return alias, nil
	}

	name := fmt.Sprintf("Universal Clearing - %s", acquirer)
	return c.CreateAccount(ctx, organizationID, ledgerID, alias, name, "external")
}

// CheckAccountExists verifies whether an account alias is registered in Midaz.
func (c *MidazClient) CheckAccountExists(ctx context.Context, organizationID, ledgerID, alias string) (bool, error) {
	url := fmt.Sprintf("%s/v1/organizations/%s/ledgers/%s/accounts?alias=%s",
		c.cfg.MidazOnboardingURL, organizationID, ledgerID, alias)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("midaz: failed creating check account request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("midaz: check account request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("midaz: unexpected check account status HTTP %d: %s", resp.StatusCode, string(body))
}

// CreateAccount provisions an account in Midaz.
func (c *MidazClient) CreateAccount(ctx context.Context, organizationID, ledgerID, alias, name, accountType string) (string, error) {
	url := fmt.Sprintf("%s/v1/organizations/%s/ledgers/%s/accounts",
		c.cfg.MidazOnboardingURL, organizationID, ledgerID)

	payload := map[string]any{
		"name":      name,
		"assetCode": "BRL",
		"type":      accountType,
		"alias":     alias,
		"metadata": map[string]string{
			"source": "midaz-plugin-receivables",
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("midaz: failed to marshal account payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("midaz: failed creating account request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("midaz: create account request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		// Account already exists
		return alias, nil
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("midaz: account creation rejected (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return alias, nil
}

// PostTransaction executes a double-entry transaction in Midaz balancing debits and credits.
func (c *MidazClient) PostTransaction(ctx context.Context, organizationID, ledgerID string, tx *domain.MidazTransactionPayload) (string, error) {
	if tx == nil {
		return "", fmt.Errorf("midaz: nil transaction payload")
	}

	url := fmt.Sprintf("%s/v1/organizations/%s/ledgers/%s/transactions/json",
		c.cfg.MidazTransactionURL, organizationID, ledgerID)

	scale := int32(2)
	centsValue := tx.Amount.Mul(decimal.NewFromInt(100)).IntPart()
	asset := tx.AssetCode
	if asset == "" {
		asset = "BRL"
	}

	// Prepare distribution legs
	var toLegs []map[string]any
	if tx.FeeAmount.GreaterThan(decimal.Zero) && tx.FeeAccountAlias != "" {
		netAmount := tx.Amount.Sub(tx.FeeAmount)
		toLegs = []map[string]any{
			{
				"accountAlias": tx.DestinationAccountAlias,
				"amount": map[string]any{
					"asset": asset,
					"value": netAmount.StringFixed(2),
				},
				"description": "Net receivable allocation",
			},
			{
				"accountAlias": tx.FeeAccountAlias,
				"amount": map[string]any{
					"asset": asset,
					"value": tx.FeeAmount.StringFixed(2),
				},
				"description": "MDR fee deduction",
			},
		}
	} else {
		toLegs = []map[string]any{
			{
				"accountAlias": tx.DestinationAccountAlias,
				"amount": map[string]any{
					"asset": asset,
					"value": tx.Amount.StringFixed(2),
				},
				"description": "Full transfer amount",
			},
		}
	}

	payload := map[string]any{
		"chartOfAccountsGroupName": tx.ChartOfAccountsGroupName,
		"description":              tx.Description,
		"code":                     tx.Code,
		"metadata":                 tx.Metadata,
		"send": map[string]any{
			"asset": asset,
			"value": centsValue,
			"scale": scale,
			"source": map[string]any{
				"from": []map[string]any{
					{
						"accountAlias": tx.SourceAccountAlias,
						"amount": map[string]any{
							"asset": asset,
							"value": tx.Amount.StringFixed(2),
						},
						"description": "Source debit",
					},
				},
			},
			"distribute": map[string]any{
				"to": toLegs,
			},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("midaz: failed to marshal transaction payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("midaz: failed creating transaction request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("midaz: post transaction request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		// Idempotent: transaction with this code was already processed
		return tx.Code, nil
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("midaz: transaction rejected (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return tx.Code, nil
}

func (c *MidazClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.MidazAuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.MidazAuthToken)
	}
}
