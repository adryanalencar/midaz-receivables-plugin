package aditum

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/shopspring/decimal"
)

var (
	ErrMissingMerchantCode = errors.New("aditum: merchant_code or merchant_id is required in credentials")
	ErrMissingMerchantToken = errors.New("aditum: token or api_key (merchant_token) is required in credentials")
	ErrInvalidWebhookAuth   = errors.New("aditum: invalid webhook authorization header")
)

type Adapter struct {
	portalBaseURL  string
	gatewayBaseURL string
	httpClient     *http.Client
}

func NewAdapter(portalBaseURL, gatewayBaseURL string, httpClient *http.Client) *Adapter {
	if portalBaseURL == "" {
		portalBaseURL = "https://portal-api.aditum.com.br"
	}
	if gatewayBaseURL == "" {
		gatewayBaseURL = "https://payment.aditum.com.br"
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Adapter{
		portalBaseURL:  strings.TrimRight(portalBaseURL, "/"),
		gatewayBaseURL: strings.TrimRight(gatewayBaseURL, "/"),
		httpClient:     httpClient,
	}
}

func (a *Adapter) Type() domain.AcquirerType {
	return domain.AcquirerAditum
}

func (a *Adapter) getMerchantCode(creds *domain.ConnectionCredentials) string {
	if creds.MerchantCode != "" {
		return creds.MerchantCode
	}
	if creds.ClientID != "" {
		return creds.ClientID
	}
	if creds.AdditionalProperties != nil {
		if val, ok := creds.AdditionalProperties["merchant_id"].(string); ok && val != "" {
			return val
		}
		if val, ok := creds.AdditionalProperties["merchant_code"].(string); ok && val != "" {
			return val
		}
	}
	return ""
}

func (a *Adapter) getMerchantToken(creds *domain.ConnectionCredentials) string {
	if creds.Token != "" {
		return creds.Token
	}
	if creds.APIKey != "" {
		return creds.APIKey
	}
	if creds.ClientSecret != "" {
		return creds.ClientSecret
	}
	if creds.AdditionalProperties != nil {
		if val, ok := creds.AdditionalProperties["merchant_token"].(string); ok && val != "" {
			return val
		}
	}
	return ""
}

// ValidateCredentials checks API connectivity against the Aditum Charges API
func (a *Adapter) ValidateCredentials(ctx context.Context, creds *domain.ConnectionCredentials) error {
	merchantCode := a.getMerchantCode(creds)
	if merchantCode == "" {
		return ErrMissingMerchantCode
	}

	merchantToken := a.getMerchantToken(creds)
	if merchantToken == "" {
		return ErrMissingMerchantToken
	}

	reqURL := fmt.Sprintf("%s/v1/Charges?pageSize=1&currentPage=1", a.portalBaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("aditum: failed to create validation request: %w", err)
	}

	req.Header.Set("merchantId", merchantCode)
	req.Header.Set("Authorization", "Bearer "+merchantToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("aditum: validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("aditum: invalid credentials (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("aditum: unexpected validation status code: %d", resp.StatusCode)
	}

	return nil
}

// FetchSales retrieves historical charges from Aditum
func (a *Adapter) FetchSales(ctx context.Context, creds *domain.ConnectionCredentials, filter domain.SalesFilter) ([]domain.Sale, error) {
	merchantCode := a.getMerchantCode(creds)
	if merchantCode == "" {
		return nil, ErrMissingMerchantCode
	}

	merchantToken := a.getMerchantToken(creds)
	if merchantToken == "" {
		return nil, ErrMissingMerchantToken
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	endpoint := fmt.Sprintf("%s/v1/Charges?pageSize=%d&currentPage=1", a.portalBaseURL, limit)
	if !filter.StartDate.IsZero() {
		endpoint += "&startDate=" + filter.StartDate.Format("2006-01-02")
	}
	if !filter.EndDate.IsZero() {
		endpoint += "&endDate=" + filter.EndDate.Format("2006-01-02")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("aditum: failed to create fetch sales request: %w", err)
	}

	req.Header.Set("merchantId", merchantCode)
	req.Header.Set("Authorization", "Bearer "+merchantToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aditum: fetch sales request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("aditum: fetch sales returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var response AditumChargesListResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("aditum: failed to decode sales response: %w", err)
	}

	sales := make([]domain.Sale, 0, len(response.Charges))
	for i := range response.Charges {
		charge := &response.Charges[i]
		// Skip canceled or failed charges
		status := strings.ToLower(charge.ChargeStatus)
		if strings.Contains(status, "cancel") || strings.Contains(status, "denied") || strings.Contains(status, "error") {
			continue
		}

		sale, err := NormalizeSale(charge, "", "")
		if err != nil {
			continue
		}
		sales = append(sales, *sale)
	}

	return sales, nil
}

// FetchPayoutSchedule calculates scheduled settlements from Aditum charges
func (a *Adapter) FetchPayoutSchedule(ctx context.Context, creds *domain.ConnectionCredentials, filter domain.PayoutFilter) ([]domain.PayoutSchedule, error) {
	sales, err := a.FetchSales(ctx, creds, domain.SalesFilter{
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
		Limit:     100,
	})
	if err != nil {
		return nil, err
	}

	type dateGroup struct {
		gross decimal.Decimal
		fee   decimal.Decimal
		net   decimal.Decimal
		nsus  []string
		date  time.Time
	}

	groups := make(map[string]*dateGroup)
	for _, sale := range sales {
		for _, inst := range sale.Installments {
			dateKey := inst.ExpectedSettlementDate.Format("2006-01-02")
			g, exists := groups[dateKey]
			if !exists {
				g = &dateGroup{
					date: inst.ExpectedSettlementDate,
					nsus: make([]string, 0),
				}
				groups[dateKey] = g
			}
			g.gross = g.gross.Add(inst.GrossAmount)
			g.fee = g.fee.Add(inst.MDRFee)
			g.net = g.net.Add(inst.NetAmount)
			if inst.NSU != "" {
				g.nsus = append(g.nsus, inst.NSU)
			}
		}
	}

	payouts := make([]domain.PayoutSchedule, 0, len(groups))
	for dateKey, g := range groups {
		payouts = append(payouts, domain.PayoutSchedule{
			Acquirer:               domain.AcquirerAditum,
			PayoutID:               fmt.Sprintf("ADITUM-PAYOUT-%s", dateKey),
			ExpectedSettlementDate: g.date,
			GrossAmount:            g.gross,
			FeeAmount:              g.fee,
			NetAmount:              g.net,
			AssociatedNSUs:         g.nsus,
		})
	}

	return payouts, nil
}

// ParseWebhook processes inbound Aditum webhook callbacks
func (a *Adapter) ParseWebhook(ctx context.Context, creds *domain.ConnectionCredentials, headers map[string]string, body []byte) (*domain.WebhookEvent, error) {
	expectedToken := a.getMerchantToken(creds)
	webhookSecret := creds.WebhookSecret

	// If security token or secret is configured, validate Authorization header
	authHeader := ""
	for k, v := range headers {
		if strings.EqualFold(k, "Authorization") {
			authHeader = v
			break
		}
	}

	if expectedToken != "" || webhookSecret != "" {
		if authHeader != "" {
			valid := a.validateWebhookAuth(authHeader, expectedToken, webhookSecret)
			if !valid {
				return nil, ErrInvalidWebhookAuth
			}
		}
	}

	sale, err := NormalizeWebhook(body, "", "")
	if err != nil {
		return nil, fmt.Errorf("aditum: failed to parse webhook sale: %w", err)
	}

	return &domain.WebhookEvent{
		Type:     "SALE_CREATED",
		RawEvent: string(body),
		Sale:     sale,
	}, nil
}

func (a *Adapter) validateWebhookAuth(header, token, secret string) bool {
	raw := strings.TrimSpace(header)
	if strings.HasPrefix(strings.ToLower(raw), "basic ") {
		raw = strings.TrimSpace(raw[6:])
	} else if strings.HasPrefix(strings.ToLower(raw), "bearer ") {
		raw = strings.TrimSpace(raw[7:])
	}

	// Check against token
	if token != "" {
		encodedToken := base64.StdEncoding.EncodeToString([]byte(token))
		if subtle.ConstantTimeCompare([]byte(raw), []byte(token)) == 1 ||
			subtle.ConstantTimeCompare([]byte(raw), []byte(encodedToken)) == 1 {
			return true
		}
	}

	// Check against secret
	if secret != "" {
		encodedSecret := base64.StdEncoding.EncodeToString([]byte(secret))
		if subtle.ConstantTimeCompare([]byte(raw), []byte(secret)) == 1 ||
			subtle.ConstantTimeCompare([]byte(raw), []byte(encodedSecret)) == 1 {
			return true
		}
	}

	return false
}
