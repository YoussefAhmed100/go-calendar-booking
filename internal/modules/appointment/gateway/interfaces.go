package gateway

import (
    "context"

    googleservice "calendar-booking/internal/modules/google/service"

    "github.com/google/uuid"
)

type calendarSelector interface {
    SelectedCalendarID(ctx context.Context, userID uuid.UUID) (string, error)
}

type eventWriter interface {
    CreateEvent(
        ctx context.Context,
        userID uuid.UUID,
        calendarID string,
        event googleservice.CalendarEvent,
    ) error

    DeleteEvent(
        ctx context.Context,
        userID uuid.UUID,
        calendarID string,
        eventID string,
    ) error
}