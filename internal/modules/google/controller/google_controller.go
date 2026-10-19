
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

