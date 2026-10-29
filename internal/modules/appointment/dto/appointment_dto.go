package dto

import (
	"time"

	"calendar-booking/internal/modules/appointment/model"
	"calendar-booking/internal/modules/appointment/service"
)

// CreateAppointmentRequest is the body of POST /api/appointments.
type CreateAppointmentRequest struct {
	Title           string    `json:"title" binding:"required,max=200"`
	CustomerName    string    `json:"customer_name" binding:"required,max=200"`
	CustomerEmail   string    `json:"customer_email" binding:"required,email,max=254"`
	StartAt         time.Time `json:"start_at" binding:"required"`
	DurationMinutes int       `json:"duration_minutes" binding:"required,min=1"`
	Timezone        string    `json:"timezone" binding:"required,max=64"`
}

// SlotsQuery is the query string of GET /api/appointments/slots.
type SlotsQuery struct {
	Date            string `form:"date" binding:"required"`
	DurationMinutes int    `form:"duration_minutes" binding:"required,min=1"`
}

// ListQuery is the query string of GET /api/appointments.
type ListQuery struct {
	From  time.Time `form:"from" time_format:"2006-01-02T15:04:05Z07:00"`
	To    time.Time `form:"to" time_format:"2006-01-02T15:04:05Z07:00"`
	Limit int       `form:"limit" binding:"omitempty,min=1,max=100"`
}

// AppointmentResponse is the public representation of an appointment.
type AppointmentResponse struct {
	ID            string     `json:"id"`
	CalendarID    string     `json:"calendar_id"`
	Title         string     `json:"title"`
	CustomerName  string     `json:"customer_name"`
	CustomerEmail string     `json:"customer_email"`
	StartAt       time.Time  `json:"start_at"`
	EndAt         time.Time  `json:"end_at"`
	Timezone      string     `json:"timezone"`
	Status        string     `json:"status"`
	GoogleEventID *string    `json:"google_event_id,omitempty"`
	CancelledAt   *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// SlotResponse is one free slot, in UTC.
type SlotResponse struct {
	StartAt time.Time `json:"start_at"`
	EndAt   time.Time `json:"end_at"`
}

// NewAppointmentResponse maps an appointment to its public representation.
func NewAppointmentResponse(a *model.Appointment) AppointmentResponse {
	return AppointmentResponse{
		ID:            a.ID.String(),
		CalendarID:    a.CalendarID,
		Title:         a.Title,
		CustomerName:  a.CustomerName,
		CustomerEmail: a.CustomerEmail,
		StartAt:       a.StartAt.UTC(),
		EndAt:         a.EndAt.UTC(),
		Timezone:      a.Timezone,
		Status:        string(a.Status),
		GoogleEventID: a.GoogleEventID,
		CancelledAt:   a.CancelledAt,
		CreatedAt:     a.CreatedAt,
	}
}

// NewAppointmentResponses maps a list of appointments.
func NewAppointmentResponses(appointments []model.Appointment) []AppointmentResponse {
	items := make([]AppointmentResponse, 0, len(appointments))
	for i := range appointments {
		items = append(items, NewAppointmentResponse(&appointments[i]))
	}

	return items
}

// NewSlotResponses maps free slots to their public representation.
func NewSlotResponses(slots []service.TimeRange) []SlotResponse {
	items := make([]SlotResponse, 0, len(slots))
	for _, slot := range slots {
		items = append(items, SlotResponse{StartAt: slot.Start.UTC(), EndAt: slot.End.UTC()})
	}

	return items
}
