package service

import (
	"path/filepath"
	"testing"
	"time"

	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

func TestBug02_CheckThresholdsPollutesCache(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer s.Close()

	ss := store.NewSensorStore(s)
	as := store.NewAlertStore(s)

	sensor := &model.Sensor{
		ID: "sensor-a", Name: "Test", Type: model.SensorTypeGasCH4,
		Direction: model.DirectionIntake, Location: "loc", AreaID: "area-1",
		Unit: "ppm", MinThreshold: 0, MaxThreshold: 50, IsActive: true,
	}
	ss.Create(sensor)

	now := time.Now()
	for i := 2; i >= 0; i-- {
		as.Create(&model.Alert{
			SensorID: "sensor-a", AreaID: "area-1",
			Level: model.AlertLevelWarning, Status: model.AlertStatusActive,
			Title: "test", Message: "test", Value: 10, Threshold: 50,
			TriggeredAt: now.Add(time.Duration(i) * time.Hour),
		})
	}

	before, _ := as.ListByArea("area-1")
	if len(before) != 3 {
		t.Fatalf("expected 3 alerts, got %d", len(before))
	}
	beforeFirst := before[0].TriggeredAt

	// Reverse the order of the returned slice in-place
	for i, j := 0, len(before)-1; i < j; i, j = i+1, j-1 {
		before[i], before[j] = before[j], before[i]
	}

	after, _ := as.ListByArea("area-1")
	if len(after) != 3 {
		t.Fatalf("expected 3 alerts after, got %d", len(after))
	}
	if !after[0].TriggeredAt.Equal(beforeFirst) {
		t.Errorf("cache polluted: first alert was %v, now %v",
			beforeFirst, after[0].TriggeredAt)
	}
}
