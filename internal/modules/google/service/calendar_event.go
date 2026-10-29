package service

import "time"

// CalendarEvent is the data needed to create an event in a Google calendar.
type CalendarEvent struct {
	ID          string // caller-chosen, makes creation idempotent
	Title       string
	Description string
	Start       time.Time
	End         time.Time
	TimeZone    string // IANA zone sent to Google
}