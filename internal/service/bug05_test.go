package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"mine-vent/internal/cache"
	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

func TestMV05_BatchIngestIgnoresContextCancellation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer s.Close()

	ss := store.NewSensorStore(s)
	rs := store.NewReadingStore(s)
	rc := cache.NewReadingCache()
	svc := NewReadingService(rs, ss, rc)

	ss.Create(&model.Sensor{
		ID: "sensor-1", Name: "Test", Type: model.SensorTypeGasCH4,
		Direction: model.DirectionIntake, Location: "loc", AreaID: "area-1",
		Unit: "ppm", MinThreshold: 0, MaxThreshold: 100, IsActive: true,
	})

	readings := make([]model.Reading, 100)
	for i := range readings {
		readings[i] = model.Reading{
			SensorID: "sensor-1", Value: float64(i), Timestamp: time.Now(),
		}
	}
	batch := &model.ReadingBatch{
		BatchID: "batch-1", Readings: readings, Source: "test",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	processed, err := svc.BatchIngest(ctx, batch)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
	if processed > 0 {
		t.Errorf("expected 0 processed with cancelled context, got %d", processed)
	}
}
