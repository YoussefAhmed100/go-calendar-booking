package google

import (
	"calendar-booking/internal/modules/auth/middleware"
	"calendar-booking/internal/modules/google/controller"

	"github.com/gin-gonic/gin"
)

// registerRoutes attaches the Google endpoints. The callback is public because Google
// calls it without a Bearer token; every other route requires authentication.
func registerRoutes(
	router *gin.RouterGroup,
	googleController *controller.GoogleController,
	authenticationMiddleware *middleware.AuthenticationMiddleware,
) {
	googleRoutes := router.Group("/google")

	googleRoutes.GET("/callback", googleController.Callback)

	protected := googleRoutes.Group("")
	protected.Use(authenticationMiddleware.RequireAuthentication())
	{
		protected.GET("/connect", googleController.Connect)
		protected.GET("/calendars", googleController.Calendars)
		protected.GET("/calendar", googleController.SelectedCalendar)
		protected.PUT("/calendar", googleController.SelectCalendar)
	}
}