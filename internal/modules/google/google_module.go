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

type Module struct {
	db                       *gorm.DB
	googleController         *controller.GoogleController
	authenticationMiddleware *authmiddleware.AuthenticationMiddleware
}

func New(
	db *gorm.DB,
	cfg *config.Config,
	authenticationMiddleware *authmiddleware.AuthenticationMiddleware,
) *Module {
	googleService := service.NewGoogleService(db, cfg)

	return &Module{
		db:                       db,
		googleController:         controller.NewGoogleController(googleService),
		authenticationMiddleware: authenticationMiddleware,
	}
}

func (m *Module) Migrate() error {
	return m.db.AutoMigrate(
		&model.OAuthState{},
		&model.GoogleAccount{},
	)
}

func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	registerRoutes(
		router,
		m.googleController,
		m.authenticationMiddleware,
	)
}
