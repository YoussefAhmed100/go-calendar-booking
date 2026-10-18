package auth

import (
	"github.com/gin-gonic/gin"

	"calendar-booking/internal/modules/auth/controller"
	"calendar-booking/internal/modules/auth/middleware"
)

func registerRoutes(
	router *gin.RouterGroup,
	authController *controller.AuthController,
	authenticationMiddleware *middleware.AuthenticationMiddleware,
) {
	authRoutes := router.Group("/auth")

	authRoutes.POST(
		"/register",
		authController.Register,
	)

	authRoutes.POST(
		"/login",
		authController.Login,
	)

	authRoutes.POST(
		"/logout",
		authenticationMiddleware.RequireAuthentication(),
		authController.Logout,
	)
}