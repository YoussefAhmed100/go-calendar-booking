package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"calendar-booking/internal/modules/appointment/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const (
	pgUniqueViolation    = "23505"
	pgExclusionViolation = "23P01"
)

var (
	// ErrNotFound means no appointment matches the query.
	ErrNotFound = errors.New("appointment not found")
	// ErrTimeConflict means the time overlaps an active booking on the same calendar.
	ErrTimeConflict = errors.New("appointment time conflicts with an active booking")
	// ErrDuplicateKey means the user already used the idempotency key.
	ErrDuplicateKey = errors.New("idempotency key already used")
)

// AppointmentRepository reads and writes appointments.
type AppointmentRepository struct {
	db *gorm.DB
}

// NewAppointmentRepository creates a repository backed by the given database.
func NewAppointmentRepository(db *gorm.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

// Create inserts the appointment. It returns ErrTimeConflict or ErrDuplicateKey
// when the database rejects it for those reasons.
func (r *AppointmentRepository) Create(ctx context.Context, appointment *model.Appointment) error {
	err := r.db.WithContext(ctx).Create(appointment).Error

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgExclusionViolation:
			return ErrTimeConflict
		case pgUniqueViolation:
			return ErrDuplicateKey
		}
	}
	if err != nil {
		return fmt.Errorf("save appointment: %w", err)
	}

	return nil
}

// FindByID returns ErrNotFound when the appointment is missing or not owned by the user.
func (r *AppointmentRepository) FindByID(
	ctx context.Context,
	userID uuid.UUID,
	id uuid.UUID,
) (*model.Appointment, error) {
	return r.findOne(ctx, "id = ? AND user_id = ?", id, userID)
}

// FindByIdempotencyKey returns ErrNotFound when the user never used the key.
func (r *AppointmentRepository) FindByIdempotencyKey(
	ctx context.Context,
	userID uuid.UUID,
	key string,
) (*model.Appointment, error) {
	return r.findOne(ctx, "user_id = ? AND idempotency_key = ?", userID, key)
}

// List returns the user's appointments ordered by start time.
// Zero from or to means unbounded on that side.
func (r *AppointmentRepository) List(
	ctx context.Context,
	userID uuid.UUID,
	from time.Time,
	to time.Time,
	limit int,
) ([]model.Appointment, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if !from.IsZero() {
		query = query.Where("end_at > ?", from)
	}
	if !to.IsZero() {
		query = query.Where("start_at < ?", to)
	}

	appointments := make([]model.Appointment, 0)
	if err := query.Order("start_at").Limit(limit).Find(&appointments).Error; err != nil {
		return nil, fmt.Errorf("list appointments: %w", err)
	}

	return appointments, nil
}

// ListActive returns the non-cancelled appointments of a calendar that intersect [from, to).
// Only the time columns are loaded.
func (r *AppointmentRepository) ListActive(
	ctx context.Context,
	calendarID string,
	from time.Time,
	to time.Time,
) ([]model.Appointment, error) {
	var appointments []model.Appointment

	err := r.db.WithContext(ctx).
		Select("start_at", "end_at").
		Where(
			"calendar_id = ? AND status <> ? AND start_at < ? AND end_at > ?",
			calendarID, model.StatusCancelled, to, from,
		).
		Find(&appointments).Error
	if err != nil {
		return nil, fmt.Errorf("list active appointments: %w", err)
	}

	return appointments, nil
}

// MarkConfirmed records the Google event of a pending appointment and updates it in place.
func (r *AppointmentRepository) MarkConfirmed(
	ctx context.Context,
	appointment *model.Appointment,
	eventID string,
) error {
	err := r.db.WithContext(ctx).Model(appointment).
		Where("status = ?", model.StatusPending).
		Updates(map[string]any{"status": model.StatusConfirmed, "google_event_id": eventID}).Error
	if err != nil {
		return fmt.Errorf("confirm appointment: %w", err)
	}

	appointment.Status = model.StatusConfirmed
	appointment.GoogleEventID = &eventID

	return nil
}

// MarkCancelling records that a cancel was requested and updates the appointment in place.
func (r *AppointmentRepository) MarkCancelling(ctx context.Context, appointment *model.Appointment) error {
	err := r.db.WithContext(ctx).Model(appointment).
		Where("status IN ?", []model.AppointmentStatus{model.StatusPending, model.StatusConfirmed}).
		Update("status", model.StatusCancelling).Error
	if err != nil {
		return fmt.Errorf("mark appointment cancelling: %w", err)
	}

	appointment.Status = model.StatusCancelling

	return nil
}

// MarkCancelled finishes a cancel and updates the appointment in place.
func (r *AppointmentRepository) MarkCancelled(ctx context.Context, appointment *model.Appointment) error {
	cancelledAt := time.Now().UTC()

	err := r.db.WithContext(ctx).Model(appointment).
		Updates(map[string]any{"status": model.StatusCancelled, "cancelled_at": cancelledAt}).Error
	if err != nil {
		return fmt.Errorf("cancel appointment: %w", err)
	}

	appointment.Status = model.StatusCancelled
	appointment.CancelledAt = &cancelledAt

	return nil
}

func (r *AppointmentRepository) findOne(
	ctx context.Context,
	query string,
	args ...any,
) (*model.Appointment, error) {
	var appointment model.Appointment

	err := r.db.WithContext(ctx).Where(query, args...).First(&appointment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find appointment: %w", err)
	}

	return &appointment, nil
}