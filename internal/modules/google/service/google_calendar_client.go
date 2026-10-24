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