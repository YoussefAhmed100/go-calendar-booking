package controller

import (
	"net/http"

	commonmiddleware "calendar-booking/internal/common/middleware"
	"calendar-booking/internal/modules/google/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GoogleController struct {
	googleService *service.GoogleService
}

func NewGoogleController(
	googleService *service.GoogleService,
) *GoogleController {
	return &GoogleController{
		googleService: googleService,
	}
}

func (controller *GoogleController) Connect(c *gin.Context) {
	userIDValue, exists := commonmiddleware.GetAuthenticatedUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid authenticated user",
		})
		return
	}

	authURL, err := controller.googleService.GenerateAuthURL(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to start Google authorization",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

func (controller *GoogleController) Callback(context *gin.Context) {
	if googleError := context.Query("error"); googleError != "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Google authorization was denied or failed",
		})
		return
	}

	code := context.Query("code")
	state := context.Query("state")

	if code == "" || state == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "missing OAuth code or state",
		})
		return
	}

	userID, err := controller.googleService.ConsumeOAuthState(
		context.Request.Context(),
		state,
	)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid or expired OAuth state",
		})
		return
	}

	token, err := controller.googleService.ExchangeCode(
		context.Request.Context(),
		code,
	)
	if err != nil {
		context.JSON(http.StatusBadGateway, gin.H{
			"error": "failed to exchange Google authorization code",
		})
		return
	}

	if err := controller.googleService.SaveGoogleAccount(
		context.Request.Context(),
		userID,
		token,
	); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save Google account",
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": "Google Calendar connected successfully",
	})
}
