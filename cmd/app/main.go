package main

import (
	"context"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libZap "github.com/LerianStudio/lib-commons/v2/commons/zap"
	httpIn "github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/http/in"
	httpOut "github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/http/out"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/postgres"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/providers"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/providers/pagbank"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/providers/sumup"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/crypto"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/services/command"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/services/query"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/workers"
)

func main() {
	// 1. Initialize local .env if present
	libCommons.InitLocalEnvConfig()

	// 2. Load configuration from environment variables
	cfg := &bootstrap.Config{}
	if err := libCommons.SetConfigFromEnvVars(cfg); err != nil {
		panic(err)
	}

	// 3. Structured Logging (Zap)
	logger := libZap.InitializeLogger()

	// 4. Validate Configuration
	if err := cfg.Validate(); err != nil {
		logger.Warnf("Configuration validation warning: %v", err)
	}

	// 5. Initialize PostgreSQL Database, Run Migrations & Setup Repository
	var repo *postgres.Repository
	db, err := postgres.ConnectPostgres(cfg.GetDatabaseURI())
	if err != nil {
		logger.Warnf("PostgreSQL database connection deferred: %v", err)
	} else {
		logger.Infof("Connected to PostgreSQL database at %s:%s (database: %s)", cfg.DBHost, cfg.DBPort, cfg.DBName)

		if cfg.DBAutoMigrate {
			logger.Infof("Executing database migration governance check...")
			migrator := postgres.NewMigrator(db, logger)
			if err := migrator.MigrateUp(context.Background()); err != nil {
				logger.Fatalf("Database migration failed: %v", err)
			}
		}

		repo = postgres.NewRepository(db)
	}

	// 6. Cryptographic Layer: AES-256-GCM AEAD Encryptor
	var encryptor crypto.Encryptor
	if cfg.CredentialsEncryptionKey != "" {
		enc, err := crypto.NewGCMEncryptor(cfg.CredentialsEncryptionKey)
		if err != nil {
			logger.Fatalf("Failed to initialize credentials encryptor: %v", err)
		}
		encryptor = enc
		logger.Infof("AES-256-GCM credentials encryption initialized")
	}

	// 7. Acquirer Providers Factory (SumUp, PagBank)
	providerFactory := providers.NewFactory()
	providerFactory.Register(sumup.NewAdapter("", nil))
	providerFactory.Register(pagbank.NewAdapter("", nil))
	logger.Infof("Registered acquirer provider adapters: SUMUP, PAGBANK")

	// 8. Midaz Ledger Core Client & Clearing Account Verification
	midazClient := httpOut.NewMidazClient(cfg, nil)
	if cfg.MidazOrganizationID != "" && cfg.MidazLedgerID != "" {
		ensureClearingCmd := command.NewEnsureAcquirerAccountsHandler(midazClient)
		if err := ensureClearingCmd.Execute(context.Background(), cfg.MidazOrganizationID, cfg.MidazLedgerID); err != nil {
			logger.Warnf("Universal clearing accounts verification deferred: %v", err)
		} else {
			logger.Infof("Verified universal acquirer clearing accounts in Midaz")
		}
	}

	// 9. Matcher 6 Reconciliation Client & Open Finance Internal Client
	matcherClient := httpOut.NewMatcherClient(cfg, nil)
	openFinanceClient := httpOut.NewOpenFinanceClient(cfg, nil)

	// 10. CQRS Command & Query Handlers
	var createAccountHandler *command.CreateReceivablesAccountHandler
	var createConnectionHandler *command.CreateAcquirerConnectionHandler
	var ingestSaleHandler *command.IngestAcquirerSaleHandler
	var settleMatchedHandler *command.SettleMatchedReceivableHandler
	var uploadMatcherHandler *command.UploadReceivablesToMatcherHandler
	var calendarHandler *query.GetCalendarHandler
	var matcherWorker *workers.MatcherUploadWorker

	if repo != nil {
		createAccountHandler = command.NewCreateReceivablesAccountHandler(repo, midazClient, matcherClient)
		ingestSaleHandler = command.NewIngestAcquirerSaleHandler(repo, midazClient)
		settleMatchedHandler = command.NewSettleMatchedReceivableHandler(repo, midazClient).WithOpenFinanceNotifier(openFinanceClient)
		uploadMatcherHandler = command.NewUploadReceivablesToMatcherHandler(repo, matcherClient)
		calendarHandler = query.NewGetCalendarHandler(repo)

		if encryptor != nil {
			createConnectionHandler = command.NewCreateAcquirerConnectionHandler(repo, providerFactory, encryptor)
		}

		matcherWorker = workers.NewMatcherUploadWorker(cfg, uploadMatcherHandler)
	}

	// 11. Inbound HTTP Adapters & Handlers
	manifestHandler := httpIn.NewManifestHandler(cfg)
	accountHandler := httpIn.NewAccountHandler(cfg, repo, createAccountHandler, calendarHandler)
	connectionHandler := httpIn.NewConnectionHandler(cfg, repo, createConnectionHandler, ingestSaleHandler, providerFactory, encryptor)
	matcherWebhookHandler := httpIn.NewMatcherWebhookHandler(cfg, settleMatchedHandler)
	acquirerWebhookHandler := httpIn.NewAcquirerWebhookHandler(cfg, repo, ingestSaleHandler, providerFactory, encryptor)

	router := httpIn.NewRouter(
		manifestHandler,
		accountHandler,
		connectionHandler,
		matcherWebhookHandler,
		acquirerWebhookHandler,
	)

	// 12. Run Service with Graceful Shutdown
	service := &bootstrap.Service{
		FiberApp: router,
		Config:   cfg,
		Logger:   logger,
		Worker:   matcherWorker,
	}

	service.Run()
}
