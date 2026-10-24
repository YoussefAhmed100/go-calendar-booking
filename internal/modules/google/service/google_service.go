package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

const (
	oauthStateExpiry     = 10 * time.Minute
	oauthStateByteLength = 32
)

// oauthStateStore saves and consumes the state that protects the OAuth callback.
type oauthStateStore interface {
	Create(ctx context.Context, userID uuid.UUID, state string, expiresAt time.Time) error
	Consume(ctx context.Context, state string) (uuid.UUID, error)
}

// accountUpserter links a Google token to an application user.
type accountUpserter interface {
	Upsert(ctx context.Context, userID uuid.UUID, token *oauth2.Token) error
}

// GoogleService runs the Google OAuth authorization flow.
type GoogleService struct {
	states      oauthStateStore
	accounts    accountUpserter
	oauthConfig *oauth2.Config
}

// NewGoogleService creates the service that connects a Google account to a user.
func NewGoogleService(
	states oauthStateStore,
	accounts accountUpserter,
	oauthConfig *oauth2.Config,
) *GoogleService {
	return &GoogleService{states: states, accounts: accounts, oauthConfig: oauthConfig}
}

// GenerateAuthURL creates a Google OAuth authorization URL for the user
// and persists the generated state so the callback can be validated.
func (s *GoogleService) GenerateAuthURL(
	ctx context.Context,
	userID uuid.UUID,
) (string, error) {
	state, err := generateRandomState()
	if err != nil {
		return "", fmt.Errorf("generate OAuth state: %w", err)
	}

	expiresAt := time.Now().UTC().Add(oauthStateExpiry)
	if err := s.states.Create(ctx, userID, state, expiresAt); err != nil {
		return "", err
	}

	return s.oauthConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	), nil
}

// CompleteAuthorization validates the state, exchanges the code, and links the Google account.
// It returns ErrInvalidOAuthState or ErrCodeExchangeFailed for client-side failures.
func (s *GoogleService) CompleteAuthorization(
	ctx context.Context,
	code string,
	state string,
) error {
	userID, err := s.states.Consume(ctx, state)
	if err != nil {
		return err
	}

	exchangeCtx, cancel := context.WithTimeout(ctx, googleRequestTimeout)
	defer cancel()

	token, err := s.oauthConfig.Exchange(exchangeCtx, code)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCodeExchangeFailed, err)
	}

	return s.accounts.Upsert(ctx, userID, token)
}

func generateRandomState() (string, error) {
	randomBytes := make([]byte, oauthStateByteLength)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}