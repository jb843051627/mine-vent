package service

import (
	"path/filepath"
	"testing"
	"time"

	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

func TestBug08_GenerateHealthReportPollutesStore(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer s.Close()

	fs := store.NewFanStore(s)
	ms := store.NewMaintenanceStore(s)
	rs := store.NewReadingStore(s)
	svc := NewFanService(fs, ms, rs)

	for i, name := range []string{"fan-a", "fan-b", "fan-c"} {
		fs.Create(&model.Fan{
			ID: name, Name: "Fan " + name, AreaID: "area-1",
			Capacity: 5000, Status: model.FanStatusRunning,
			PowerKW: 75, InstalledAt: time.Now(), LastServiceAt: time.Now(),
			IsActive: true, CurrentRPM: i * 100,
		})
	}

	before, _ := fs.ListByArea("area-1")
	if len(before) != 3 {
		t.Fatalf("expected 3 fans, got %d", len(before))
	}
	beforeOrder := make([]string, len(before))
	for i, f := range before {
		beforeOrder[i] = f.ID
	}

	svc.GenerateHealthReport("area-1")

	after, _ := fs.ListByArea("area-1")
	for i := range beforeOrder {
		if i < len(after) && after[i].ID != beforeOrder[i] {
			t.Errorf("fan order changed at index %d: before=%s after=%s",
				i, beforeOrder[i], after[i].ID)
		}
	}
}
