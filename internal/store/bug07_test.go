package store

import (
	"path/filepath"
	"testing"
	"time"

	"mine-vent/internal/model"
)

func TestMV07_BatchCreatePartialCommitOnFailure(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer s.Close()

	ms := NewMaintenanceStore(s)

	// Add a unique constraint to force INSERT failure for duplicate fan_id + scheduled_at
	_, err = s.DB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_test_unique ON maintenances(fan_id, scheduled_at)")
	if err != nil {
		t.Fatalf("create unique index: %v", err)
	}

	scheduledAt := time.Now().Add(24 * time.Hour)
	items := []*model.Maintenance{
		{
			Type: model.MaintenanceTypeRoutine, Status: model.MaintenanceStatusScheduled,
			FanID: "fan-1", AreaID: "area-1", Technician: "tech-1",
			ScheduledAt: scheduledAt, Priority: 5,
			Description: "first item",
		},
		{
			Type: model.MaintenanceTypeRoutine, Status: model.MaintenanceStatusScheduled,
			FanID: "fan-1", AreaID: "area-1", Technician: "tech-1",
			ScheduledAt: scheduledAt, // same fan_id + scheduled_at -> UNIQUE violation
			Priority: 5,
			Description: "duplicate item",
		},
	}

	err = ms.BatchCreate(items)
	// On bug version: continue skips the error, first item gets committed
	// On fix version: return on error, tx.Rollback() rolls back everything
	scheduled, _ := ms.ListScheduled()
	if len(scheduled) > 0 {
		t.Errorf("batch create should rollback all on failure, but found %d committed items", len(scheduled))
	}
}
