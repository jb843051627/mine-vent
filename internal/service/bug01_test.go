package service

import (
	"path/filepath"
	"testing"
	"time"

	"mine-vent/internal/cache"
	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

func TestBug01_RecordReadingNilSensorPanics(t *testing.T) {
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

	reading := &model.Reading{
		SensorID: "nonexistent",
		Value:    42.5,
		Timestamp: time.Now(),
	}
	err = svc.RecordReading(reading)
	if err == nil {
		t.Fatal("expected error for non-existent sensor, got nil")
	}
}
