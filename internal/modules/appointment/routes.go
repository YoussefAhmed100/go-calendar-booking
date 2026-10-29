package appointment

import (
	"calendar-booking/internal/modules/appointment/controller"
	"calendar-booking/internal/modules/auth/middleware"

	"github.com/gin-gonic/gin"
)

// registerRoutes attaches the appointments endpoints. Every route requires authentication.
func registerRoutes(
	router *gin.RouterGroup,
	appointmentController *controller.AppointmentController,
	authenticationMiddleware *middleware.AuthenticationMiddleware,
) {
	routes := router.Group("/appointments")
	routes.Use(authenticationMiddleware.RequireAuthentication())
	{
		routes.GET("/slots", appointmentController.Slots)
		routes.POST("", appointmentController.Create)
		routes.GET("", appointmentController.List)
		routes.GET("/:id", appointmentController.Get)
		routes.POST("/:id/cancel", appointmentController.Cancel)
	}
}
