package workers

import (
	"context"
	"log"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/services/command"
)

// MatcherUploadWorker periodically uploads pending receivables to Matcher 6.
type MatcherUploadWorker struct {
	cfg           *bootstrap.Config
	uploadHandler *command.UploadReceivablesToMatcherHandler
	stopChan      chan struct{}
}

// NewMatcherUploadWorker creates a new worker instance.
func NewMatcherUploadWorker(
	cfg *bootstrap.Config,
	uploadHandler *command.UploadReceivablesToMatcherHandler,
) *MatcherUploadWorker {
	return &MatcherUploadWorker{
		cfg:           cfg,
		uploadHandler: uploadHandler,
		stopChan:      make(chan struct{}),
	}
}

// Start initiates the background upload worker loop.
func (w *MatcherUploadWorker) Start(ctx context.Context) {
	interval := time.Duration(w.cfg.SyncIntervalMinutes) * time.Minute
	if interval <= 0 {
		interval = 60 * time.Minute
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		// Execute immediately on startup
		w.runCycle(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-w.stopChan:
				return
			case <-ticker.C:
				w.runCycle(ctx)
			}
		}
	}()
}

// Stop terminates the worker loop.
func (w *MatcherUploadWorker) Stop() {
	close(w.stopChan)
}

func (w *MatcherUploadWorker) runCycle(ctx context.Context) {
	count, err := w.uploadHandler.Execute(ctx, command.UploadReceivablesToMatcherInput{
		WindowDays: w.cfg.MatcherUploadWindowDays,
		Limit:      w.cfg.WorkerBatchLimit,
	})
	if err != nil {
		log.Printf("[MatcherUploadWorker] cycle error: %v", err)
		return
	}
	if count > 0 {
		log.Printf("[MatcherUploadWorker] successfully uploaded %d pending receivables to Matcher 6", count)
	}
}
