package model

import (
	"time"

	"github.com/google/uuid"
)

type GoogleAccount struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	AccessToken  string    `gorm:"type:text;not null"`
	RefreshToken string    `gorm:"type:text"`
	TokenType    string    `gorm:"not null"`
	ExpiresAt    time.Time `gorm:"not null"`
	SelectedCalendarID *string   `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
