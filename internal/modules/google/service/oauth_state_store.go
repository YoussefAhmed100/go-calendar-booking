package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"calendar-booking/internal/modules/google/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OAuthStateStore persists OAuth state values used to validate the Google callback.
type OAuthStateStore struct {
	db *gorm.DB
}

// NewOAuthStateStore creates a store backed by the given database.
func NewOAuthStateStore(db *gorm.DB) *OAuthStateStore {
	return &OAuthStateStore{db: db}
}

// Create saves a new state for the user.
func (s *OAuthStateStore) Create(
	ctx context.Context,
	userID uuid.UUID,
	state string,
	expiresAt time.Time,
) error {
	oauthState := model.OAuthState{
		ID:        uuid.New(),
		State:     state,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	if err := s.db.WithContext(ctx).Create(&oauthState).Error; err != nil {
		return fmt.Errorf("save OAuth state: %w", err)
	}

	return nil
}

// Consume marks the state as used and returns the user it was issued to.
// It returns ErrInvalidOAuthState when the state is unknown, expired, or already used.
func (s *OAuthStateStore) Consume(ctx context.Context, state string) (uuid.UUID, error) {
	var oauthState model.OAuthState

	now := time.Now().UTC()

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.OAuthState{}).
			Where("state = ? AND expires_at > ? AND consumed_at IS NULL", state, now).
			Update("consumed_at", now)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}

		return tx.Where("state = ?", state).First(&oauthState).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, ErrInvalidOAuthState
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("consume OAuth state: %w", err)
	}

	return oauthState.UserID, nil
}