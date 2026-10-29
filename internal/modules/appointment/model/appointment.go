package model

import (
	"time"

	"github.com/google/uuid"
)

// AppointmentStatus is the lifecycle and Google sync state of an appointment.
type AppointmentStatus string

const (
	StatusPending    AppointmentStatus = "pending"    // saved locally, Google not confirmed yet
	StatusConfirmed  AppointmentStatus = "confirmed"  // Google event exists
	StatusCancelling AppointmentStatus = "cancelling" // cancel requested, Google delete not confirmed yet
	StatusCancelled  AppointmentStatus = "cancelled"
)

// Appointment is a booking on the user's selected Google calendar.
// Times are stored as UTC instants; Timezone is the IANA zone the user booked in.
type Appointment struct {
	ID             uuid.UUID         `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex:idx_appointments_user_idempotency,priority:1"`
	IdempotencyKey string            `gorm:"not null;uniqueIndex:idx_appointments_user_idempotency,priority:2"`
	CalendarID     string            `gorm:"not null"`
	Title          string            `gorm:"not null"`
	CustomerName   string            `gorm:"not null"`
	CustomerEmail  string            `gorm:"not null"`
	StartAt        time.Time         `gorm:"not null"`
	EndAt          time.Time         `gorm:"not null"`
	Timezone       string            `gorm:"not null"`
	Status         AppointmentStatus `gorm:"type:text;not null"`
	GoogleEventID  *string
	CancelledAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
} 
