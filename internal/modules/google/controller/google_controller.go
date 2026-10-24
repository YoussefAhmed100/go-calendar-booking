package controller

import (
	"errors"
	"log"
	"net/http"

	commonmiddleware "calendar-booking/internal/common/middleware"
	"calendar-booking/internal/modules/google/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GoogleController handles the Google integration HTTP endpoints.
type GoogleController struct {
	googleService    *service.GoogleService
	calendarService  *service.CalendarService
	selectionService *service.CalendarSelectionService
}

// NewGoogleController creates the controller with its services.
func NewGoogleController(
	googleService *service.GoogleService,
	calendarService *service.CalendarService,
	selectionService *service.CalendarSelectionService,
) *GoogleController {
	return &GoogleController{
		googleService:    googleService,
		calendarService:  calendarService,
		selectionService: selectionService,
	}
}

type selectCalendarRequest struct {
	CalendarID string `json:"calendar_id" binding:"required"`
}

// Connect returns the Google authorization URL for the authenticated user.
func (h *GoogleController) Connect(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	authURL, err := h.googleService.GenerateAuthURL(c.Request.Context(), userID)
	if err != nil {
		respondGoogleError(c, "connect", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"auth_url": authURL})
}

// Callback completes the OAuth flow. It is called by Google, so it has no Bearer token;
// the user is identified through the persisted OAuth state.
func (h *GoogleController) Callback(c *gin.Context) {
	if c.Query("error") != "" {
		respondError(c, http.StatusBadRequest, "Google authorization was denied or failed")
		return
	}

	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		respondError(c, http.StatusBadRequest, "missing OAuth code or state")
		return
	}

	if err := h.googleService.CompleteAuthorization(c.Request.Context(), code, state); err != nil {
		respondGoogleError(c, "callback", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Google Calendar connected successfully"})
}

// Calendars lists the calendars of the Google account linked to the authenticated user.
func (h *GoogleController) Calendars(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	calendars, err := h.calendarService.ListCalendars(c.Request.Context(), userID)
	if err != nil {
		respondGoogleError(c, "calendars", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"calendars": calendars})
}

// SelectCalendar chooses the calendar used for bookings.
func (h *GoogleController) SelectCalendar(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var request selectCalendarRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "calendar_id is required")
		return
	}

	calendar, err := h.selectionService.Select(c.Request.Context(), userID, request.CalendarID)
	if err != nil {
		respondGoogleError(c, "select calendar", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"calendar": calendar})
}

// SelectedCalendar returns the calendar currently chosen for bookings.
func (h *GoogleController) SelectedCalendar(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	calendar, err := h.selectionService.Selected(c.Request.Context(), userID)
	if err != nil {
		respondGoogleError(c, "selected calendar", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"calendar": calendar})
}

// authenticatedUserID reads the user ID set by the authentication middleware.
// On failure it writes the 401 response itself and returns false.
func authenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDValue, exists := commonmiddleware.GetAuthenticatedUserID(c)
	if !exists {
		respondError(c, http.StatusUnauthorized, "authentication required")
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid authenticated user")
		return uuid.Nil, false
	}

	return userID, true
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// respondGoogleError maps Google module domain errors to HTTP responses.
// Unexpected and provider errors are logged and never exposed to the client.
func respondGoogleError(c *gin.Context, operation string, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidOAuthState):
		respondError(c, http.StatusBadRequest, "invalid or expired OAuth state")
	case errors.Is(err, service.ErrGoogleAccountNotConnected):
		respondError(c, http.StatusConflict, "google account not connected, connect it first")
	case errors.Is(err, service.ErrGoogleAuthorizationRevoked):
		respondError(c, http.StatusConflict, "google authorization expired or revoked, reconnect your account")
	case errors.Is(err, service.ErrNoCalendarSelected):
		respondError(c, http.StatusNotFound, "no calendar selected")
	case errors.Is(err, service.ErrCalendarNotFound):
		respondError(c, http.StatusNotFound, "calendar not found in your google account")
	case errors.Is(err, service.ErrCalendarNotWritable):
		respondError(c, http.StatusUnprocessableEntity, "calendar is read-only, choose one you can write to")
	case errors.Is(err, service.ErrCodeExchangeFailed):
		log.Printf("google %s: %v", operation, err)
		respondError(c, http.StatusBadGateway, "failed to exchange Google authorization code")
	case errors.Is(err, service.ErrGoogleAPIFailure):
		log.Printf("google %s: %v", operation, err)
		respondError(c, http.StatusBadGateway, "failed to reach Google Calendar")
	default:
		log.Printf("google %s: %v", operation, err)
		respondError(c, http.StatusInternalServerError, "failed to process Google request")
	}
}