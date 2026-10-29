package gateway

import (
	"context"
	"errors"
	"fmt"

	"calendar-booking/internal/modules/appointment/service"
	googleservice "calendar-booking/internal/modules/google/service"

	"github.com/google/uuid"
)

type calendarSelector interface {
	SelectedCalendarID(ctx context.Context, userID uuid.UUID) (string, error)
}

type eventWriter interface {
	CreateEvent(ctx context.Context, userID uuid.UUID, calendarID string, event googleservice.CalendarEvent) error
	DeleteEvent(ctx context.Context, userID uuid.UUID, calendarID, eventID string) error
}

// CalendarGateway lets the appointments module read the selected calendar and write events to it.
// It translates Google module errors into appointments domain errors.
type CalendarGateway struct {
	selector calendarSelector
	events   eventWriter
}

// NewCalendarGateway creates the adapter between appointments and the Google module.
func NewCalendarGateway(selector calendarSelector, events eventWriter) *CalendarGateway {
	return &CalendarGateway{selector: selector, events: events}
}

// SelectedCalendarID returns the calendar chosen for bookings.
func (g *CalendarGateway) SelectedCalendarID(ctx context.Context, userID uuid.UUID) (string, error) {
	calendarID, err := g.selector.SelectedCalendarID(ctx, userID)
	if err != nil {
		return "", translateError(err)
	}

	return calendarID, nil
}

// CreateEvent creates the event in the given calendar.
func (g *CalendarGateway) CreateEvent(
	ctx context.Context,
	userID uuid.UUID,
	calendarID string,
	event service.CalendarEvent,
) error {
	err := g.events.CreateEvent(ctx, userID, calendarID, googleservice.CalendarEvent{
		ID:          event.ID,
		Title:       event.Title,
		Description: event.Description,
		Start:       event.Start,
		End:         event.End,
		TimeZone:    event.TimeZone,
	})

	return translateError(err)
}

// DeleteEvent deletes the event from the given calendar.
func (g *CalendarGateway) DeleteEvent(
	ctx context.Context,
	userID uuid.UUID,
	calendarID string,
	eventID string,
) error {
	return translateError(g.events.DeleteEvent(ctx, userID, calendarID, eventID))
}

// translateError maps Google module errors to appointments domain errors,
// keeping the original error for logs. Other errors pass through unchanged.
func translateError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, googleservice.ErrGoogleAccountNotConnected):
		return fmt.Errorf("%w: %w", service.ErrCalendarNotConnected, err)
	case errors.Is(err, googleservice.ErrGoogleAuthorizationRevoked):
		return fmt.Errorf("%w: %w", service.ErrCalendarAuthorizationRevoked, err)
	case errors.Is(err, googleservice.ErrNoCalendarSelected):
		return fmt.Errorf("%w: %w", service.ErrNoCalendarSelected, err)
	case errors.Is(err, googleservice.ErrGoogleAPIFailure):
		return fmt.Errorf("%w: %w", service.ErrCalendarSyncFailed, err)
	default:
		return err
	}
}
// Compile-time check that the gateway satisfies the appointments contract.
var _ service.CalendarGateway = (*CalendarGateway)(nil)