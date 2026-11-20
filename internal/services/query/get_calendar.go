package query

import (
	"context"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
)

type GetCalendarInput struct {
	ReceivablesAccountID string
	Acquirer             *domain.AcquirerType
	Status               *domain.InstallmentStatus
	StartDate            *time.Time
	EndDate              *time.Time
	Limit                int
	Offset               int
}

type GetCalendarHandler struct {
	repo domain.IReceivablesRepository
}

func NewGetCalendarHandler(repo domain.IReceivablesRepository) *GetCalendarHandler {
	return &GetCalendarHandler{repo: repo}
}

func (h *GetCalendarHandler) Execute(ctx context.Context, input GetCalendarInput) ([]domain.Installment, error) {
	filter := domain.CalendarFilter{
		ReceivablesAccountID: input.ReceivablesAccountID,
		Acquirer:             input.Acquirer,
		Status:               input.Status,
		StartDate:            input.StartDate,
		EndDate:              input.EndDate,
		Limit:                input.Limit,
		Offset:               input.Offset,
	}

	return h.repo.GetReceivablesCalendar(ctx, filter)
}
