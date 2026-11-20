package in

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/providers"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/crypto"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/services/command"
	"github.com/gofiber/fiber/v2"
)

type ConnectionHandler struct {
	cfg               *bootstrap.Config
	repo              domain.IReceivablesRepository
	createConnHandler *command.CreateAcquirerConnectionHandler
	ingestSaleHandler *command.IngestAcquirerSaleHandler
	providerFactory   *providers.Factory
	encryptor         crypto.Encryptor
}

func NewConnectionHandler(
	cfg *bootstrap.Config,
	repo domain.IReceivablesRepository,
	createConnHandler *command.CreateAcquirerConnectionHandler,
	ingestSaleHandler *command.IngestAcquirerSaleHandler,
	providerFactory *providers.Factory,
	encryptor crypto.Encryptor,
) *ConnectionHandler {
	return &ConnectionHandler{
		cfg:               cfg,
		repo:              repo,
		createConnHandler: createConnHandler,
		ingestSaleHandler: ingestSaleHandler,
		providerFactory:   providerFactory,
		encryptor:         encryptor,
	}
}

// CreateConnection registers and validates an Acquirer connection bound to a ReceivablesAccount.
func (h *ConnectionHandler) CreateConnection(c *fiber.Ctx) error {
	accountID := c.Params("id")

	var input command.CreateAcquirerConnectionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body JSON",
		})
	}
	input.ReceivablesAccountID = accountID

	conn, err := h.createConnHandler.Execute(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(conn)
}

// ListConnections lists all acquirer connections bound to a receivables account.
func (h *ConnectionHandler) ListConnections(c *fiber.Ctx) error {
	accountID := c.Params("id")
	connections, err := h.repo.ListConnectionsByAccountID(c.Context(), accountID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(connections)
}

// SyncConnection triggers an on-demand historical sales ingestion from the acquirer.
func (h *ConnectionHandler) SyncConnection(c *fiber.Ctx) error {
	connID := c.Params("id")

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

	// 2. Decrypt credentials in runtime
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
	provider, err := h.providerFactory.Get(conn.AcquirerType)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("unsupported acquirer: %s", conn.AcquirerType),
		})
	}

	// 4. Fetch sales from Acquirer
	filter := domain.SalesFilter{
		Limit: 100,
	}
	sales, err := provider.FetchSales(c.Context(), &creds, filter)
	if err != nil {
		errMsg := err.Error()
		_ = h.repo.UpdateConnectionSyncState(c.Context(), conn.ID, "ERROR", &errMsg, time.Now().UTC())
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": fmt.Sprintf("failed fetching sales from %s: %s", conn.AcquirerType, err.Error()),
		})
	}

	// 5. Ingest each sale and post D+0 entries
	ingestedCount := 0
	for i := range sales {
		sale := &sales[i]
		sale.ConnectionID = conn.ID
		sale.ReceivablesAccountID = conn.ReceivablesAccountID

		err := h.ingestSaleHandler.Execute(c.Context(), command.IngestAcquirerSaleInput{
			OrganizationID: h.cfg.MidazOrganizationID,
			LedgerID:       h.cfg.MidazLedgerID,
			Sale:           sale,
		})
		if err != nil {
			continue // Log and continue remainder
		}
		ingestedCount++
	}

	now := time.Now().UTC()
	_ = h.repo.UpdateConnectionSyncState(c.Context(), conn.ID, "SUCCESS", nil, now)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":         "success",
		"connection_id":  conn.ID,
		"sales_ingested": ingestedCount,
		"last_sync_at":   now.Format(time.RFC3339),
	})
}
