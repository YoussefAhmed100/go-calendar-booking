package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"calendar-booking/internal/common/config"
	"calendar-booking/internal/modules/google/model"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

const oauthStateExpiry = 10 * time.Minute

type GoogleService struct {
	db          *gorm.DB
	oauthConfig *oauth2.Config
}

// @desc => NewGoogleService creates a new instance of GoogleService with the provided database connection and configuration.
func NewGoogleService(
	db *gorm.DB,
	cfg *config.Config,
) *GoogleService {
	return &GoogleService{
		db: db,
		oauthConfig: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Endpoint:     google.Endpoint,
			Scopes: []string{
				"https://www.googleapis.com/auth/calendar.events",
			},
		},
	}
}

// GenerateAuthURL generates a Google OAuth authorization URL for a user and saves the associated state in the database.
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
	), nil
}

func generateRandomState() (string, error) {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

// ConsumeOAuthState consumes an OAuth state and returns the associated user ID.
func (googleService *GoogleService) ConsumeOAuthState(
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

		return tx.
			Where("state = ?", stateValue).
			First(&oauthState).Error
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("consume OAuth state: %w", err)
	}

	return oauthState.UserID, nil
}

// ExchangeCode exchanges an authorization code for an access token and refresh token.
func (googleService *GoogleService) ExchangeCode(
	ctx context.Context,
	code string,
) (*oauth2.Token, error) {
	token, err := googleService.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange Google authorization code: %w", err)
	}

	return token, nil
}

// SaveGoogleAccount saves a Google account for a user.

func (googleService *GoogleService) SaveGoogleAccount(
	ctx context.Context,
	userID uuid.UUID,
	token *oauth2.Token,
) error {
	var account model.GoogleAccount

	err := googleService.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&account).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("find Google account: %w", err)
	}

	// Preserve the existing refresh token if Google doesn't return a new one.
	refreshToken := token.RefreshToken
	if refreshToken == "" && err == nil {
		refreshToken = account.RefreshToken
	}

	if err == gorm.ErrRecordNotFound {
		account = model.GoogleAccount{
			ID:     uuid.New(),
			UserID: userID,
		}
	}

	account.AccessToken = token.AccessToken
	account.RefreshToken = refreshToken
	account.TokenType = token.TokenType
	account.ExpiresAt = token.Expiry.UTC()

	if err := googleService.db.WithContext(ctx).Save(&account).Error; err != nil {
		return fmt.Errorf("save Google account: %w", err)
	}

	return nil
}
