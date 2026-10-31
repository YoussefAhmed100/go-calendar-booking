package controller

import (
	"errors"
	"log"
	"net/http"

	"calendar-booking/internal/common/response"
	"calendar-booking/internal/modules/appointment/service"

	"github.com/gin-gonic/gin"
)

// errorResponses maps domain errors to HTTP responses.
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

// handleError maps domain errors to HTTP responses.
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