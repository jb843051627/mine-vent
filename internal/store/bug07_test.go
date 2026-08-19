package store

import (
	"path/filepath"
	"testing"
	"time"

	"mine-vent/internal/model"
)

func TestBug07_BatchCreatePartialCommitOnFailure(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer s.Close()

	ms := NewMaintenanceStore(s)

	items := []*model.Maintenance{
		{
			Type: model.MaintenanceTypeRoutine, Status: model.MaintenanceStatusScheduled,
			FanID: "fan-1", AreaID: "area-1", Technician: "tech-1",
			ScheduledAt: time.Now().Add(24 * time.Hour), Priority: 5,
			Description: "valid",
		},
		{
			Type: model.MaintenanceTypeRoutine, Status: model.MaintenanceStatusScheduled,
			FanID: "", AreaID: "area-1", Technician: "tech-1",
			ScheduledAt: time.Now().Add(24 * time.Hour), Priority: 5,
			Description: "invalid - empty fan_id",
		},
	}

	err = ms.BatchCreate(items)
	// On bug version: continue skips the error, valid items get committed
	// On fix version: return on error, tx.Rollback() is called via defer
	scheduled, _ := ms.ListScheduled()
	if len(scheduled) > 0 {
		t.Errorf("batch create should rollback all on failure, but found %d committed items", len(scheduled))
	}
}
