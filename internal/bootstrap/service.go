package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	libLog "github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/gofiber/fiber/v2"
)

// BackgroundWorker represents a runnable background process.
type BackgroundWorker interface {
	Start(ctx context.Context)
	Stop()
}

// Service encapsulates the lifecycle of the Receivables Plugin.
type Service struct {
	FiberApp *fiber.App
	Config   *Config
	Logger   libLog.Logger
	Worker   BackgroundWorker
}

// Run starts the Fiber HTTP server, background worker, and handles graceful shutdown.
func (s *Service) Run() {
	s.Logger.Infof("Starting %s on %s", ApplicationName, s.Config.ServerAddress)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Start Background Upload Worker
	if s.Worker != nil {
		s.Worker.Start(ctx)
		s.Logger.Infof("Background Matcher upload worker started")
	}

	// 2. Graceful shutdown signal channel
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.FiberApp.Listen(s.Config.ServerAddress); err != nil {
			s.Logger.Errorf("Server error: %v", err)
		}
	}()

	<-shutdown
	s.Logger.Infof("Shutting down %s gracefully...", ApplicationName)

	if s.Worker != nil {
		s.Worker.Stop()
	}

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := s.FiberApp.ShutdownWithContext(shutdownCtx); err != nil {
		s.Logger.Errorf("Error during shutdown: %v", err)
	}

	fmt.Println("Receivables Plugin stopped cleanly.")
}
