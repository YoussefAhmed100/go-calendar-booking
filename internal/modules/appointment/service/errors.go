package service

import "errors"

// Domain errors for the appointments module.
// Google provider errors are translated by the calendar gateway adapter,

var (
    // Input validation errors.
    ErrInvalidTimezone     = errors.New("invalid IANA timezone")
    ErrInvalidDuration     = errors.New("duration is outside the allowed range")
    ErrInvalidDate         = errors.New("invalid date, expected YYYY-MM-DD")
    ErrStartInPast         = errors.New("appointment start is in the past")
    ErrNonexistentTime     = errors.New("local time does not exist in the timezone")
    ErrOutsideWorkingHours = errors.New("appointment is outside working hours")
    ErrIdempotencyConflict = errors.New("idempotency key was used with different data")

    // Appointment state errors.
    ErrAppointmentNotFound = errors.New("appointment not found")
    ErrSlotUnavailable     = errors.New("the requested time slot is not available")

    // Calendar integration errors.
    ErrNoCalendarSelected           = errors.New("no calendar selected")
    ErrCalendarNotConnected         = errors.New("google account not connected")
    ErrCalendarAuthorizationRevoked = errors.New("google authorization expired or revoked")
    ErrCalendarSyncFailed           = errors.New("calendar provider request failed")
)