package store

import (
	"database/sql"
	"fmt"
	"time"

	"mine-vent/internal/model"
)

type ReadingStore struct {
	db *sql.DB
}

func NewReadingStore(s *Store) *ReadingStore {
	return &ReadingStore{db: s.db}
}

func (rs *ReadingStore) Create(reading *model.Reading) error {
	reading.ID = fmt.Sprintf("rdg-%d", time.Now().UnixNano())
	reading.RecordedAt = time.Now()
	_, err := rs.db.Exec(
		`INSERT INTO readings (id, sensor_id, value, unit, timestamp, recorded_at, quality) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		reading.ID, reading.SensorID, reading.Value, reading.Unit, reading.Timestamp, reading.RecordedAt, reading.Quality,
	)
	if err != nil {
		return fmt.Errorf("create reading: %w", err)
	}
	return nil
}

func (rs *ReadingStore) GetByID(id string) (*model.Reading, error) {
	row := rs.db.QueryRow(
		`SELECT id, sensor_id, value, unit, timestamp, recorded_at, quality FROM readings WHERE id = ?`,
		id,
	)
	var r model.Reading
	err := row.Scan(&r.ID, &r.SensorID, &r.Value, &r.Unit, &r.Timestamp, &r.RecordedAt, &r.Quality)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("reading not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get reading: %w", err)
	}
	return &r, nil
}

func (rs *ReadingStore) ListBySensor(sensorID string, limit int) ([]*model.Reading, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := rs.db.Query(
		`SELECT id, sensor_id, value, unit, timestamp, recorded_at, quality FROM readings WHERE sensor_id = ? ORDER BY timestamp DESC LIMIT ?`,
		sensorID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list readings: %w", err)
	}
	defer rows.Close()
	var readings []*model.Reading
	for rows.Next() {
		var r model.Reading
		if err := rows.Scan(&r.ID, &r.SensorID, &r.Value, &r.Unit, &r.Timestamp, &r.RecordedAt, &r.Quality); err != nil {
			return nil, fmt.Errorf("scan reading: %w", err)
		}
		readings = append(readings, &r)
	}
	return readings, nil
}

func (rs *ReadingStore) ListByArea(areaID string, limit int) ([]*model.Reading, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := rs.db.Query(
		`SELECT r.id, r.sensor_id, r.value, r.unit, r.timestamp, r.recorded_at, r.quality FROM readings r INNER JOIN sensors s ON r.sensor_id = s.id WHERE s.area_id = ? ORDER BY r.timestamp DESC LIMIT ?`,
		areaID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list readings by area: %w", err)
	}
	defer rows.Close()
	var readings []*model.Reading
	for rows.Next() {
		var r model.Reading
		if err := rows.Scan(&r.ID, &r.SensorID, &r.Value, &r.Unit, &r.Timestamp, &r.RecordedAt, &r.Quality); err != nil {
			return nil, fmt.Errorf("scan reading: %w", err)
		}
		readings = append(readings, &r)
	}
	return readings, nil
}

func (rs *ReadingStore) ListByTimeRange(start, end time.Time, limit int) ([]*model.Reading, error) {
	if limit <= 0 {
		limit = 1000
	}
	rows, err := rs.db.Query(
		`SELECT id, sensor_id, value, unit, timestamp, recorded_at, quality FROM readings WHERE timestamp >= ? AND timestamp <= ? ORDER BY timestamp DESC LIMIT ?`,
		start, end, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list readings by time: %w", err)
	}
	defer rows.Close()
	var readings []*model.Reading
	for rows.Next() {
		var r model.Reading
		if err := rows.Scan(&r.ID, &r.SensorID, &r.Value, &r.Unit, &r.Timestamp, &r.RecordedAt, &r.Quality); err != nil {
			return nil, fmt.Errorf("scan reading: %w", err)
		}
		readings = append(readings, &r)
	}
	return readings, nil
}

func (rs *ReadingStore) AggregateBySensor(sensorID string, startTime, endTime time.Time) (*model.AggregatedReading, error) {
	row := rs.db.QueryRow(
		`SELECT COUNT(*), MIN(value), MAX(value), AVG(value), (SELECT value FROM readings WHERE sensor_id = ? AND timestamp >= ? AND timestamp <= ? ORDER BY timestamp DESC LIMIT 1) FROM readings WHERE sensor_id = ? AND timestamp >= ? AND timestamp <= ?`,
		sensorID, startTime, endTime, sensorID, startTime, endTime,
	)
	var agg model.AggregatedReading
	agg.SensorID = sensorID
	err := row.Scan(&agg.Count, &agg.Min, &agg.Max, &agg.Avg, &agg.Latest)
	if err != nil {
		return nil, fmt.Errorf("aggregate readings: %w", err)
	}
	return &agg, nil
}

func (rs *ReadingStore) DeleteBySensor(sensorID string) error {
	_, err := rs.db.Exec(`DELETE FROM readings WHERE sensor_id = ?`, sensorID)
	if err != nil {
		return fmt.Errorf("delete readings: %w", err)
	}
	return nil
}

func (rs *ReadingStore) CountBySensor(sensorID string) (int, error) {
	var count int
	err := rs.db.QueryRow(`SELECT COUNT(*) FROM readings WHERE sensor_id = ?`, sensorID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count readings: %w", err)
	}
	return count, nil
}

func (rs *ReadingStore) GetLatestBySensor(sensorID string) (*model.Reading, error) {
	row := rs.db.QueryRow(
		`SELECT id, sensor_id, value, unit, timestamp, recorded_at, quality FROM readings WHERE sensor_id = ? ORDER BY timestamp DESC LIMIT 1`,
		sensorID,
	)
	var r model.Reading
	err := row.Scan(&r.ID, &r.SensorID, &r.Value, &r.Unit, &r.Timestamp, &r.RecordedAt, &r.Quality)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no readings for sensor: %s", sensorID)
	}
	if err != nil {
		return nil, fmt.Errorf("get latest reading: %w", err)
	}
	return &r, nil
}
