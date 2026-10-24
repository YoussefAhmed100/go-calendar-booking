package service

import (
	"calendar-booking/internal/common/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	scopeCalendarEvents       = "https://www.googleapis.com/auth/calendar.events"
	scopeCalendarListReadonly = "https://www.googleapis.com/auth/calendar.calendarlist.readonly"
)

// NewOAuthConfig builds the Google OAuth2 configuration from the app config.
func NewOAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       []string{scopeCalendarEvents, scopeCalendarListReadonly},
	}
}
