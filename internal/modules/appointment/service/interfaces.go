package service

import (
	"context"
	"time"

	"calendar-booking/internal/modules/appointment/model"
	"calendar-booking/internal/modules/appointment/repository"

	"github.com/google/uuid"
)

// AppointmentService is the use-case contract the controller depends on.
type AppointmentService interface {
	// Slots returns the free slots of a day (YYYY-MM-DD, in the working-hours timezone).
	Slots(ctx context.Context, userID uuid.UUID, date string, durationMinutes int) ([]TimeRange, error)
	// Create books a slot. The bool is false when the request replayed an earlier idempotency key.
	Create(ctx context.Context, input CreateInput) (*model.Appointment, bool, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*model.Appointment, error)
	List(ctx context.Context, userID uuid.UUID, filter ListFilter) ([]model.Appointment, error)
	Cancel(ctx context.Context, userID, id uuid.UUID) (*model.Appointment, error)
}

// AppointmentRepository is the persistence contract the service depends on.
type AppointmentRepository interface {
	Create(ctx context.Context, appointment *model.Appointment) error
	FindByID(ctx context.Context, userID, id uuid.UUID) (*model.Appointment, error)
	FindByIdempotencyKey(ctx context.Context, userID uuid.UUID, key string) (*model.Appointment, error)
	List(ctx context.Context, userID uuid.UUID, from, to time.Time, limit int) ([]model.Appointment, error)
	ListActive(ctx context.Context, calendarID string, from, to time.Time) ([]model.Appointment, error)
	MarkConfirmed(ctx context.Context, appointment *model.Appointment, eventID string) error
	MarkCancelling(ctx context.Context, appointment *model.Appointment) error
	MarkCancelled(ctx context.Context, appointment *model.Appointment) error
}

// CalendarGateway is the part of the Google integration that appointments depend on.
type CalendarGateway interface {
	SelectedCalendarID(ctx context.Context, userID uuid.UUID) (string, error)
	CreateEvent(ctx context.Context, userID uuid.UUID, calendarID string, event CalendarEvent) error
	DeleteEvent(ctx context.Context, userID uuid.UUID, calendarID, eventID string) error
}

// Compile-time check that the repository satisfies its contract.
// It lives here because the repository package cannot import service without a cycle.
var _ AppointmentRepository = (*repository.AppointmentRepository)(nil)