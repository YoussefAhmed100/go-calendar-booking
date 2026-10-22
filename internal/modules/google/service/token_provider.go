package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// TokenProvider returns a valid Google access token for a user,
// refreshing and persisting it when it has expired.
type TokenProvider struct {
	accounts    *GoogleAccountStore
	oauthConfig *oauth2.Config
}

func NewTokenProvider(accounts *GoogleAccountStore, oauthConfig *oauth2.Config) *TokenProvider {
	return &TokenProvider{accounts: accounts, oauthConfig: oauthConfig}
}

func (provider *TokenProvider) ValidToken(
	ctx context.Context,
	userID uuid.UUID,
) (*oauth2.Token, error) {
	account, err := provider.accounts.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	storedToken := &oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
		TokenType:    account.TokenType,
		Expiry:       account.ExpiresAt,
	}

	if !storedToken.Valid() && storedToken.RefreshToken == "" {
		return nil, ErrGoogleAuthorizationRevoked
	}

	freshToken, err := provider.oauthConfig.TokenSource(ctx, storedToken).Token()
	if err != nil {
		if isInvalidGrant(err) {
			return nil, ErrGoogleAuthorizationRevoked
		}
		return nil, fmt.Errorf("refresh Google token: %w", err)
	}

	if freshToken.AccessToken != storedToken.AccessToken {
		if err := provider.accounts.Upsert(ctx, userID, freshToken); err != nil {
			return nil, err
		}
	}

	return freshToken, nil
}

func isInvalidGrant(err error) bool {
	var retrieveErr *oauth2.RetrieveError
	return errors.As(err, &retrieveErr) && retrieveErr.ErrorCode == "invalid_grant"
}