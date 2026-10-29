package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const (
	googleRequestTimeout         = 15 * time.Second
	reasonInsufficientPermission = "insufficientPermissions"
)

// CalendarClient is the only code that calls the Google Calendar API.
// It translates Google errors into domain errors.
type CalendarClient struct{}

// NewCalendarClient creates a Google Calendar API client.
func NewCalendarClient() *CalendarClient {
	return &CalendarClient{}
}

// ListCalendars returns every calendar visible to the token, following pagination.
func (c *CalendarClient) ListCalendars(
	ctx context.Context,
	token *oauth2.Token,
) ([]Calendar, error) {
	ctx, cancel := context.WithTimeout(ctx, googleRequestTimeout)
	defer cancel()

	calendarService, err := newCalendarService(ctx, token)
	if err != nil {
		return nil, err
	}

	calendars := make([]Calendar, 0)

	err = calendarService.CalendarList.List().Context(ctx).Pages(ctx, func(page *calendar.CalendarList) error {
		for _, entry := range page.Items {
			calendars = append(calendars, calendarFromEntry(entry))
		}
		return nil
	})
	if err != nil {
		return nil, translateGoogleError(err)
	}

	return calendars, nil
}

func newCalendarService(ctx context.Context, token *oauth2.Token) (*calendar.Service, error) {
	calendarService, err := calendar.NewService(
		ctx,
		option.WithTokenSource(oauth2.StaticTokenSource(token)),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: create calendar service: %w", ErrGoogleAPIFailure, err)
	}

	return calendarService, nil
}

func calendarFromEntry(entry *calendar.CalendarListEntry) Calendar {
	return Calendar{
		ID:         entry.Id,
		Summary:    entry.Summary,
		Primary:    entry.Primary,
		AccessRole: entry.AccessRole,
		TimeZone:   entry.TimeZone,
	}
}

func translateGoogleError(err error) error {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) && isAuthorizationFailure(apiErr) {
		return ErrGoogleAuthorizationRevoked
	}

	return fmt.Errorf("%w: %w", ErrGoogleAPIFailure, err)
}

func isAuthorizationFailure(apiErr *googleapi.Error) bool {
	switch apiErr.Code {
	case http.StatusUnauthorized:
		return true
	case http.StatusForbidden:
		for _, detail := range apiErr.Errors {
			if detail.Reason == reasonInsufficientPermission {
				return true
			}
		}
	}

	return false
}


// CreateEvent creates the event with the caller-chosen ID.
// A 409 means an earlier attempt already created it, which counts as success.
func (c *CalendarClient) CreateEvent(
	ctx context.Context,
	token *oauth2.Token,
	calendarID string,
	event CalendarEvent,
) error {
	ctx, cancel := context.WithTimeout(ctx, googleRequestTimeout)
	defer cancel()

	calendarService, err := newCalendarService(ctx, token)
	if err != nil {
		return err
	}

	body := &calendar.Event{
		Id:          event.ID,
		Summary:     event.Title,
		Description: event.Description,
		Start:       &calendar.EventDateTime{DateTime: event.Start.Format(time.RFC3339), TimeZone: event.TimeZone},
		End:         &calendar.EventDateTime{DateTime: event.End.Format(time.RFC3339), TimeZone: event.TimeZone},
	}

	_, err = calendarService.Events.Insert(calendarID, body).Context(ctx).Do()
	if hasStatusCode(err, http.StatusConflict) {
		return nil
	}
	if err != nil {
		return translateGoogleError(err)
	}

	return nil
}

// DeleteEvent deletes the event. An event that is already gone (404 or 410) counts as success.
func (c *CalendarClient) DeleteEvent(
	ctx context.Context,
	token *oauth2.Token,
	calendarID string,
	eventID string,
) error {
	ctx, cancel := context.WithTimeout(ctx, googleRequestTimeout)
	defer cancel()

	calendarService, err := newCalendarService(ctx, token)
	if err != nil {
		return err
	}

	err = calendarService.Events.Delete(calendarID, eventID).Context(ctx).Do()
	if hasStatusCode(err, http.StatusNotFound, http.StatusGone) {
		return nil
	}
	if err != nil {
		return translateGoogleError(err)
	}

	return nil
}

func hasStatusCode(err error, codes ...int) bool {
	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) {
		return false
	}

	for _, code := range codes {
		if apiErr.Code == code {
			return true
		}
	}

	return false
}