package sumup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
)

var (
	ErrMissingAPIKey = errors.New("sumup: api_key is required in credentials")
)

type Adapter struct {
	baseURL    string
	httpClient *http.Client
}

func NewAdapter(baseURL string, httpClient *http.Client) *Adapter {
	if baseURL == "" {
		baseURL = "https://api.sumup.com"
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Adapter{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (a *Adapter) Type() domain.AcquirerType {
	return domain.AcquirerSumUp
}

// ValidateCredentials checks API connectivity by querying the merchant profile.
func (a *Adapter) ValidateCredentials(ctx context.Context, creds *domain.ConnectionCredentials) error {
	if creds.APIKey == "" {
		return ErrMissingAPIKey
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/v0.1/me", nil)
	if err != nil {
		return fmt.Errorf("sumup: failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+creds.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sumup: validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("sumup: invalid credentials (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sumup: unexpected validation status code: %d", resp.StatusCode)
	}

	return nil
}

// FetchSales retrieves historical card sales from SumUp.
func (a *Adapter) FetchSales(ctx context.Context, creds *domain.ConnectionCredentials, filter domain.SalesFilter) ([]domain.Sale, error) {
	if creds.APIKey == "" {
		return nil, ErrMissingAPIKey
	}

	endpoint := fmt.Sprintf("%s/v0.1/me/transactions/history?order=desc&limit=%d", a.baseURL, filter.Limit)
	if !filter.StartDate.IsZero() {
		endpoint += "&start_date=" + filter.StartDate.Format("2006-01-02")
	}
	if !filter.EndDate.IsZero() {
		endpoint += "&end_date=" + filter.EndDate.Format("2006-01-02")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("sumup: failed to create fetch sales request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+creds.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sumup: fetch sales request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sumup: fetch sales returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var hist SumUpHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&hist); err != nil {
		return nil, fmt.Errorf("sumup: failed to decode sales response: %w", err)
	}

	sales := make([]domain.Sale, 0, len(hist.Items))
	for _, item := range hist.Items {
		if item.Status != "SUCCESSFUL" {
			continue
		}
		sale, err := NormalizeSale(&item, "", "")
		if err != nil {
			continue
		}
		sales = append(sales, *sale)
	}

	return sales, nil
}

// FetchPayoutSchedule retrieves scheduled bank payouts from SumUp.
func (a *Adapter) FetchPayoutSchedule(ctx context.Context, creds *domain.ConnectionCredentials, filter domain.PayoutFilter) ([]domain.PayoutSchedule, error) {
	if creds.APIKey == "" {
		return nil, ErrMissingAPIKey
	}

	endpoint := fmt.Sprintf("%s/v0.1/me/payouts?limit=50", a.baseURL)
	if !filter.StartDate.IsZero() {
		endpoint += "&start_date=" + filter.StartDate.Format("2006-01-02")
	}
	if !filter.EndDate.IsZero() {
		endpoint += "&end_date=" + filter.EndDate.Format("2006-01-02")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("sumup: failed to create payout request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+creds.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sumup: payout request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sumup: payout returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var rawItems []SumUpPayoutItem
	if err := json.NewDecoder(resp.Body).Decode(&rawItems); err != nil {
		return nil, fmt.Errorf("sumup: failed to decode payout response: %w", err)
	}

	payouts := make([]domain.PayoutSchedule, 0, len(rawItems))
	for _, item := range rawItems {
		pDate, err := time.Parse("2006-01-02", item.Date)
		if err != nil {
			pDate = time.Now().UTC()
		}

		payouts = append(payouts, domain.PayoutSchedule{
			Acquirer:               domain.AcquirerSumUp,
			PayoutID:               item.ID,
			ExpectedSettlementDate: pDate,
			GrossAmount:            decimal.NewFromFloat(item.Amount),
			FeeAmount:              decimal.NewFromFloat(item.Fee),
			NetAmount:              decimal.NewFromFloat(item.NetAmount),
			AssociatedNSUs:         item.TransactionIDs,
		})
	}

	return payouts, nil
}

// ParseWebhook processes inbound SumUp webhook events.
func (a *Adapter) ParseWebhook(ctx context.Context, creds *domain.ConnectionCredentials, headers map[string]string, body []byte) (*domain.WebhookEvent, error) {
	var tx SumUpTransaction
	if err := json.Unmarshal(body, &tx); err != nil {
		return nil, fmt.Errorf("sumup: invalid webhook JSON body: %w", err)
	}

	sale, err := NormalizeSale(&tx, "", "")
	if err != nil {
		return nil, fmt.Errorf("sumup: failed to parse webhook sale: %w", err)
	}

	return &domain.WebhookEvent{
		Type:     "SALE_CREATED",
		RawEvent: string(body),
		Sale:     sale,
	}, nil
}
