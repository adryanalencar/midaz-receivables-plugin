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

// EnsureClearingAccount verifies or provisions the universal @external/BRL:universal account in Midaz.
func (c *MidazClient) EnsureClearingAccount(ctx context.Context, organizationID, ledgerID string, acquirer domain.AcquirerType) (string, error) {
	alias := domain.UniversalFiatClearingAccount
	if c.cfg != nil && c.cfg.MidazClearingAccountAlias != "" {
		alias = c.cfg.MidazClearingAccountAlias
	}

	exists, err := c.CheckAccountExists(ctx, organizationID, ledgerID, alias)
	if err != nil {
		return "", fmt.Errorf("midaz: failed checking universal clearing account: %w", err)
	}
	if exists {
		return alias, nil
	}

	name := "Universal Fiat External Account"
	return c.CreateAccount(ctx, organizationID, ledgerID, alias, name, "external", nil)
}

// EnsureAcquirerInternalAccount verifies or provisions the internal liquidity account (e.g. @acquirer/sumup) in Midaz.
func (c *MidazClient) EnsureAcquirerInternalAccount(ctx context.Context, organizationID, ledgerID string, acquirer domain.AcquirerType) (string, error) {
	alias := domain.AcquirerInternalAccountAlias(acquirer)

	exists, err := c.CheckAccountExists(ctx, organizationID, ledgerID, alias)
	if err != nil {
		return "", fmt.Errorf("midaz: failed checking acquirer internal account: %w", err)
	}
	if exists {
		return alias, nil
	}

	name := fmt.Sprintf("%s Acquirer Liquidity Account", acquirer)
	return c.CreateAccount(ctx, organizationID, ledgerID, alias, name, "deposit", nil)
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

// ListExternalAccounts queries Midaz onboarding and filters accounts with type="external".
func (c *MidazClient) ListExternalAccounts(ctx context.Context, organizationID, ledgerID string) ([]domain.ExternalAccount, error) {
	url := fmt.Sprintf("%s/v1/organizations/%s/ledgers/%s/accounts?limit=100",
		c.cfg.MidazOnboardingURL, organizationID, ledgerID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("midaz: failed creating list accounts request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("midaz: list accounts request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("midaz: list accounts rejected (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var pageResp struct {
		Items []struct {
			ID        string  `json:"id"`
			Name      string  `json:"name"`
			Alias     *string `json:"alias"`
			Type      string  `json:"type"`
			AssetCode string  `json:"assetCode"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
		return nil, fmt.Errorf("midaz: failed parsing accounts response: %w", err)
	}

	var externalAccounts []domain.ExternalAccount
	for _, item := range pageResp.Items {
		if strings.EqualFold(item.Type, "external") {
			alias := ""
			if item.Alias != nil {
				alias = *item.Alias
			}
			externalAccounts = append(externalAccounts, domain.ExternalAccount{
				ID:        item.ID,
				Alias:     alias,
				Name:      item.Name,
				AssetCode: item.AssetCode,
				Type:      item.Type,
			})
		}
	}

	return externalAccounts, nil
}

// CreateAccount provisions an account in Midaz, supporting native parentAccountId hierarchy.
func (c *MidazClient) CreateAccount(ctx context.Context, organizationID, ledgerID, alias, name, accountType string, parentAccountID *string) (string, error) {
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
	if parentAccountID != nil && *parentAccountID != "" {
		payload["parentAccountId"] = *parentAccountID
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

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusConflict {
		// Account already exists
		return alias, nil
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("midaz: account creation rejected (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var createdAcc struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &createdAcc); err == nil && createdAcc.ID != "" {
		return createdAcc.ID, nil
	}

	return alias, nil
}

// PostTransaction executes a double-entry transaction in Midaz (supports pending: true).
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
		"pending":                  tx.Pending,
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

	var respData struct {
		ID   string `json:"id"`
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&respData)
	if respData.ID != "" {
		return respData.ID, nil
	}

	return tx.Code, nil
}

// CommitTransaction confirms a pending transaction in Midaz, moving funds to available balance.
func (c *MidazClient) CommitTransaction(ctx context.Context, organizationID, ledgerID, transactionID string) error {
	url := fmt.Sprintf("%s/v1/organizations/%s/ledgers/%s/transactions/%s/commit",
		c.cfg.MidazTransactionURL, organizationID, ledgerID, transactionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		return fmt.Errorf("midaz: failed creating commit request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("midaz: commit transaction request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		// Already committed
		return nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("midaz: commit transaction rejected (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// PostMultiSourceTransaction posts an atomic N:1 transaction transferring from N receivables sources to 1 destination.
func (c *MidazClient) PostMultiSourceTransaction(ctx context.Context, organizationID, ledgerID string, tx *domain.MultiSourceTransactionPayload) (string, error) {
	if tx == nil {
		return "", fmt.Errorf("midaz: nil multi-source transaction payload")
	}

	url := fmt.Sprintf("%s/v1/organizations/%s/ledgers/%s/transactions/json",
		c.cfg.MidazTransactionURL, organizationID, ledgerID)

	scale := int32(2)
	centsValue := tx.TotalAmount.Mul(decimal.NewFromInt(100)).IntPart()
	asset := tx.AssetCode
	if asset == "" {
		asset = "BRL"
	}

	fromLegs := make([]map[string]any, len(tx.Sources))
	for i, src := range tx.Sources {
		fromLegs[i] = map[string]any{
			"accountAlias": src.AccountAlias,
			"amount": map[string]any{
				"asset": asset,
				"value": src.Amount.StringFixed(2),
			},
			"description": src.Description,
		}
	}

	toLegs := make([]map[string]any, 0, 1+len(tx.Deductions))
	primaryAmount := tx.TotalAmount
	for _, ded := range tx.Deductions {
		primaryAmount = primaryAmount.Sub(ded.Amount)
	}

	toLegs = append(toLegs, map[string]any{
		"accountAlias": tx.DestinationAccountAlias,
		"amount": map[string]any{
			"asset": asset,
			"value": primaryAmount.StringFixed(2),
		},
		"description": tx.Description,
	})

	for _, ded := range tx.Deductions {
		toLegs = append(toLegs, map[string]any{
			"accountAlias": ded.AccountAlias,
			"amount": map[string]any{
				"asset": asset,
				"value": ded.Amount.StringFixed(2),
			},
			"description": ded.Description,
		})
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
				"from": fromLegs,
			},
			"distribute": map[string]any{
				"to": toLegs,
			},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("midaz: failed marshaling multi-source transaction: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("midaz: failed creating multi-source transaction request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("midaz: multi-source transaction request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return tx.Code, nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("midaz: multi-source transaction rejected (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var respData struct {
		ID   string `json:"id"`
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&respData)
	if respData.ID != "" {
		return respData.ID, nil
	}

	return tx.Code, nil
}

func (c *MidazClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.MidazAuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.MidazAuthToken)
	}
}
