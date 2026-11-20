package command

import (
	"context"
	"fmt"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
)

type UploadReceivablesToMatcherInput struct {
	WindowDays int
	Limit      int
}

type UploadReceivablesToMatcherHandler struct {
	repo          domain.IReceivablesRepository
	matcherClient domain.IMatcherClient
}

func NewUploadReceivablesToMatcherHandler(
	repo domain.IReceivablesRepository,
	matcherClient domain.IMatcherClient,
) *UploadReceivablesToMatcherHandler {
	return &UploadReceivablesToMatcherHandler{
		repo:          repo,
		matcherClient: matcherClient,
	}
}

func (h *UploadReceivablesToMatcherHandler) Execute(ctx context.Context, input UploadReceivablesToMatcherInput) (int, error) {
	windowDays := input.WindowDays
	if windowDays <= 0 {
		windowDays = 7
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 500
	}

	// 1. Ensure Context and Sources are ready in Matcher 6
	contextID, sourceAID, _, err := h.matcherClient.EnsureContextAndSources(ctx)
	if err != nil {
		return 0, fmt.Errorf("upload_receivables_to_matcher: failed provisioning matcher context: %w", err)
	}

	now := time.Now().UTC()
	windowStart := now.AddDate(0, 0, -3) // Allow small retroactive window
	windowEnd := now.AddDate(0, 0, windowDays)

	// 2. Fetch pending installments within the settlement window
	installments, err := h.repo.ListPendingInstallmentsForMatcher(ctx, windowStart, windowEnd, limit)
	if err != nil {
		return 0, fmt.Errorf("upload_receivables_to_matcher: failed querying pending installments: %w", err)
	}

	if len(installments) == 0 {
		return 0, nil
	}

	// 3. Transform installments into Matcher Source A records
	records := make([]domain.MatcherSourceRecord, len(installments))
	for i, inst := range installments {
		recordID := inst.DeterministicReceivableCode()
		records[i] = domain.MatcherSourceRecord{
			ID:           recordID,
			Date:         inst.ExpectedSettlementDate.Format("2006-01-02"),
			Amount:       inst.NetAmount,
			MerchantID:   inst.ReceivablesAccountID,
			Acquirer:     string(inst.Acquirer),
			ConnectionID: inst.ConnectionID,
		}
	}

	// 4. Upload batch to Matcher 6
	if err := h.matcherClient.UploadExpectedReceivables(ctx, contextID, sourceAID, records); err != nil {
		return 0, fmt.Errorf("upload_receivables_to_matcher: failed uploading to matcher: %w", err)
	}

	// 5. Update local state marking installments as uploaded to Matcher
	for _, inst := range installments {
		recordID := inst.DeterministicReceivableCode()
		if err := h.repo.MarkInstallmentUploadedToMatcher(ctx, inst.ID, recordID); err != nil {
			// Log error but continue updating remainder
			continue
		}
	}

	return len(installments), nil
}
