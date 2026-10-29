package service

import (
	"fmt"
	"strings"
	"time"

	"calendar-booking/internal/common/config"
)

const maxAppointmentDuration = 8 * time.Hour

var weekdayByName = map[string]time.Weekday{
	"sun": time.Sunday,
	"mon": time.Monday,
	"tue": time.Tuesday,
	"wed": time.Wednesday,
	"thu": time.Thursday,
	"fri": time.Friday,
	"sat": time.Saturday,
}

// TimeRange is a half-open interval [Start, End).
type TimeRange struct {
	Start time.Time
	End   time.Time
}

func (r TimeRange) overlaps(other TimeRange) bool {
	return r.Start.Before(other.End) && other.Start.Before(r.End)
}

type clock struct {
	hour   int
	minute int
}

// WorkingHours defines when appointments can be booked.
type WorkingHours struct {
	location *time.Location
	open     clock
	close    clock
	workDays map[time.Weekday]bool
	step     time.Duration
}

// NewWorkingHours builds the working hours from configuration.
func NewWorkingHours(cfg *config.Config) (WorkingHours, error) {
	location, err := time.LoadLocation(cfg.BookingTimezone)
	if err != nil {
		return WorkingHours{}, fmt.Errorf("invalid BOOKING_TIMEZONE: %w", err)
	}

	open, err := parseClock(cfg.BookingDayStart)
	if err != nil {
		return WorkingHours{}, fmt.Errorf("invalid BOOKING_DAY_START: %w", err)
	}

	closing, err := parseClock(cfg.BookingDayEnd)
	if err != nil {
		return WorkingHours{}, fmt.Errorf("invalid BOOKING_DAY_END: %w", err)
	}
	if closing.minutes() <= open.minutes() {
		return WorkingHours{}, fmt.Errorf("BOOKING_DAY_END must be after BOOKING_DAY_START")
	}

	workDays, err := parseWorkDays(cfg.BookingWorkDays)
	if err != nil {
		return WorkingHours{}, fmt.Errorf("invalid BOOKING_WORK_DAYS: %w", err)
	}

	if cfg.BookingSlotMinutes <= 0 {
		return WorkingHours{}, fmt.Errorf("BOOKING_SLOT_MINUTES must be positive")
	}

	return WorkingHours{
		location: location,
		open:     open,
		close:    closing,
		workDays: workDays,
		step:     time.Duration(cfg.BookingSlotMinutes) * time.Minute,
	}, nil
}

// Location is the timezone the working hours are defined in.
func (w WorkingHours) Location() *time.Location {
	return w.location
}

// ValidDuration reports whether the duration is a positive multiple of the slot step
// and not longer than the maximum.
func (w WorkingHours) ValidDuration(duration time.Duration) bool {
	return duration >= w.step && duration <= maxAppointmentDuration && duration%w.step == 0
}

// FreeSlots returns the slots of the given day that are in the future and do not overlap
// any booked range. The date must be in the working-hours location.
func (w WorkingHours) FreeSlots(
	date time.Time,
	duration time.Duration,
	booked []TimeRange,
	now time.Time,
) []TimeRange {
	slots := make([]TimeRange, 0)
	if !w.workDays[date.Weekday()] {
		return slots
	}

	opensAt, closesAt := w.dayBounds(date)

	for start := opensAt; !start.Add(duration).After(closesAt); start = start.Add(w.step) {
		slot := TimeRange{Start: start, End: start.Add(duration)}
		if slot.Start.After(now) && !overlapsAny(slot, booked) {
			slots = append(slots, slot)
		}
	}

	return slots
}

// IsBookable reports whether [start, start+duration) fits the working hours and the slot grid.
func (w WorkingHours) IsBookable(start time.Time, duration time.Duration) bool {
	local := start.In(w.location)
	if !w.workDays[local.Weekday()] {
		return false
	}

	opensAt, closesAt := w.dayBounds(local)
	end := start.Add(duration)

	onGrid := start.Sub(opensAt)%w.step == 0

	return !start.Before(opensAt) && !end.After(closesAt) && onGrid
}

func (w WorkingHours) dayBounds(day time.Time) (opensAt, closesAt time.Time) {
	year, month, dayOfMonth := day.In(w.location).Date()

	opensAt = time.Date(year, month, dayOfMonth, w.open.hour, w.open.minute, 0, 0, w.location)
	closesAt = time.Date(year, month, dayOfMonth, w.close.hour, w.close.minute, 0, 0, w.location)

	return opensAt, closesAt
}

func overlapsAny(slot TimeRange, booked []TimeRange) bool {
	for _, b := range booked {
		if slot.overlaps(b) {
			return true
		}
	}

	return false
}

func (c clock) minutes() int {
	return c.hour*60 + c.minute
}

func parseClock(value string) (clock, error) {
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return clock{}, fmt.Errorf("expected HH:MM, got %q", value)
	}

	return clock{hour: parsed.Hour(), minute: parsed.Minute()}, nil
}

func parseWorkDays(value string) (map[time.Weekday]bool, error) {
	days := make(map[time.Weekday]bool)

	for _, name := range strings.Split(value, ",") {
		weekday, ok := weekdayByName[strings.ToLower(strings.TrimSpace(name))]
		if !ok {
			return nil, fmt.Errorf("unknown day %q", name)
		}
		days[weekday] = true
	}

	return days, nil
}