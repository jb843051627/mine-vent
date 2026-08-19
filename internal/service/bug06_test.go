package service

import (
	"path/filepath"
	"testing"
	"time"

	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

func TestMV06_AssignFanNilPanic(t *testing.T) {
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

	schedule := &model.Schedule{
		ID: "sch-1", Name: "Test", Type: model.ScheduleTypeContinuous,
		FanID: "fan-1", AreaID: "area-1",
		StartTime: time.Now(), EndTime: time.Now().Add(24 * time.Hour),
		TargetRPM: 1000, Priority: 5, CreatedBy: "test",
	}
	ss.Create(schedule)

	err = svc.AssignFan("sch-1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent fan, got nil")
	}
}
