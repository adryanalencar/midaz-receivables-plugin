package in

import (
	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
	"github.com/gofiber/fiber/v2"
)

type ManifestHandler struct {
	cfg *bootstrap.Config
}

func NewManifestHandler(cfg *bootstrap.Config) *ManifestHandler {
	return &ManifestHandler{cfg: cfg}
}

// GetManifest returns the Midaz Console Microfrontend plugin manifest.
func (h *ManifestHandler) GetManifest(c *fiber.Ctx) error {
	manifest := fiber.Map{
		"name":        "midaz-plugin-receivables",
		"version":     "1.0.0",
		"displayName": "Acquirer Receivables & Settlement",
		"description": "Multi-tenant card acquirers agenda orchestration and automated three-way bank settlement reconciliation with Matcher 6",
		"category":    "reconciliation",
		"icon":        "CreditCard",
		"routes": []fiber.Map{
			{
				"path":         "/receivables",
				"label":        "Receivables Agenda",
				"componentUrl": "/remoteEntry.js",
				"scope":        "midazReceivables",
				"module":       "./ReceivablesDashboard",
			},
			{
				"path":         "/receivables/connections",
				"label":        "Acquirer Connections",
				"componentUrl": "/remoteEntry.js",
				"scope":        "midazReceivables",
				"module":       "./ConnectionsManager",
			},
			{
				"path":         "/receivables/reconciliation",
				"label":        "Matcher Reconciliation",
				"componentUrl": "/remoteEntry.js",
				"scope":        "midazReceivables",
				"module":       "./ReconciliationView",
			},
		},
		"configuration": fiber.Map{
			"supportedAcquirers":   []string{"SUMUP", "PAGBANK", "ADITUM"},
			"matcherEngineVersion": "6.0.0-beta.1",
			"pluginHost":           h.cfg.PluginHost,
		},
	}

	return c.Status(fiber.StatusOK).JSON(manifest)
}
