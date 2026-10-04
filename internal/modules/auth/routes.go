package auth

import (
	"github.com/gin-gonic/gin"

	"calendar-booking/internal/modules/auth/controller"
)

func registerRoutes(
	router *gin.RouterGroup,
	authController *controller.AuthController,
) {
	authRoutes := router.Group("/auth")

	authRoutes.POST(
		"/register",
		authController.Register,
	)
}