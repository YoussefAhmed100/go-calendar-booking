package controller

import (

	"net/http"
	"strings"

	commonmiddleware "calendar-booking/internal/common/middleware"
	"calendar-booking/internal/common/response"
	"calendar-booking/internal/modules/appointment/dto"
	"calendar-booking/internal/modules/appointment/service"

	"github.com/gin-gonic/gin"
	
)

const (
	idempotencyKeyHeader    = "Idempotency-Key"
	maxIdempotencyKeyLength = 255
)



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



