package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"calendar-booking/internal/modules/appointment/model"
	"calendar-booking/internal/modules/appointment/repository"

	"github.com/google/uuid"
)

const (
	dateLayout   = "2006-01-02"
	defaultLimit = 50
	maxLimit     = 100
)

// CreateInput is a booking request.
type CreateInput struct {
	UserID          uuid.UUID
	IdempotencyKey  string
	Title           string
	CustomerName    string
	CustomerEmail   string
	StartAt         time.Time
	DurationMinutes int
	Timezone        string
}

// ListFilter narrows a list of appointments. Zero times mean unbounded.
type ListFilter struct {
	From  time.Time
	To    time.Time
	Limit int
}

// appointmentService implements AppointmentService.
type appointmentService struct {
	appointments AppointmentRepository
	calendar     CalendarGateway
	hours        WorkingHours
}

// NewAppointmentService creates the appointment service.
func NewAppointmentService(
	appointments AppointmentRepository,
	calendar CalendarGateway,
	hours WorkingHours,
) AppointmentService {
	return &appointmentService{appointments: appointments, calendar: calendar, hours: hours}
}

func (s *appointmentService) Slots(
	ctx context.Context,
	userID uuid.UUID,
	date string,
	durationMinutes int,
) ([]TimeRange, error) {
	duration := time.Duration(durationMinutes) * time.Minute
	if !s.hours.ValidDuration(duration) {
		return nil, ErrInvalidDuration
	}

	day, err := time.ParseInLocation(dateLayout, date, s.hours.Location())
	if err != nil {
		return nil, ErrInvalidDate
	}

	calendarID, err := s.calendar.SelectedCalendarID(ctx, userID)
	if err != nil {
		return nil, err
	}

	booked, err := s.appointments.ListActive(ctx, calendarID, day, day.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}

	return s.hours.FreeSlots(day, duration, toRanges(booked), time.Now()), nil
}

// Create saves the row as pending first, then calls Google outside any DB transaction.
// If Google fails, the row stays pending and retrying with the same key finishes the sync.
func (s *appointmentService) Create(
	ctx context.Context,
	input CreateInput,
) (*model.Appointment, bool, error) {
	existing, err := s.appointments.FindByIdempotencyKey(ctx, input.UserID, input.IdempotencyKey)
	if err == nil {
		return s.replay(ctx, existing, input)
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, false, err
	}

	duration := time.Duration(input.DurationMinutes) * time.Minute
	if err := s.validate(input, duration); err != nil {
		return nil, false, err
	}

	calendarID, err := s.calendar.SelectedCalendarID(ctx, input.UserID)
	if err != nil {
		return nil, false, err
	}

	appointment := newAppointment(input, calendarID, duration)

	if err := s.appointments.Create(ctx, appointment); err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			return s.replayConcurrent(ctx, input)
		}
		return nil, false, translateRepositoryError(err)
	}

	if err := s.syncCreate(ctx, appointment); err != nil {
		return nil, false, err
	}

	return appointment, true, nil
}

func (s *appointmentService) Get(
	ctx context.Context,
	userID uuid.UUID,
	id uuid.UUID,
) (*model.Appointment, error) {
	appointment, err := s.appointments.FindByID(ctx, userID, id)
	if err != nil {
		return nil, translateRepositoryError(err)
	}

	return appointment, nil
}

func (s *appointmentService) List(
	ctx context.Context,
	userID uuid.UUID,
	filter ListFilter,
) ([]model.Appointment, error) {
	if filter.Limit <= 0 {
		filter.Limit = defaultLimit
	}
	if filter.Limit > maxLimit {
		filter.Limit = maxLimit
	}

	return s.appointments.List(ctx, userID, filter.From, filter.To, filter.Limit)
}

// Cancel is idempotent: cancelling twice, or after the event was deleted by hand, succeeds.
// If Google fails, the row stays cancelling and the same call can be repeated.
func (s *appointmentService) Cancel(
	ctx context.Context,
	userID uuid.UUID,
	id uuid.UUID,
) (*model.Appointment, error) {
	appointment, err := s.appointments.FindByID(ctx, userID, id)
	if err != nil {
		return nil, translateRepositoryError(err)
	}
	if appointment.Status == model.StatusCancelled {
		return appointment, nil
	}

	if err := s.appointments.MarkCancelling(ctx, appointment); err != nil {
		return nil, err
	}

	err = s.calendar.DeleteEvent(ctx, appointment.UserID, appointment.CalendarID, googleEventID(appointment.ID))
	if err != nil {
		return nil, err
	}

	if err := s.appointments.MarkCancelled(ctx, appointment); err != nil {
		return nil, err
	}

	return appointment, nil
}

func (s *appointmentService) validate(input CreateInput, duration time.Duration) error {
	if input.Timezone == "" {
		return ErrInvalidTimezone
	}
	if _, err := time.LoadLocation(input.Timezone); err != nil {
		return ErrInvalidTimezone
	}
	if !s.hours.ValidDuration(duration) {
		return ErrInvalidDuration
	}
	if !input.StartAt.After(time.Now()) {
		return ErrStartInPast
	}
	if !s.hours.IsBookable(input.StartAt, duration) {
		return ErrOutsideWorkingHours
	}

	return nil
}

// replay handles a request whose idempotency key was already used.
func (s *appointmentService) replay(
	ctx context.Context,
	existing *model.Appointment,
	input CreateInput,
) (*model.Appointment, bool, error) {
	if !matchesRequest(existing, input) {
		return nil, false, ErrIdempotencyConflict
	}

	if existing.Status == model.StatusPending {
		if err := s.syncCreate(ctx, existing); err != nil {
			return nil, false, err
		}
	}

	return existing, false, nil
}

// replayConcurrent handles two requests with the same key racing on the insert.
func (s *appointmentService) replayConcurrent(
	ctx context.Context,
	input CreateInput,
) (*model.Appointment, bool, error) {
	existing, err := s.appointments.FindByIdempotencyKey(ctx, input.UserID, input.IdempotencyKey)
	if err != nil {
		return nil, false, translateRepositoryError(err)
	}

	return s.replay(ctx, existing, input)
}

// syncCreate writes the Google event and confirms the appointment.
// The event ID is derived from the appointment ID, so repeating this step never duplicates the event.
func (s *appointmentService) syncCreate(ctx context.Context, appointment *model.Appointment) error {
	eventID := googleEventID(appointment.ID)

	err := s.calendar.CreateEvent(ctx, appointment.UserID, appointment.CalendarID, CalendarEvent{
		ID:          eventID,
		Title:       appointment.Title,
		Description: fmt.Sprintf("Customer: %s <%s>", appointment.CustomerName, appointment.CustomerEmail),
		Start:       appointment.StartAt,
		End:         appointment.EndAt,
		TimeZone:    appointment.Timezone,
	})
	if err != nil {
		return err
	}

	return s.appointments.MarkConfirmed(ctx, appointment, eventID)
}

// translateRepositoryError maps repository errors to appointments domain errors.
// Other errors pass through unchanged.
func translateRepositoryError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return ErrAppointmentNotFound
	case errors.Is(err, repository.ErrTimeConflict):
		return ErrSlotUnavailable
	default:
		return err
	}
}

func toRanges(appointments []model.Appointment) []TimeRange {
	ranges := make([]TimeRange, 0, len(appointments))
	for _, appointment := range appointments {
		ranges = append(ranges, TimeRange{Start: appointment.StartAt, End: appointment.EndAt})
	}

	return ranges
}

func newAppointment(input CreateInput, calendarID string, duration time.Duration) *model.Appointment {
	return &model.Appointment{
		ID:             uuid.New(),
		UserID:         input.UserID,
		IdempotencyKey: input.IdempotencyKey,
		CalendarID:     calendarID,
		Title:          input.Title,
		CustomerName:   input.CustomerName,
		CustomerEmail:  input.CustomerEmail,
		StartAt:        input.StartAt.UTC(),
		EndAt:          input.StartAt.Add(duration).UTC(),
		Timezone:       input.Timezone,
		Status:         model.StatusPending,
	}
}

func matchesRequest(appointment *model.Appointment, input CreateInput) bool {
	return appointment.Title == input.Title &&
		appointment.CustomerName == input.CustomerName &&
		appointment.CustomerEmail == input.CustomerEmail &&
		appointment.StartAt.Equal(input.StartAt) &&
		appointment.EndAt.Sub(appointment.StartAt) == time.Duration(input.DurationMinutes)*time.Minute &&
		appointment.Timezone == input.Timezone
}

// googleEventID is the appointment UUID without dashes, which Google accepts as an event ID.
func googleEventID(id uuid.UUID) string {
	return strings.ReplaceAll(id.String(), "-", "")
}
