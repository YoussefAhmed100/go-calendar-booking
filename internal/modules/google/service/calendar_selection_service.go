package service

import (
	"context"

	"github.com/google/uuid"
)

// calendarLister lists the calendars of the Google account linked to a user.
type calendarLister interface {
	ListCalendars(ctx context.Context, userID uuid.UUID) ([]Calendar, error)
}

// selectedCalendarStore persists the calendar chosen for bookings.
type selectedCalendarStore interface {
	SetSelectedCalendar(ctx context.Context, userID uuid.UUID, calendarID string) error
	SelectedCalendarID(ctx context.Context, userID uuid.UUID) (string, error)
}

// CalendarSelectionService lets a user choose the calendar used for bookings.
type CalendarSelectionService struct {
	calendars calendarLister
	store     selectedCalendarStore
}

// NewCalendarSelectionService creates a service that validates and saves the chosen calendar.
func NewCalendarSelectionService(
	calendars calendarLister,
	store selectedCalendarStore,
) *CalendarSelectionService {
	return &CalendarSelectionService{calendars: calendars, store: store}
}

// Select saves the calendar after checking it belongs to the user's Google account
// and that bookings can be written to it.
func (s *CalendarSelectionService) Select(
	ctx context.Context,
	userID uuid.UUID,
	calendarID string,
) (*Calendar, error) {
	calendars, err := s.calendars.ListCalendars(ctx, userID)
	if err != nil {
		return nil, err
	}

	selected := findCalendar(calendars, calendarID)
	if selected == nil {
		return nil, ErrCalendarNotFound
	}
	if !selected.IsWritable() {
		return nil, ErrCalendarNotWritable
	}

	if err := s.store.SetSelectedCalendar(ctx, userID, selected.ID); err != nil {
		return nil, err
	}

	return selected, nil
}

// Selected returns the calendar currently chosen for bookings.
func (s *CalendarSelectionService) Selected(
	ctx context.Context,
	userID uuid.UUID,
) (*Calendar, error) {
	calendarID, err := s.store.SelectedCalendarID(ctx, userID)
	if err != nil {
		return nil, err
	}

	calendars, err := s.calendars.ListCalendars(ctx, userID)
	if err != nil {
		return nil, err
	}

	selected := findCalendar(calendars, calendarID)
	if selected == nil {
		return nil, ErrCalendarNotFound
	}

	return selected, nil
}

func findCalendar(calendars []Calendar, calendarID string) *Calendar {
	for i := range calendars {
		if calendars[i].ID == calendarID {
			return &calendars[i]
		}
	}

	return nil
}