package service

import (
	"path/filepath"
	"testing"
	"time"

	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

func TestMV10_GetNextMaintenanceWindowMissingFallback(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer s.Close()

	ss := store.NewScheduleStore(s)
	fs := store.NewFanStore(s)
	ms := store.NewMaintenanceStore(s)
	svc := NewScheduleService(ss, fs, ms)

	installedAt := time.Now().Add(-20 * 24 * time.Hour)
	fs.Create(&model.Fan{
		ID: "fan-1", Name: "Fan 1", AreaID: "area-1",
		Capacity: 5000, Status: model.FanStatusRunning,
		PowerKW: 75, InstalledAt: installedAt, LastServiceAt: installedAt,
		IsActive: true, CurrentRPM: 3000,
	})

	window, err := svc.GetNextMaintenanceWindow("fan-1")
	if err != nil {
		t.Fatalf("GetNextMaintenanceWindow: %v", err)
	}

	expectedStart := installedAt.Add(30 * 24 * time.Hour)
	diff := window.StartTime.Sub(expectedStart)
	if diff < -2*time.Hour || diff > 2*time.Hour {
		t.Errorf("next maintenance window should be ~10 days from now, got diff=%v (window=%v, expected=%v)",
			diff, window.StartTime, expectedStart)
	}
}
