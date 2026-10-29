package appointment

import (
	"fmt"

	"calendar-booking/internal/common/config"
	authmiddleware "calendar-booking/internal/modules/auth/middleware"
	"calendar-booking/internal/modules/appointment/controller"
	"calendar-booking/internal/modules/appointment/gateway"
	"calendar-booking/internal/modules/appointment/repository"
	"calendar-booking/internal/modules/appointment/service"
	"calendar-booking/internal/modules/google"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module wires the appointments feature: slots, booking, listing, and cancellation.
type Module struct {
	db                       *gorm.DB
	appointmentController    *controller.AppointmentController
	authenticationMiddleware *authmiddleware.AuthenticationMiddleware
}

// New builds the appointments module. It fails when the booking configuration is invalid.
func New(
	db *gorm.DB,
	cfg *config.Config,
	googleModule *google.Module,
	authenticationMiddleware *authmiddleware.AuthenticationMiddleware,
) (*Module, error) {
	hours, err := service.NewWorkingHours(cfg)
	if err != nil {
		return nil, fmt.Errorf("booking configuration: %w", err)
	}

	appointmentRepository := repository.NewAppointmentRepository(db)
	calendarGateway := gateway.NewCalendarGateway(googleModule.Accounts(), googleModule.Events())
	appointmentService := service.NewAppointmentService(appointmentRepository, calendarGateway, hours)

	return &Module{
		db:                       db,
		appointmentController:    controller.NewAppointmentController(appointmentService),
		authenticationMiddleware: authenticationMiddleware,
	}, nil
}

// Migrate prepares the appointments schema.
func (m *Module) Migrate() error {
	return migrate(m.db)
}

// RegisterRoutes attaches the appointments endpoints to the router group.
func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	registerRoutes(router, m.appointmentController, m.authenticationMiddleware)
}