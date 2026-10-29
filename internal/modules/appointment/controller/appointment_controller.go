package controller

import (
	"errors"
	"log"
	"net/http"
	"strings"

	commonmiddleware "calendar-booking/internal/common/middleware"
	"calendar-booking/internal/common/response"
	"calendar-booking/internal/modules/appointment/dto"
	"calendar-booking/internal/modules/appointment/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	idempotencyKeyHeader    = "Idempotency-Key"
	maxIdempotencyKeyLength = 255
)

// errorResponses maps domain errors to HTTP responses. The first match wins.
var errorResponses = []struct {
	err     error
	status  int
	message string
}{
	{service.ErrInvalidTimezone, http.StatusBadRequest, "invalid timezone, use an IANA name like Africa/Cairo"},
	{service.ErrInvalidDate, http.StatusBadRequest, "invalid date, expected YYYY-MM-DD"},
	{service.ErrInvalidDuration, http.StatusUnprocessableEntity, "duration must be a multiple of the slot length and at most 8 hours"},
	{service.ErrStartInPast, http.StatusUnprocessableEntity, "appointment start is in the past"},
	{service.ErrOutsideWorkingHours, http.StatusUnprocessableEntity, "appointment is outside working hours or not aligned to a slot"},
	{service.ErrIdempotencyConflict, http.StatusUnprocessableEntity, "idempotency key was already used with different data"},
	{service.ErrAppointmentNotFound, http.StatusNotFound, "appointment not found"},
	{service.ErrSlotUnavailable, http.StatusConflict, "the requested time slot is not available"},
	{service.ErrNoCalendarSelected, http.StatusConflict, "no calendar selected, choose one first"},
	{service.ErrCalendarNotConnected, http.StatusConflict, "google account not connected, connect it first"},
	{service.ErrCalendarAuthorizationRevoked, http.StatusConflict, "google authorization expired or revoked, reconnect your account"},
	{service.ErrCalendarSyncFailed, http.StatusBadGateway, "failed to sync with Google Calendar, retry the same request"},
}

// AppointmentController handles the appointments HTTP endpoints.
type AppointmentController struct {
	appointments service.AppointmentService
}

// NewAppointmentController creates the controller.
func NewAppointmentController(appointments service.AppointmentService) *AppointmentController {
	return &AppointmentController{appointments: appointments}
}

// Slots lists the free slots of a day.
func (h *AppointmentController) Slots(c *gin.Context) {
	userID, ok := commonmiddleware.AuthenticatedUserID(c)
	if !ok {
		return
	}

	var query dto.SlotsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BindError(c, err)
		return
	}

	slots, err := h.appointments.Slots(c.Request.Context(), userID, query.Date, query.DurationMinutes)
	if err != nil {
		handleError(c, "slots", err)
		return
	}

	response.JSON(c, http.StatusOK, dto.NewSlotResponses(slots))
}

// Create books a slot. A replay of an earlier Idempotency-Key returns 200 instead of 201.
func (h *AppointmentController) Create(c *gin.Context) {
	userID, ok := commonmiddleware.AuthenticatedUserID(c)
	if !ok {
		return
	}

	key := strings.TrimSpace(c.GetHeader(idempotencyKeyHeader))
	if key == "" || len(key) > maxIdempotencyKeyLength {
		response.Error(c, http.StatusBadRequest, "Idempotency-Key header is required (max 255 characters)")
		return
	}

	var request dto.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BindError(c, err)
		return
	}

	appointment, created, err := h.appointments.Create(c.Request.Context(), service.CreateInput{
		UserID:          userID,
		IdempotencyKey:  key,
		Title:           request.Title,
		CustomerName:    request.CustomerName,
		CustomerEmail:   request.CustomerEmail,
		StartAt:         request.StartAt,
		DurationMinutes: request.DurationMinutes,
		Timezone:        request.Timezone,
	})
	if err != nil {
		handleError(c, "create", err)
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}

	response.JSON(c, status, dto.NewAppointmentResponse(appointment))
}

// List returns the user's appointments.
func (h *AppointmentController) List(c *gin.Context) {
	userID, ok := commonmiddleware.AuthenticatedUserID(c)
	if !ok {
		return
	}

	var query dto.ListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BindError(c, err)
		return
	}

	appointments, err := h.appointments.List(c.Request.Context(), userID, service.ListFilter{
		From:  query.From,
		To:    query.To,
		Limit: query.Limit,
	})
	if err != nil {
		handleError(c, "list", err)
		return
	}

	response.JSON(c, http.StatusOK, dto.NewAppointmentResponses(appointments))
}

// Get returns one appointment owned by the user.
func (h *AppointmentController) Get(c *gin.Context) {
	userID, id, ok := userAndAppointmentID(c)
	if !ok {
		return
	}

	appointment, err := h.appointments.Get(c.Request.Context(), userID, id)
	if err != nil {
		handleError(c, "get", err)
		return
	}

	response.JSON(c, http.StatusOK, dto.NewAppointmentResponse(appointment))
}

// Cancel cancels the appointment and removes its Google event.
func (h *AppointmentController) Cancel(c *gin.Context) {
	userID, id, ok := userAndAppointmentID(c)
	if !ok {
		return
	}

	appointment, err := h.appointments.Cancel(c.Request.Context(), userID, id)
	if err != nil {
		handleError(c, "cancel", err)
		return
	}

	response.JSON(c, http.StatusOK, dto.NewAppointmentResponse(appointment))
}

// userAndAppointmentID reads the authenticated user and the :id path parameter.
// On failure it writes the response itself and returns false.
func userAndAppointmentID(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	userID, ok := commonmiddleware.AuthenticatedUserID(c)
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid appointment id")
		return uuid.Nil, uuid.Nil, false
	}

	return userID, id, true
}

// handleError maps domain errors to HTTP responses.
// Server-side failures are logged and never exposed to the client.
func handleError(c *gin.Context, operation string, err error) {
	for _, mapped := range errorResponses {
		if errors.Is(err, mapped.err) {
			if mapped.status >= http.StatusInternalServerError {
				log.Printf("appointments %s: %v", operation, err)
			}
			response.Error(c, mapped.status, mapped.message)
			return
		}
	}

	log.Printf("appointments %s: %v", operation, err)
	response.Error(c, http.StatusInternalServerError, "failed to process appointment request")
}
