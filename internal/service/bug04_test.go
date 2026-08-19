package service

import (
	"errors"
	"path/filepath"
	"testing"

	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

func TestBug04_ErrorWrappingBreaksErrorsIs(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer s.Close()

	ss := store.NewSensorStore(s)
	as := store.NewAlertStore(s)
	svc := NewAlertService(as, ss)

	alert := &model.Alert{
		SensorID: "sensor-1", AreaID: "area-1",
		Level: model.AlertLevelWarning, Title: "test",
		Value: 100, Threshold: 50,
	}
	err = svc.Create(alert)
	if err == nil {
		t.Fatal("expected error for value > threshold")
	}
	if !errors.Is(err, ErrThresholdExceeded) {
		t.Errorf("expected errors.Is(err, ErrThresholdExceeded) to be true, got: %v", err)
	}
}
