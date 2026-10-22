package service

import (
	"calendar-booking/internal/common/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// NewOAuthConfig builds the Google OAuth settings used across the Google module.
func NewOAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.events",
			"https://www.googleapis.com/auth/calendar.calendarlist.readonly",
		},
	}
}
