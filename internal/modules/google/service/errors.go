package service

import "errors"

var (
	ErrInvalidOAuthState          = errors.New("invalid or expired OAuth state")
	ErrCodeExchangeFailed         = errors.New("google authorization code exchange failed")
	ErrGoogleAccountNotConnected  = errors.New("google account not connected")
	ErrGoogleAuthorizationRevoked = errors.New("google authorization expired or revoked")
	ErrGoogleAPIFailure           = errors.New("google api request failed")
	ErrCalendarNotFound           = errors.New("calendar not found in the connected google account")
	ErrCalendarNotWritable        = errors.New("calendar is read-only")
	ErrNoCalendarSelected         = errors.New("no calendar selected")
)