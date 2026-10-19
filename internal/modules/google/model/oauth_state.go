package model

import (
	"time"

	"github.com/google/uuid"
)

type OAuthState struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	State      string     `gorm:"uniqueIndex;not null"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	ExpiresAt  time.Time  `gorm:"not null;index"`
	ConsumedAt *time.Time
	CreatedAt  time.Time
}