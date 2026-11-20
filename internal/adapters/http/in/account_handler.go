package in

import (
	"strings"
	"time"

	"github.com/LerianStudio/midaz-plugin-receivables/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/domain"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/services/command"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/services/query"
	"github.com/gofiber/fiber/v2"
)

type AccountHandler struct {
	cfg             *bootstrap.Config
	repo            domain.IReceivablesRepository
	createHandler   *command.CreateReceivablesAccountHandler
	calendarHandler *query.GetCalendarHandler
}

func NewAccountHandler(
	cfg *bootstrap.Config,
	repo domain.IReceivablesRepository,
	createHandler *command.CreateReceivablesAccountHandler,
	calendarHandler *query.GetCalendarHandler,
) *AccountHandler {
	return &AccountHandler{
		cfg:             cfg,
		repo:            repo,
		createHandler:   createHandler,
		calendarHandler: calendarHandler,
	}
}

// CreateAccount provisions a new Receivables wallet in Midaz and the local repository.
func (h *AccountHandler) CreateAccount(c *fiber.Ctx) error {
	var input command.CreateReceivablesAccountInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body JSON",
		})
	}

	if input.OrganizationID == "" {
		input.OrganizationID = h.cfg.MidazOrganizationID
	}
	if input.LedgerID == "" {
		input.LedgerID = h.cfg.MidazLedgerID
	}

	acc, err := h.createHandler.Execute(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(acc)
}

// ListAccounts lists all managed receivables accounts.
func (h *AccountHandler) ListAccounts(c *fiber.Ctx) error {
	accounts, err := h.repo.ListReceivablesAccounts(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(accounts)
}

// GetAccount retrieves a specific account by ID.
func (h *AccountHandler) GetAccount(c *fiber.Ctx) error {
	id := c.Params("id")
	acc, err := h.repo.FindReceivablesAccountByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if acc == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "receivables account not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(acc)
}

// GetCalendar retrieves the scheduled receivables calendar for an account.
func (h *AccountHandler) GetCalendar(c *fiber.Ctx) error {
	id := c.Params("id")

	var startDate, endDate *time.Time
	if startStr := c.Query("start_date"); startStr != "" {
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = &t
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		if t, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = &t
		}
	}

	var status *domain.InstallmentStatus
	if stStr := c.Query("status"); stStr != "" {
		st := domain.InstallmentStatus(strings.ToUpper(stStr))
		status = &st
	}

	var acquirer *domain.AcquirerType
	if acqStr := c.Query("acquirer"); acqStr != "" {
		acq := domain.AcquirerType(strings.ToUpper(acqStr))
		acquirer = &acq
	}

	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)

	installments, err := h.calendarHandler.Execute(c.Context(), query.GetCalendarInput{
		ReceivablesAccountID: id,
		Acquirer:             acquirer,
		Status:               status,
		StartDate:            startDate,
		EndDate:              endDate,
		Limit:                limit,
		Offset:               offset,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(installments)
}
