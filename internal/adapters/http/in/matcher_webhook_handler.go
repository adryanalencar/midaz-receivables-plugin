package in

import (
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/services/command"
	"github.com/gofiber/fiber/v2"
)

type MatcherWebhookHandler struct {
	cfg                  *bootstrap.Config
	settleMatchedHandler *command.SettleMatchedReceivableHandler
}

func NewMatcherWebhookHandler(
	cfg *bootstrap.Config,
	settleMatchedHandler *command.SettleMatchedReceivableHandler,
) *MatcherWebhookHandler {
	return &MatcherWebhookHandler{
		cfg:                  cfg,
		settleMatchedHandler: settleMatchedHandler,
	}
}

// HandleMatcherWebhook processes the reconciliation completion webhook from Matcher 6.
func (h *MatcherWebhookHandler) HandleMatcherWebhook(c *fiber.Ctx) error {
	var payload domain.MatcherWebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid matcher webhook JSON payload",
		})
	}

	settledCount := 0
	for _, match := range payload.Matches {
		settleDate := match.SettlementDate
		if settleDate.IsZero() {
			settleDate = time.Now().UTC()
		}

		input := command.SettleMatchedReceivableInput{
			OrganizationID:  h.cfg.MidazOrganizationID,
			LedgerID:        h.cfg.MidazLedgerID,
			MatchID:         match.MatchID,
			SourceARecordID: match.SourceARecordID,
			SourceBRecordID: match.SourceBRecordID,
			SettledAmount:   match.SettlementAmount,
			SettlementDate:  settleDate,
			MatcherRunID:    payload.RunID,
		}

		if err := h.settleMatchedHandler.Execute(c.Context(), input); err != nil {
			// Log error and continue with remaining matches
			continue
		}
		settledCount++
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":          "processed",
		"run_id":          payload.RunID,
		"matches_total":   len(payload.Matches),
		"matches_settled": settledCount,
	})
}
