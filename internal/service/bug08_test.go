package service

import (
	"path/filepath"
	"testing"
	"time"

	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

func TestMV08_GenerateHealthReportPollutesStore(t *testing.T) {
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
	_ = svc

	fs.Create(&model.Fan{
		ID: "fan-1", Name: "Fan 1", AreaID: "area-1",
		Capacity: 5000, Status: model.FanStatusRunning,
		PowerKW: 75, InstalledAt: time.Now(), LastServiceAt: time.Now(),
		IsActive: true, CurrentRPM: 3000,
	})

	// Get fan from store (returns cached pointer on bug version)
	fan, _ := fs.GetByID("fan-1")
	originalStatus := fan.Status

	// Modify the returned fan's status
	fan.Status = model.FanStatusFault

	// Get fan again - should still have original status
	fan2, _ := fs.GetByID("fan-1")
	if fan2.Status != originalStatus {
		t.Errorf("cache polluted: fan status was %v, now %v",
			originalStatus, fan2.Status)
	}
}
