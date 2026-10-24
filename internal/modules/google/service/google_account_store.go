package service

import (
	"context"
	"errors"
	"fmt"

	"calendar-booking/internal/modules/google/model"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GoogleAccountStore persists the Google account linked to an application user.
type GoogleAccountStore struct {
	db *gorm.DB
}

// NewGoogleAccountStore creates a store backed by the given database.
func NewGoogleAccountStore(db *gorm.DB) *GoogleAccountStore {
	return &GoogleAccountStore{db: db}
}

// FindByUserID returns ErrGoogleAccountNotConnected when the user has no linked account.
func (s *GoogleAccountStore) FindByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*model.GoogleAccount, error) {
	var account model.GoogleAccount

	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrGoogleAccountNotConnected
	}
	if err != nil {
		return nil, fmt.Errorf("find Google account: %w", err)
	}

	return &account, nil
}

// Upsert creates or updates the user's Google account.
// An empty refresh token in the incoming token keeps the stored one.
func (s *GoogleAccountStore) Upsert(
	ctx context.Context,
	userID uuid.UUID,
	token *oauth2.Token,
) error {
	account := model.GoogleAccount{
		ID:           uuid.New(),
		UserID:       userID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry.UTC(),
	}

	updateColumns := []string{"access_token", "token_type", "expires_at", "updated_at"}
	if token.RefreshToken != "" {
		updateColumns = append(updateColumns, "refresh_token")
	}

	err := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns(updateColumns),
		}).
		Create(&account).Error
	if err != nil {
		return fmt.Errorf("save Google account: %w", err)
	}

	return nil
}

// SetSelectedCalendar stores the calendar chosen for bookings.
// It returns ErrGoogleAccountNotConnected when the user has no linked account.
func (s *GoogleAccountStore) SetSelectedCalendar(
	ctx context.Context,
	userID uuid.UUID,
	calendarID string,
) error {
	result := s.db.WithContext(ctx).
		Model(&model.GoogleAccount{}).
		Where("user_id = ?", userID).
		Update("selected_calendar_id", calendarID)
	if result.Error != nil {
		return fmt.Errorf("save selected calendar: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrGoogleAccountNotConnected
	}

	return nil
}

// SelectedCalendarID returns ErrNoCalendarSelected when the user has not chosen one yet.
func (s *GoogleAccountStore) SelectedCalendarID(
	ctx context.Context,
	userID uuid.UUID,
) (string, error) {
	account, err := s.FindByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if account.SelectedCalendarID == nil {
		return "", ErrNoCalendarSelected
	}

	return *account.SelectedCalendarID, nil
}