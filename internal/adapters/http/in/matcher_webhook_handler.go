package in

import (
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/services/command"
	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
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

	if h.settleMatchedHandler == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "acknowledged",
		})
	}

	// Group matches by SourceBRecordID (the bank statement deposit line) to support N:1 batch settlements
	groups := make(map[string][]command.MatchedReceivableItem)
	groupDates := make(map[string]time.Time)
	groupAmounts := make(map[string]decimal.Decimal)

	for _, match := range payload.Matches {
		bankTxID := match.SourceBRecordID
		if bankTxID == "" {
			bankTxID = match.MatchID
		}

		groups[bankTxID] = append(groups[bankTxID], command.MatchedReceivableItem{
			SourceARecordID: match.SourceARecordID,
			Amount:          match.SettlementAmount,
		})

		if !match.SettlementDate.IsZero() {
			groupDates[bankTxID] = match.SettlementDate
		}

		currTotal := groupAmounts[bankTxID]
		groupAmounts[bankTxID] = currTotal.Add(match.SettlementAmount)
	}

	settledGroups := 0
	for bankTxID, items := range groups {
		settleDate := groupDates[bankTxID]
		if settleDate.IsZero() {
			settleDate = time.Now().UTC()
		}

		input := command.SettleMatchedReceivableInput{
			OrganizationID:     h.cfg.MidazOrganizationID,
			LedgerID:           h.cfg.MidazLedgerID,
			MatchID:            payload.RunID,
			MatcherRunID:       payload.RunID,
			BankTransactionID:  bankTxID,
			TotalDepositAmount: groupAmounts[bankTxID],
			SettlementDate:     settleDate,
			Items:              items,
		}

		if err := h.settleMatchedHandler.Execute(c.Context(), input); err != nil {
			continue
		}
		settledGroups++
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":          "processed",
		"run_id":          payload.RunID,
		"matches_total":   len(payload.Matches),
		"deposits_settled": settledGroups,
	})
}
