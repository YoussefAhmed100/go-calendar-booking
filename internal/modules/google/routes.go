package google

import (
	"calendar-booking/internal/modules/auth/middleware"
	"calendar-booking/internal/modules/google/controller"

	"github.com/gin-gonic/gin"
)

func registerRoutes(
	router *gin.RouterGroup,
	googleController *controller.GoogleController,
	authenticationMiddleware *middleware.AuthenticationMiddleware,
) {
	googleRoutes := router.Group("/google")

	googleRoutes.GET(
		"/connect",
		authenticationMiddleware.RequireAuthentication(),
		googleController.Connect,
	)

	googleRoutes.GET(
		"/calendars",
		authenticationMiddleware.RequireAuthentication(),
		googleController.Calendars,
	)

	googleRoutes.GET(
		"/callback",
		googleController.Callback,
	)
}