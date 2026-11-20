package out

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
)

// MatcherClient implements IMatcherClient connecting to Matcher 6 on port :4018.
type MatcherClient struct {
	cfg        *bootstrap.Config
	httpClient *http.Client
}

// NewMatcherClient creates a new Matcher 6 API client.
func NewMatcherClient(cfg *bootstrap.Config, httpClient *http.Client) *MatcherClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &MatcherClient{
		cfg:        cfg,
		httpClient: httpClient,
	}
}

// HealthCheck verifies connectivity to Matcher 6.
func (c *MatcherClient) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.MatcherURL+"/version", nil)
	if err != nil {
		return fmt.Errorf("matcher: health check request failed: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("matcher: health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("matcher: health check returned HTTP %d", resp.StatusCode)
	}
	return nil
}

type ContextResponse struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type ContextListResponse struct {
	Items []ContextResponse `json:"items"`
}

type SourceResponse struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type SourceListResponse struct {
	Items []SourceResponse `json:"items"`
}

// EnsureContextAndSources idempotently provisions the reconciliation Context and Sources in Matcher 6.
func (c *MatcherClient) EnsureContextAndSources(ctx context.Context) (string, string, string, error) {
	slug := c.cfg.MatcherContextSlug
	if slug == "" {
		slug = "acquirers-settlement"
	}

	// 1. Find or create Context
	contextID, err := c.findOrCreateContext(ctx, slug)
	if err != nil {
		return "", "", "", fmt.Errorf("matcher: failed to ensure context: %w", err)
	}

	// 2. Find or create Source A (Acquirer Receivables Schedule)
	sourceAID, err := c.findOrCreateSource(ctx, contextID, "acquirer_receivables", "acquirer-receivables")
	if err != nil {
		return "", "", "", fmt.Errorf("matcher: failed to ensure source A: %w", err)
	}

	// 3. Find or create Source B (Bank Statement Credits)
	sourceBID, err := c.findOrCreateSource(ctx, contextID, "bank_statement_credits", "bank-statement-credits")
	if err != nil {
		return "", "", "", fmt.Errorf("matcher: failed to ensure source B: %w", err)
	}

	return contextID, sourceAID, sourceBID, nil
}

func (c *MatcherClient) findOrCreateContext(ctx context.Context, slug string) (string, error) {
	// Query existing contexts
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.MatcherURL+"/v1/contexts", nil)
	if err != nil {
		return "", err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var list ContextListResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err == nil {
			for _, item := range list.Items {
				if item.Slug == slug {
					return item.ID, nil
				}
			}
		}
	}

	// Create new context
	createPayload := map[string]string{
		"name":        "Acquirers Settlement Reconciliation",
		"slug":        slug,
		"description": "Cross-reconciliation between Acquirer Receivables Schedules and Open Finance Bank Statements",
	}
	body, _ := json.Marshal(createPayload)

	reqCreate, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.MatcherURL+"/v1/contexts", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	c.setHeaders(reqCreate)

	respCreate, err := c.httpClient.Do(reqCreate)
	if err != nil {
		return "", err
	}
	defer respCreate.Body.Close()

	if respCreate.StatusCode != http.StatusOK && respCreate.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(respCreate.Body)
		return "", fmt.Errorf("failed creating context (HTTP %d): %s", respCreate.StatusCode, string(respBody))
	}

	var created ContextResponse
	if err := json.NewDecoder(respCreate.Body).Decode(&created); err != nil {
		return "", err
	}
	return created.ID, nil
}

func (c *MatcherClient) findOrCreateSource(ctx context.Context, contextID, name, slug string) (string, error) {
	url := fmt.Sprintf("%s/v1/contexts/%s/sources", c.cfg.MatcherURL, contextID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var list SourceListResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err == nil {
			for _, item := range list.Items {
				if item.Slug == slug || item.Name == name {
					return item.ID, nil
				}
			}
		}
	}

	createPayload := map[string]string{
		"name": name,
		"slug": slug,
		"type": "INTERNAL",
	}
	body, _ := json.Marshal(createPayload)

	reqCreate, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	c.setHeaders(reqCreate)

	respCreate, err := c.httpClient.Do(reqCreate)
	if err != nil {
		return "", err
	}
	defer respCreate.Body.Close()

	if respCreate.StatusCode != http.StatusOK && respCreate.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(respCreate.Body)
		return "", fmt.Errorf("failed creating source %s (HTTP %d): %s", name, respCreate.StatusCode, string(respBody))
	}

	var created SourceResponse
	if err := json.NewDecoder(respCreate.Body).Decode(&created); err != nil {
		return "", err
	}
	return created.ID, nil
}

// UploadExpectedReceivables posts pending installment lines (Source A) to Matcher 6.
func (c *MatcherClient) UploadExpectedReceivables(ctx context.Context, contextID, sourceAID string, records []domain.MatcherSourceRecord) error {
	if len(records) == 0 {
		return nil
	}

	url := fmt.Sprintf("%s/v1/imports/contexts/%s/sources/%s/upload", c.cfg.MatcherURL, contextID, sourceAID)
	payloadBytes, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("matcher: failed marshaling records: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("matcher: failed creating upload request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("matcher: upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("matcher: upload rejected (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

type RunResponse struct {
	RunID  string `json:"runId"`
	Status string `json:"status"`
}

// TriggerReconciliationRun instructs Matcher 6 to execute a matching pass.
func (c *MatcherClient) TriggerReconciliationRun(ctx context.Context, contextID string) (string, error) {
	url := fmt.Sprintf("%s/v1/contexts/%s/reconciliations/runs", c.cfg.MatcherURL, contextID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		return "", fmt.Errorf("matcher: failed creating trigger run request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("matcher: trigger run request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("matcher: trigger run rejected (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var runResp RunResponse
	if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
		return "", fmt.Errorf("matcher: failed decoding run response: %w", err)
	}

	return runResp.RunID, nil
}

func (c *MatcherClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
}
