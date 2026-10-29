package service

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// eventWriter creates and deletes events in Google Calendar.
type eventWriter interface {
	CreateEvent(ctx context.Context, token *oauth2.Token, calendarID string, event CalendarEvent) error
	DeleteEvent(ctx context.Context, token *oauth2.Token, calendarID, eventID string) error
}

// EventService writes events to the calendar of the Google account linked to a user.
type EventService struct {
	tokens userTokenProvider
	events eventWriter
}

// NewEventService creates a service that writes events using the user's token.
func NewEventService(tokens userTokenProvider, events eventWriter) *EventService {
	return &EventService{tokens: tokens, events: events}
}

// CreateEvent creates the event in the given calendar.
func (s *EventService) CreateEvent(
	ctx context.Context,
	userID uuid.UUID,
	calendarID string,
	event CalendarEvent,
) error {
	token, err := s.tokens.ValidToken(ctx, userID)
	if err != nil {
		return err
	}

	return s.events.CreateEvent(ctx, token, calendarID, event)
}

// DeleteEvent deletes the event from the given calendar.
func (s *EventService) DeleteEvent(
	ctx context.Context,
	userID uuid.UUID,
	calendarID string,
	eventID string,
) error {
	token, err := s.tokens.ValidToken(ctx, userID)
	if err != nil {
		return err
	}

	return s.events.DeleteEvent(ctx, token, calendarID, eventID)
}