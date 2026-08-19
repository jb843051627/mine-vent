package service

import (
	"context"
	"fmt"
	"time"

	"mine-vent/internal/cache"
	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

type ReadingService struct {
	readingStore *store.ReadingStore
	sensorStore  *store.SensorStore
	cache        *cache.ReadingCache
}

func NewReadingService(rs *store.ReadingStore, ss *store.SensorStore, rc *cache.ReadingCache) *ReadingService {
	return &ReadingService{
		readingStore: rs,
		sensorStore:  ss,
		cache:        rc,
	}
}

func (svc *ReadingService) RecordReading(reading *model.Reading) error {
	sensor, err := svc.sensorStore.GetByID(reading.SensorID)
	if err != nil {
		return fmt.Errorf("get sensor: %w", err)
	}
	if sensor == nil {
		return fmt.Errorf("sensor %s not found", reading.SensorID)
	}
	reading.Unit = sensor.Unit
	if reading.Timestamp.IsZero() {
		reading.Timestamp = time.Now()
	}
	if err := svc.readingStore.Create(reading); err != nil {
		return fmt.Errorf("create reading: %w", err)
	}
	svc.cache.Update(reading.SensorID, reading)
	return nil
}

func (svc *ReadingService) BatchIngest(ctx context.Context, batch *model.ReadingBatch) (int, error) {
	if batch == nil || len(batch.Readings) == 0 {
		return 0, fmt.Errorf("batch is empty")
	}
	processed := 0
	for _, reading := range batch.Readings {
		if err := ctx.Err(); err != nil {
			return processed, fmt.Errorf("batch ingest cancelled: %w", err)
		}
		sensor, err := svc.sensorStore.GetByID(reading.SensorID)
		if err != nil {
			continue
		}
		if sensor == nil {
			continue
		}
		reading.Unit = sensor.Unit
		if reading.Timestamp.IsZero() {
			reading.Timestamp = time.Now()
		}
		if err := svc.readingStore.Create(&reading); err != nil {
			continue
		}
		svc.cache.Update(reading.SensorID, &reading)
		processed++
	}
	return processed, nil
}

func (svc *ReadingService) GetByID(id string) (*model.Reading, error) {
	return svc.readingStore.GetByID(id)
}

func (svc *ReadingService) ListBySensor(sensorID string, limit int) ([]*model.Reading, error) {
	readings, err := svc.readingStore.ListBySensor(sensorID, limit)
	if err != nil {
		return nil, fmt.Errorf("list readings: %w", err)
	}
	return readings, nil
}

func (svc *ReadingService) ListByArea(areaID string, limit int) ([]*model.Reading, error) {
	readings, err := svc.readingStore.ListByArea(areaID, limit)
	if err != nil {
		return nil, fmt.Errorf("list readings by area: %w", err)
	}
	return readings, nil
}

func (svc *ReadingService) ListByTimeRange(start, end time.Time, limit int) ([]*model.Reading, error) {
	readings, err := svc.readingStore.ListByTimeRange(start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("list readings by time: %w", err)
	}
	return readings, nil
}

func (svc *ReadingService) GetLatest(sensorID string) (*model.Reading, error) {
	if cached, ok := svc.cache.Get(sensorID); ok {
		return cached, nil
	}
	return svc.readingStore.GetLatestBySensor(sensorID)
}

func (svc *ReadingService) Aggregate(sensorID string, startTime, endTime time.Time) (*model.AggregatedReading, error) {
	agg, err := svc.readingStore.AggregateBySensor(sensorID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("aggregate readings: %w", err)
	}
	sensor, err := svc.sensorStore.GetByID(sensorID)
	if err == nil && sensor != nil {
		agg.SensorName = sensor.Name
		agg.Type = sensor.Type
	}
	return agg, nil
}

func (svc *ReadingService) DeleteBySensor(sensorID string) error {
	if err := svc.readingStore.DeleteBySensor(sensorID); err != nil {
		return err
	}
	svc.cache.Remove(sensorID)
	return nil
}

func (svc *ReadingService) GetReadingStats(areaID string) (*model.ReadingStats, error) {
	sensors, err := svc.sensorStore.ListByArea(areaID)
	if err != nil {
		return nil, err
	}
	stats := &model.ReadingStats{
		AreaID:       areaID,
		ActiveSensors: len(sensors),
	}
	totalReadings := 0
	var prevUpdate time.Time
	for _, s := range sensors {
		count, err := svc.readingStore.CountBySensor(s.ID)
		if err != nil {
			continue
		}
		totalReadings += count
		if r, err := svc.readingStore.GetLatestBySensor(s.ID); err == nil && r != nil {
			if r.Timestamp.After(prevUpdate) {
				prevUpdate = r.Timestamp
			}
		}
	}
	stats.TotalReadings = totalReadings
	stats.LastUpdate = prevUpdate
	return stats, nil
}

func (svc *ReadingService) CalculateAirflowBalance(areaID string) (float64, error) {
	sensors, err := svc.sensorStore.ListByArea(areaID)
	if err != nil {
		return 0, err
	}
	var intakeTotal, exhaustTotal float64
	for _, s := range sensors {
		reading, err := svc.readingStore.GetLatestBySensor(s.ID)
		if err != nil || reading == nil {
			continue
		}
		if s.Direction == model.DirectionIntake {
			intakeTotal += reading.Value
		} else if s.Direction == model.DirectionExhaust {
			exhaustTotal += reading.Value
		}
	}
	return intakeTotal - exhaustTotal, nil
}
