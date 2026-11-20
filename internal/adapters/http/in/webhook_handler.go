package in

import (
	"encoding/json"
	"fmt"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/providers"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/crypto"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/services/command"
	"github.com/gofiber/fiber/v2"
)

type AcquirerWebhookHandler struct {
	cfg               *bootstrap.Config
	repo              domain.IReceivablesRepository
	ingestSaleHandler *command.IngestAcquirerSaleHandler
	providerFactory   *providers.Factory
	encryptor         crypto.Encryptor
}

func NewAcquirerWebhookHandler(
	cfg *bootstrap.Config,
	repo domain.IReceivablesRepository,
	ingestSaleHandler *command.IngestAcquirerSaleHandler,
	providerFactory *providers.Factory,
	encryptor crypto.Encryptor,
) *AcquirerWebhookHandler {
	return &AcquirerWebhookHandler{
		cfg:               cfg,
		repo:              repo,
		ingestSaleHandler: ingestSaleHandler,
		providerFactory:   providerFactory,
		encryptor:         encryptor,
	}
}

// HandleSumUpWebhook processes incoming SumUp sale events.
func (h *AcquirerWebhookHandler) HandleSumUpWebhook(c *fiber.Ctx) error {
	return h.handleAcquirerWebhook(c, domain.AcquirerSumUp)
}

// HandlePagBankWebhook processes incoming PagBank charge notifications.
func (h *AcquirerWebhookHandler) HandlePagBankWebhook(c *fiber.Ctx) error {
	return h.handleAcquirerWebhook(c, domain.AcquirerPagBank)
}

func (h *AcquirerWebhookHandler) handleAcquirerWebhook(c *fiber.Ctx, acquirer domain.AcquirerType) error {
	connID := c.Params("connection_id")

	// 1. Fetch connection
	conn, err := h.repo.FindAcquirerConnectionByID(c.Context(), connID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if conn == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "connection not found",
		})
	}

	// 2. Decrypt credentials
	decryptedBytes, err := h.encryptor.Decrypt(conn.CredentialsEncrypted)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed decrypting credentials",
		})
	}

	var creds domain.ConnectionCredentials
	if err := json.Unmarshal(decryptedBytes, &creds); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed parsing decrypted credentials",
		})
	}

	// 3. Resolve provider adapter
	provider, err := h.providerFactory.Get(acquirer)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("unsupported acquirer: %s", acquirer),
		})
	}

	// Extract headers
	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(key, val []byte) {
		headers[string(key)] = string(val)
	})

	body := c.Body()

	// 4. Parse webhook through provider
	event, err := provider.ParseWebhook(c.Context(), &creds, headers, body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("failed parsing webhook: %s", err.Error()),
		})
	}

	// 5. If event is a sale, ingest and book agenda in Midaz
	if event != nil && event.Sale != nil {
		event.Sale.ConnectionID = conn.ID
		event.Sale.ReceivablesAccountID = conn.ReceivablesAccountID

		err := h.ingestSaleHandler.Execute(c.Context(), command.IngestAcquirerSaleInput{
			OrganizationID: h.cfg.MidazOrganizationID,
			LedgerID:       h.cfg.MidazLedgerID,
			Sale:           event.Sale,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("failed ingesting sale: %s", err.Error()),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "processed",
		"event":  event.Type,
	})
}
