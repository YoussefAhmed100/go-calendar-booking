package middleware

import (
	"net/http"
	"strings"
	"time"

	commonmiddleware "calendar-booking/internal/common/middleware"
	"calendar-booking/internal/modules/auth/model"
	"calendar-booking/internal/modules/auth/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthenticationMiddleware struct {
	databaseConnection *gorm.DB
	jwtService         *service.JWTService
}

func NewAuthenticationMiddleware(
	databaseConnection *gorm.DB,
	jwtService *service.JWTService,
) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		databaseConnection: databaseConnection,
		jwtService:         jwtService,
	}
}

func (middleware *AuthenticationMiddleware) RequireAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader("Authorization")

		authenticationScheme, token, found := strings.Cut(
			authorizationHeader,
			" ",
		)

		if !found ||
			!strings.EqualFold(authenticationScheme, "Bearer") ||
			strings.TrimSpace(token) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}

		token = strings.TrimSpace(token)

		claims, err := middleware.jwtService.ParseToken(token)
		if err != nil || claims.Subject == "" || claims.ID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		authenticatedUserID, err := uuid.Parse(claims.Subject)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		authenticatedSessionID, err := uuid.Parse(claims.ID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		var authenticatedSession model.Session

		err = middleware.databaseConnection.
			WithContext(c.Request.Context()).
			Where(
				"id = ? AND user_id = ? AND revoked_at IS NULL AND expires_at > ?",
				authenticatedSessionID,
				authenticatedUserID,
				time.Now().UTC(),
			).
			First(&authenticatedSession).
			Error

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		c.Set(
			commonmiddleware.AuthenticatedUserIDKey,
			authenticatedUserID.String(),
		)

		c.Set(
			commonmiddleware.AuthenticatedSessionIDKey,
			authenticatedSessionID.String(),
		)

		c.Next()
	}
}