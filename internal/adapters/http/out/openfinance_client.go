package out

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
)

type IOpenFinanceClient interface {
	MarkTransactionProcessed(ctx context.Context, bankTxID, midazTxCode string) error
}

type OpenFinanceClient struct {
	cfg        *bootstrap.Config
	httpClient *http.Client
}

func NewOpenFinanceClient(cfg *bootstrap.Config, httpClient *http.Client) *OpenFinanceClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &OpenFinanceClient{
		cfg:        cfg,
		httpClient: httpClient,
	}
}

type MarkProcessedRequest struct {
	BankTransactionID    string `json:"bankTransactionId"`
	MidazTransactionCode string `json:"midazTransactionCode"`
	Status               string `json:"status"`
}

func (c *OpenFinanceClient) MarkTransactionProcessed(ctx context.Context, bankTxID, midazTxCode string) error {
	if c.cfg.OpenFinancePluginURL == "" {
		return nil
	}

	url := fmt.Sprintf("%s/v1/internal/transactions/mark-processed", c.cfg.OpenFinancePluginURL)
	payload := MarkProcessedRequest{
		BankTransactionID:    bankTxID,
		MidazTransactionCode: midazTxCode,
		Status:               "SETTLED_VIA_RECEIVABLES",
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("openfinance_client: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("openfinance_client: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("openfinance_client: mark-processed request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("openfinance_client: unexpected response status HTTP %d", resp.StatusCode)
	}

	return nil
}
