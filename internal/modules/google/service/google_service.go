package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"calendar-booking/internal/modules/google/model"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

const (
	oauthStateExpiry     = 10 * time.Minute
	oauthStateBytesCount = 32
)

// GoogleService runs the Google OAuth authorization flow.
type GoogleService struct {
	db          *gorm.DB
	oauthConfig *oauth2.Config
	accounts    *GoogleAccountStore
}

func NewGoogleService(
	db *gorm.DB,
	oauthConfig *oauth2.Config,
	accounts *GoogleAccountStore,
) *GoogleService {
	return &GoogleService{db: db, oauthConfig: oauthConfig, accounts: accounts}
}

// GenerateAuthURL creates a Google OAuth authorization URL for the user
// and persists the generated state so the callback can be validated.
func (googleService *GoogleService) GenerateAuthURL(
	ctx context.Context,
	userID uuid.UUID,
) (string, error) {
	stateValue, err := generateRandomState()
	if err != nil {
		return "", fmt.Errorf("generate OAuth state: %w", err)
	}

	oauthState := model.OAuthState{
		ID:        uuid.New(),
		State:     stateValue,
		UserID:    userID,
		ExpiresAt: time.Now().UTC().Add(oauthStateExpiry),
	}

	if err := googleService.db.WithContext(ctx).Create(&oauthState).Error; err != nil {
		return "", fmt.Errorf("save OAuth state: %w", err)
	}

	return googleService.oauthConfig.AuthCodeURL(
		stateValue,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	), nil
}

// CompleteAuthorization validates the state, exchanges the code, and links the Google account.
// It returns ErrInvalidOAuthState or ErrCodeExchangeFailed for client-side failures.
func (googleService *GoogleService) CompleteAuthorization(
	ctx context.Context,
	code string,
	stateValue string,
) error {
	userID, err := googleService.consumeOAuthState(ctx, stateValue)
	if err != nil {
		return err
	}

	token, err := googleService.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCodeExchangeFailed, err)
	}

	return googleService.accounts.Upsert(ctx, userID, token)
}

func (googleService *GoogleService) consumeOAuthState(
	ctx context.Context,
	stateValue string,
) (uuid.UUID, error) {
	var oauthState model.OAuthState

	now := time.Now().UTC()

	err := googleService.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.OAuthState{}).
			Where(
				"state = ? AND expires_at > ? AND consumed_at IS NULL",
				stateValue,
				now,
			).
			Update("consumed_at", now)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}

		return tx.Where("state = ?", stateValue).First(&oauthState).Error
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, ErrInvalidOAuthState
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("consume OAuth state: %w", err)
	}

	return oauthState.UserID, nil
}

func generateRandomState() (string, error) {
	randomBytes := make([]byte, oauthStateBytesCount)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}