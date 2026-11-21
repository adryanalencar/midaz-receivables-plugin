package bootstrap

import (
	"fmt"
	"strings"
)

const ApplicationName = "midaz-plugin-receivables"

// Config encapsulates all environment configurations for the Receivables Plugin.
type Config struct {
	EnvName       string `env:"ENV_NAME" default:"development"`
	LogLevel      string `env:"LOG_LEVEL" default:"info"`
	ServerAddress string `env:"SERVER_ADDRESS" default:"0.0.0.0:8082"`
	PluginHost    string `env:"PLUGIN_HOST" default:"http://localhost:8082"`

	// Cryptographic master key for credentials encryption at rest (32 bytes or 64 hex characters)
	CredentialsEncryptionKey string `env:"CREDENTIALS_ENCRYPTION_KEY"`

	// PostgreSQL Database Settings (DATABASE_URL connection string takes precedence over discrete DB_* parameters)
	DatabaseURL   string `env:"DATABASE_URL" default:""`
	DBURI         string `env:"DB_URI" default:""`
	DBHost        string `env:"DB_HOST" default:"localhost"`
	DBPort        string `env:"DB_PORT" default:"5432"`
	DBUser        string `env:"DB_USER" default:"midaz"`
	DBPassword    string `env:"DB_PASSWORD" default:"midaz"`
	DBName        string `env:"DB_NAME" default:"plugin_receivables"`
	DBSSLMode     string `env:"DB_SSLMODE" default:"disable"`
	DBAutoMigrate bool   `env:"DB_AUTO_MIGRATE" default:"true"`

	// Midaz Ledger Core Settings
	MidazOnboardingURL  string `env:"MIDAZ_ONBOARDING_URL" default:"http://localhost:3000"`
	MidazTransactionURL string `env:"MIDAZ_TRANSACTION_URL" default:"http://localhost:3001"`
	MidazOrganizationID string `env:"MIDAZ_ORGANIZATION_ID"`
	MidazLedgerID       string `env:"MIDAZ_LEDGER_ID"`
	MidazClearingAccountAlias string `env:"MIDAZ_CLEARING_ACCOUNT_ALIAS" default:"@external/BRL:universal"`
	MidazAuthToken      string `env:"MIDAZ_AUTH_TOKEN" default:""`

	// Matcher 6 Reconciliation Settings (:4018)
	MatcherURL              string `env:"MATCHER_URL" default:"http://localhost:4018"`
	MatcherContextSlug      string `env:"MATCHER_CONTEXT_SLUG" default:"acquirers-settlement"`
	MatcherWebhookSecret    string `env:"MATCHER_WEBHOOK_SECRET" default:"secret_matcher_local_123"`
	MatcherUploadWindowDays int    `env:"MATCHER_UPLOAD_WINDOW_DAYS" default:"7"`

	// Inter-Plugin Integration (Open Finance Plugin :8085)
	OpenFinancePluginURL string `env:"OPENFINANCE_PLUGIN_URL" default:"http://localhost:8085"`

	// Worker Schedule Settings
	SyncIntervalMinutes int `env:"SYNC_INTERVAL_MINUTES" default:"60"`
	WorkerBatchLimit    int `env:"WORKER_BATCH_LIMIT" default:"500"`
}

// GetDatabaseURI resolves the active PostgreSQL connection string.
func (c *Config) GetDatabaseURI() string {
	if strings.TrimSpace(c.DatabaseURL) != "" {
		return c.DatabaseURL
	}
	if strings.TrimSpace(c.DBURI) != "" {
		return c.DBURI
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

// Validate ensures all mandatory configurations are present and valid.
func (c *Config) Validate() error {
	if len(c.CredentialsEncryptionKey) != 32 && len(c.CredentialsEncryptionKey) != 64 {
		return fmt.Errorf("config: CREDENTIALS_ENCRYPTION_KEY must be 32 raw bytes or 64 hex characters (received %d bytes)", len(c.CredentialsEncryptionKey))
	}
	if strings.TrimSpace(c.MidazOrganizationID) == "" {
		return fmt.Errorf("config: MIDAZ_ORGANIZATION_ID is required")
	}
	if strings.TrimSpace(c.MidazLedgerID) == "" {
		return fmt.Errorf("config: MIDAZ_LEDGER_ID is required")
	}
	return nil
}
