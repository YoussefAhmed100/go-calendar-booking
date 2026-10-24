package service

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// userTokenProvider supplies a valid Google access token for a user.
type userTokenProvider interface {
	ValidToken(ctx context.Context, userID uuid.UUID) (*oauth2.Token, error)
}

// calendarFetcher fetches the calendars visible to a token.
type calendarFetcher interface {
	ListCalendars(ctx context.Context, token *oauth2.Token) ([]Calendar, error)
}

// CalendarService lists the calendars of the Google account linked to a user.
type CalendarService struct {
	tokens    userTokenProvider
	calendars calendarFetcher
}

// NewCalendarService creates a service that lists calendars using the user's token.
func NewCalendarService(tokens userTokenProvider, calendars calendarFetcher) *CalendarService {
	return &CalendarService{tokens: tokens, calendars: calendars}
}

// ListCalendars returns the calendars available to the user's linked Google account.
func (s *CalendarService) ListCalendars(
	ctx context.Context,
	userID uuid.UUID,
) ([]Calendar, error) {
	token, err := s.tokens.ValidToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.calendars.ListCalendars(ctx, token)
}