package auth

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"calendar-booking/internal/common/config"
	"calendar-booking/internal/modules/auth/controller"
	"calendar-booking/internal/modules/auth/model"
	"calendar-booking/internal/modules/auth/service"
)

type Module struct {
	db             *gorm.DB
	authController *controller.AuthController
}

func New(db *gorm.DB, cfg *config.Config) *Module {
	jwtService := service.NewJWTService(
		cfg.JWTSecret,
		cfg.JWTExpiryHours,
	)

	authService := service.NewAuthService(
		db,
		jwtService,
	)

	authController := controller.NewAuthController(
		authService,
	)

	return &Module{
		db:             db,
		authController: authController,
	}
}

func (m *Module) Migrate() error {
	return m.db.AutoMigrate(
		&model.User{},
		&model.Session{},
	)
}

func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	registerRoutes(router, m.authController)
}