package service

import (
	"path/filepath"
	"testing"
	"time"

	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

func TestMV09_ActivateScheduleErrorShadowing(t *testing.T) {
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

	fs.Create(&model.Fan{
		ID: "fan-1", Name: "Fan 1", AreaID: "area-1",
		Capacity: 5000, Status: model.FanStatusStopped,
		PowerKW: 75, InstalledAt: time.Now(), LastServiceAt: time.Now(),
		IsActive: true,
	})

	schedule := &model.Schedule{
		ID: "sch-1", Name: "Test", Type: model.ScheduleTypeContinuous,
		FanID: "fan-1", AreaID: "area-1",
		StartTime: time.Time{},
		EndTime: time.Now().Add(24 * time.Hour),
		TargetRPM: 1000, Priority: 5, CreatedBy: "test",
	}
	ss.Create(schedule)
	schedule.Status = model.ScheduleStatusDraft
	ss.Update(schedule)

	err = svc.ActivateSchedule("sch-1")
	if err == nil {
		t.Fatal("expected validation error for zero start time, got nil - error was shadowed")
	}

	updated, _ := ss.GetByID("sch-1")
	if updated.Status == model.ScheduleStatusActive {
		t.Error("schedule should not be activated when validation fails")
	}
}
