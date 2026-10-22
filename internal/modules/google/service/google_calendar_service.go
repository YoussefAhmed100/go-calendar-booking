package service

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type tokenSource interface {
	ValidToken(ctx context.Context, userID uuid.UUID) (*oauth2.Token, error)
}

type calendarProvider interface {
	ListCalendars(ctx context.Context, token *oauth2.Token) ([]Calendar, error)
}

// CalendarService lists the calendars of the Google account linked to a user.
type CalendarService struct {
	tokens    tokenSource
	calendars calendarProvider
}

func NewCalendarService(tokens tokenSource, calendars calendarProvider) *CalendarService {
	return &CalendarService{tokens: tokens, calendars: calendars}
}

func (service *CalendarService) ListCalendars(
	ctx context.Context,
	userID uuid.UUID,
) ([]Calendar, error) {
	token, err := service.tokens.ValidToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	return service.calendars.ListCalendars(ctx, token)
}