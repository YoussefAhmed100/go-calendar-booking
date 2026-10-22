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

const googleRequestTimeout = 15 * time.Second

// CalendarClient talks to the Google Calendar API and translates its errors into domain errors.
type CalendarClient struct{}

func NewCalendarClient() *CalendarClient {
	return &CalendarClient{}
}

// ListCalendars returns every calendar visible to the token, following pagination.
func (client *CalendarClient) ListCalendars(
	ctx context.Context,
	token *oauth2.Token,
) ([]Calendar, error) {
	ctx, cancel := context.WithTimeout(ctx, googleRequestTimeout)
	defer cancel()

	calendarService, err := calendar.NewService(
		ctx,
		option.WithTokenSource(oauth2.StaticTokenSource(token)),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: create calendar service: %v", ErrGoogleAPIFailure, err)
	}

	calendars := make([]Calendar, 0)

	err = calendarService.CalendarList.List().Context(ctx).Pages(ctx, func(page *calendar.CalendarList) error {
		for _, item := range page.Items {
			calendars = append(calendars, Calendar{
				ID:         item.Id,
				Summary:    item.Summary,
				Primary:    item.Primary,
				AccessRole: item.AccessRole,
				TimeZone:   item.TimeZone,
			})
		}
		return nil
	})
	if err != nil {
		return nil, translateGoogleError(err)
	}

	return calendars, nil
}

func translateGoogleError(err error) error {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) && isAuthorizationFailure(apiErr) {
		return ErrGoogleAuthorizationRevoked
	}

	return fmt.Errorf("%w: %v", ErrGoogleAPIFailure, err)
}

func isAuthorizationFailure(apiErr *googleapi.Error) bool {
	if apiErr.Code == http.StatusUnauthorized {
		return true
	}

	if apiErr.Code != http.StatusForbidden {
		return false
	}

	for _, detail := range apiErr.Errors {
		if detail.Reason == "insufficientPermissions" {
			return true
		}
	}

	return false
}