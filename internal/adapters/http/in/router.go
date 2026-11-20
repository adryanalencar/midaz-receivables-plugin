package in

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// NewRouter registers all HTTP routes for the Receivables Plugin.
func NewRouter(
	manifestHandler *ManifestHandler,
	accountHandler *AccountHandler,
	connectionHandler *ConnectionHandler,
	matcherWebhookHandler *MatcherWebhookHandler,
	acquirerWebhookHandler *AcquirerWebhookHandler,
) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "midaz-plugin-receivables",
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Webhook-Secret",
	}))

	// Health and readiness probes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "healthy",
			"app":    "midaz-plugin-receivables",
		})
	})
	app.Get("/ready", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ready",
		})
	})

	// Midaz Console Manifest
	app.Get("/api/manifest", manifestHandler.GetManifest)

	// API v1 Group
	v1 := app.Group("/v1")

	// Accounts endpoints
	accounts := v1.Group("/accounts/receivables")
	accounts.Post("/", accountHandler.CreateAccount)
	accounts.Get("/", accountHandler.ListAccounts)
	accounts.Get("/:id", accountHandler.GetAccount)
	accounts.Post("/:id/connections", connectionHandler.CreateConnection)
	accounts.Get("/:id/connections", connectionHandler.ListConnections)
	accounts.Get("/:id/calendar", accountHandler.GetCalendar)

	// Connections endpoints
	connections := v1.Group("/connections")
	connections.Post("/:id/sync", connectionHandler.SyncConnection)

	// Webhook endpoints
	webhooks := v1.Group("/webhooks")
	if matcherWebhookHandler != nil {
		webhooks.Post("/matcher", matcherWebhookHandler.HandleMatcherWebhook)
	}
	if acquirerWebhookHandler != nil {
		webhooks.Post("/sumup/:connection_id", acquirerWebhookHandler.HandleSumUpWebhook)
		webhooks.Post("/pagbank/:connection_id", acquirerWebhookHandler.HandlePagBankWebhook)
	}

	return app
}
