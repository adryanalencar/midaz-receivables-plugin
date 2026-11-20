package pagbank

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
	ErrMissingToken = errors.New("pagbank: token or api_key is required in credentials")
)

type Adapter struct {
	baseURL    string
	httpClient *http.Client
}

func NewAdapter(baseURL string, httpClient *http.Client) *Adapter {
	if baseURL == "" {
		baseURL = "https://api.pagseguro.com"
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
	return domain.AcquirerPagBank
}

func (a *Adapter) getToken(creds *domain.ConnectionCredentials) string {
	if creds.Token != "" {
		return creds.Token
	}
	return creds.APIKey
}

// ValidateCredentials verifies API token against the PagBank charges API.
func (a *Adapter) ValidateCredentials(ctx context.Context, creds *domain.ConnectionCredentials) error {
	token := a.getToken(creds)
	if token == "" {
		return ErrMissingToken
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/charges?limit=1", nil)
	if err != nil {
		return fmt.Errorf("pagbank: failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("pagbank: validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("pagbank: invalid credentials (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagbank: unexpected validation status code: %d", resp.StatusCode)
	}

	return nil
}

// FetchSales retrieves settled and authorized charges from PagBank.
func (a *Adapter) FetchSales(ctx context.Context, creds *domain.ConnectionCredentials, filter domain.SalesFilter) ([]domain.Sale, error) {
	token := a.getToken(creds)
	if token == "" {
		return nil, ErrMissingToken
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	endpoint := fmt.Sprintf("%s/charges?limit=%d", a.baseURL, limit)
	if !filter.StartDate.IsZero() {
		endpoint += "&begin_date=" + filter.StartDate.Format("2006-01-02")
	}
	if !filter.EndDate.IsZero() {
		endpoint += "&end_date=" + filter.EndDate.Format("2006-01-02")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("pagbank: failed to create fetch sales request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pagbank: fetch sales request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pagbank: fetch sales returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var chargesResp PagBankChargesResponse
	if err := json.NewDecoder(resp.Body).Decode(&chargesResp); err != nil {
		return nil, fmt.Errorf("pagbank: failed to decode charges response: %w", err)
	}

	sales := make([]domain.Sale, 0, len(chargesResp.Charges))
	for _, charge := range chargesResp.Charges {
		if charge.Status != "PAID" && charge.Status != "AUTHORIZED" {
			continue
		}
		sale, err := NormalizeSale(&charge, "", "")
		if err != nil {
			continue
		}
		sales = append(sales, *sale)
	}

	return sales, nil
}

// FetchPayoutSchedule retrieves scheduled transfers and settlements.
func (a *Adapter) FetchPayoutSchedule(ctx context.Context, creds *domain.ConnectionCredentials, filter domain.PayoutFilter) ([]domain.PayoutSchedule, error) {
	token := a.getToken(creds)
	if token == "" {
		return nil, ErrMissingToken
	}

	endpoint := fmt.Sprintf("%s/transfers?limit=50", a.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("pagbank: failed to create payout request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pagbank: payout request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pagbank: payout returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var rawItems []PagBankPayoutItem
	if err := json.NewDecoder(resp.Body).Decode(&rawItems); err != nil {
		return nil, fmt.Errorf("pagbank: failed to decode payout response: %w", err)
	}

	hundred := decimal.NewFromInt(100)
	payouts := make([]domain.PayoutSchedule, 0, len(rawItems))
	for _, item := range rawItems {
		pDate, err := time.Parse("2006-01-02", item.ScheduledDate)
		if err != nil {
			pDate = time.Now().UTC()
		}

		payouts = append(payouts, domain.PayoutSchedule{
			Acquirer:               domain.AcquirerPagBank,
			PayoutID:               item.ID,
			ExpectedSettlementDate: pDate,
			GrossAmount:            decimal.NewFromInt(item.Amount).DivRound(hundred, 2),
			FeeAmount:              decimal.NewFromInt(item.Fee).DivRound(hundred, 2),
			NetAmount:              decimal.NewFromInt(item.NetAmount).DivRound(hundred, 2),
			AssociatedNSUs:         item.ChargeIDs,
		})
	}

	return payouts, nil
}

// ParseWebhook processes inbound PagBank notification webhooks.
func (a *Adapter) ParseWebhook(ctx context.Context, creds *domain.ConnectionCredentials, headers map[string]string, body []byte) (*domain.WebhookEvent, error) {
	var charge PagBankCharge
	if err := json.Unmarshal(body, &charge); err != nil {
		return nil, fmt.Errorf("pagbank: invalid webhook JSON body: %w", err)
	}

	sale, err := NormalizeSale(&charge, "", "")
	if err != nil {
		return nil, fmt.Errorf("pagbank: failed to parse webhook sale: %w", err)
	}

	return &domain.WebhookEvent{
		Type:     "SALE_CREATED",
		RawEvent: string(body),
		Sale:     sale,
	}, nil
}
