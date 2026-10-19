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

func (s *GoogleService) GenerateAuthURL(
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

	if err := s.db.WithContext(ctx).Create(&oauthState).Error; err != nil {
		return "", fmt.Errorf("save OAuth state: %w", err)
	}

	return s.oauthConfig.AuthCodeURL(
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
