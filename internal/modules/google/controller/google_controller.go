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

type GoogleController struct {
	googleService   *service.GoogleService
	calendarService *service.CalendarService
}

func NewGoogleController(
	googleService *service.GoogleService,
	calendarService *service.CalendarService,
) *GoogleController {
	return &GoogleController{
		googleService:   googleService,
		calendarService: calendarService,
	}
}

// Connect returns the Google authorization URL for the authenticated user.
func (controller *GoogleController) Connect(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	authURL, err := controller.googleService.GenerateAuthURL(c.Request.Context(), userID)
	if err != nil {
		log.Printf("google connect: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to start Google authorization")
		return
	}

	c.JSON(http.StatusOK, gin.H{"auth_url": authURL})
}

// Callback completes the OAuth flow. It is called by Google, so it has no Bearer token;
// the user is identified through the persisted OAuth state.
func (controller *GoogleController) Callback(c *gin.Context) {
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

	err := controller.googleService.CompleteAuthorization(c.Request.Context(), code, state)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"message": "Google Calendar connected successfully"})
	case errors.Is(err, service.ErrInvalidOAuthState):
		respondError(c, http.StatusBadRequest, "invalid or expired OAuth state")
	case errors.Is(err, service.ErrCodeExchangeFailed):
		log.Printf("google callback: %v", err)
		respondError(c, http.StatusBadGateway, "failed to exchange Google authorization code")
	default:
		log.Printf("google callback: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to connect Google account")
	}
}

// Calendars lists the calendars of the Google account linked to the authenticated user.
func (controller *GoogleController) Calendars(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	calendars, err := controller.calendarService.ListCalendars(c.Request.Context(), userID)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"calendars": calendars})
	case errors.Is(err, service.ErrGoogleAccountNotConnected):
		respondError(c, http.StatusConflict, "google account not connected, connect it first")
	case errors.Is(err, service.ErrGoogleAuthorizationRevoked):
		respondError(c, http.StatusConflict, "google authorization expired or revoked, reconnect your account")
	case errors.Is(err, service.ErrGoogleAPIFailure):
		log.Printf("google calendars: %v", err)
		respondError(c, http.StatusBadGateway, "failed to fetch calendars from Google")
	default:
		log.Printf("google calendars: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to list calendars")
	}
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