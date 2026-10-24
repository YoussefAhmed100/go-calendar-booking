package google

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"calendar-booking/internal/common/config"
	authmiddleware "calendar-booking/internal/modules/auth/middleware"
	"calendar-booking/internal/modules/google/controller"
	"calendar-booking/internal/modules/google/model"
	"calendar-booking/internal/modules/google/service"
)

// Module wires the Google integration: OAuth connection, calendar listing, and calendar selection.
type Module struct {
	db                       *gorm.DB
	googleController         *controller.GoogleController
	authenticationMiddleware *authmiddleware.AuthenticationMiddleware
}

// New builds the Google module and its dependencies.
func New(
	db *gorm.DB,
	cfg *config.Config,
	authenticationMiddleware *authmiddleware.AuthenticationMiddleware,
) *Module {
	oauthConfig := service.NewOAuthConfig(cfg)
	accountStore := service.NewGoogleAccountStore(db)
	oauthStateStore := service.NewOAuthStateStore(db)
	tokenProvider := service.NewTokenProvider(accountStore, oauthConfig)

	googleService := service.NewGoogleService(oauthStateStore, accountStore, oauthConfig)
	calendarService := service.NewCalendarService(tokenProvider, service.NewCalendarClient())
	selectionService := service.NewCalendarSelectionService(calendarService, accountStore)

	return &Module{
		db: db,
		googleController: controller.NewGoogleController(
			googleService,
			calendarService,
			selectionService,
		),
		authenticationMiddleware: authenticationMiddleware,
	}
}

// Migrate creates or updates the Google module tables.
func (m *Module) Migrate() error {
	return m.db.AutoMigrate(
		&model.OAuthState{},
		&model.GoogleAccount{},
	)
}

// RegisterRoutes attaches the Google endpoints to the router group.
func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	registerRoutes(
		router,
		m.googleController,
		m.authenticationMiddleware,
	)
}