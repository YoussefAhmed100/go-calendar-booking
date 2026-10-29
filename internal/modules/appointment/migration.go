package appointment

import (
	"fmt"

	"calendar-booking/internal/modules/appointment/model"

	"gorm.io/gorm"
)

const noOverlapConstraint = "appointments_no_overlap"

// migrate creates the appointments table and the no-overlap rule,
// which GORM tags cannot express. It is safe to run on every startup.
func migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.Appointment{}); err != nil {
		return fmt.Errorf("migrate appointments: %w", err)
	}

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS btree_gist`).Error; err != nil {
		return fmt.Errorf("enable btree_gist: %w", err)
	}

	var existing int64
	err := db.Raw(`SELECT count(*) FROM pg_constraint WHERE conname = ?`, noOverlapConstraint).
		Scan(&existing).Error
	if err != nil {
		return fmt.Errorf("check no-overlap constraint: %w", err)
	}
	if existing > 0 {
		return nil
	}

	err = db.Exec(`
		ALTER TABLE appointments
		ADD CONSTRAINT ` + noOverlapConstraint + `
		EXCLUDE USING gist (
			calendar_id WITH =,
			tstzrange(start_at, end_at, '[)') WITH &&
		) WHERE (status <> 'cancelled')`).Error
	if err != nil {
		return fmt.Errorf("add no-overlap constraint: %w", err)
	}

	return nil
}
