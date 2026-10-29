package service

import "time"

// CalendarEvent is the event written to the booking calendar for an appointment.
type CalendarEvent struct {
	ID          string
	Title       string
	Description string
	Start       time.Time
	End         time.Time
	TimeZone    string
}