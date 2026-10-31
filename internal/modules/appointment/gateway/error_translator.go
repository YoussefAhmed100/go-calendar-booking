package gateway

import (
	"errors"
	"fmt"

	"calendar-booking/internal/modules/appointment/service"
	googleservice "calendar-booking/internal/modules/google/service"
)
// translateError maps Google module errors to appointments domain errors,
// keeping the original error for logs. Other errors pass through unchanged.
func translateError(err error) error {
	switch {
	case err == nil:
		return nil

	case errors.Is(err, googleservice.ErrGoogleAccountNotConnected):
		return fmt.Errorf("%w: %w", service.ErrCalendarNotConnected, err)

	case errors.Is(err, googleservice.ErrGoogleAuthorizationRevoked):
		return fmt.Errorf("%w: %w", service.ErrCalendarAuthorizationRevoked, err)

	case errors.Is(err, googleservice.ErrNoCalendarSelected):
		return fmt.Errorf("%w: %w", service.ErrNoCalendarSelected, err)

	case errors.Is(err, googleservice.ErrGoogleAPIFailure):
		return fmt.Errorf("%w: %w", service.ErrCalendarSyncFailed, err)

	default:
		return err
	}
}