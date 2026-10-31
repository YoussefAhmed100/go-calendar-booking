package controller

import (
	"net/http"

	commonmiddleware "calendar-booking/internal/common/middleware"
	"calendar-booking/internal/common/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// userAndAppointmentID extracts the authenticated user ID and appointment ID.
// It writes an HTTP response if either value is invalid.
func userAndAppointmentID(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	userID, ok := commonmiddleware.AuthenticatedUserID(c)
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}

	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid appointment id")
		return uuid.Nil, uuid.Nil, false
	}

	return userID, appointmentID, true
}