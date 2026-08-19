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
	svc := NewAlertService(as, ss)

	sensor := &model.Sensor{
		ID: "sensor-a", Name: "Test", Type: model.SensorTypeGasCH4,
		Direction: model.DirectionIntake, Location: "loc", AreaID: "area-1",
		Unit: "ppm", MinThreshold: 0, MaxThreshold: 50, IsActive: true,
	}
	ss.Create(sensor)

	now := time.Now()
	for i := 0; i < 3; i++ {
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
	beforeOrder := []time.Time{}
	for _, a := range before {
		beforeOrder = append(beforeOrder, a.TriggeredAt)
	}

	svc.CheckThresholds("area-1")

	after, _ := as.ListByArea("area-1")
	for i := range beforeOrder {
		if i < len(after) && !after[i].TriggeredAt.Equal(beforeOrder[i]) {
			t.Errorf("cache polluted at index %d: before=%v after=%v",
				i, beforeOrder[i], after[i].TriggeredAt)
		}
	}
}
