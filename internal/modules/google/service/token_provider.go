package service

import (
	"context"
	"errors"
	"fmt"

	"calendar-booking/internal/modules/google/model"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// tokenAccountStore is the part of the account store that TokenProvider needs.
type tokenAccountStore interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*model.GoogleAccount, error)
	Upsert(ctx context.Context, userID uuid.UUID, token *oauth2.Token) error
}

// TokenProvider returns a valid Google access token for a user,
// refreshing and persisting it when it has expired.
type TokenProvider struct {
	accounts    tokenAccountStore
	oauthConfig *oauth2.Config
}

// NewTokenProvider creates a provider that reads and saves tokens through accounts.
func NewTokenProvider(accounts tokenAccountStore, oauthConfig *oauth2.Config) *TokenProvider {
	return &TokenProvider{accounts: accounts, oauthConfig: oauthConfig}
}

// ValidToken returns a usable access token for the user.
// It returns ErrGoogleAccountNotConnected when no account is linked, and
// ErrGoogleAuthorizationRevoked when the user must reconnect Google.
func (p *TokenProvider) ValidToken(
	ctx context.Context,
	userID uuid.UUID,
) (*oauth2.Token, error) {
	account, err := p.accounts.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	storedToken := tokenFromAccount(account)
	if !storedToken.Valid() && storedToken.RefreshToken == "" {
		return nil, ErrGoogleAuthorizationRevoked
	}

	freshToken, err := p.oauthConfig.TokenSource(ctx, storedToken).Token()
	if err != nil {
		if isInvalidGrant(err) {
			return nil, ErrGoogleAuthorizationRevoked
		}
		return nil, fmt.Errorf("%w: refresh Google token: %w", ErrGoogleAPIFailure, err)
	}

	if freshToken.AccessToken != storedToken.AccessToken {
		if err := p.accounts.Upsert(ctx, userID, freshToken); err != nil {
			return nil, err
		}
	}

	return freshToken, nil
}

func tokenFromAccount(account *model.GoogleAccount) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
		TokenType:    account.TokenType,
		Expiry:       account.ExpiresAt,
	}
}

func isInvalidGrant(err error) bool {
	var retrieveErr *oauth2.RetrieveError
	return errors.As(err, &retrieveErr) && retrieveErr.ErrorCode == "invalid_grant"
}