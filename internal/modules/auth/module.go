package auth

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"calendar-booking/internal/modules/auth/model"
)

type Module struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Module {
	return &Module{
		db: db,
	}
}

func (m *Module) Migrate() error {
	return m.db.AutoMigrate(
		&model.User{},
		&model.Session{},
	)
}
func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	// Auth routes will be registered in a later step.
}
