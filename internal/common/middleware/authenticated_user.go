package middleware

import (
	"net/http"

	"calendar-booking/internal/common/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthenticatedUserID reads the user ID set by the authentication middleware.
// On failure it writes the 401 response itself and returns false.
func AuthenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := GetAuthenticatedUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "authentication required")
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(value)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return uuid.Nil, false
	}

	return userID, true
}
