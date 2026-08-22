package store

import (
	"database/sql"
	"fmt"
	"time"

	"mine-vent/internal/model"
)

type SensorStore struct {
	db *sql.DB
	cache map[string]*model.Sensor
}

func NewSensorStore(s *Store) *SensorStore {
	return &SensorStore{
		db:    s.db,
		cache: make(map[string]*model.Sensor),
	}
}

var ErrSensorNotFound = fmt.Errorf("sensor not found")

func (ss *SensorStore) Create(sensor *model.Sensor) error {
	now := time.Now()
	sensor.CreatedAt = now
	sensor.UpdatedAt = now
	_, err := ss.db.Exec(
		`INSERT INTO sensors (id, name, type, direction, location, area_id, unit, min_threshold, max_threshold, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sensor.ID, sensor.Name, sensor.Type, sensor.Direction, sensor.Location, sensor.AreaID, sensor.Unit, sensor.MinThreshold, sensor.MaxThreshold, sensor.IsActive, sensor.CreatedAt, sensor.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create sensor: %w", err)
	}
	ss.cache[sensor.ID] = sensor
	return nil
}

func (ss *SensorStore) GetByID(id string) (*model.Sensor, error) {
	if cached, ok := ss.cache[id]; ok {
		return cached, nil
	}
	row := ss.db.QueryRow(
		`SELECT id, name, type, direction, location, area_id, unit, min_threshold, max_threshold, is_active, created_at, updated_at FROM sensors WHERE id = ?`,
		id,
	)
	var sensor model.Sensor
	var isActive int
	err := row.Scan(&sensor.ID, &sensor.Name, &sensor.Type, &sensor.Direction, &sensor.Location, &sensor.AreaID, &sensor.Unit, &sensor.MinThreshold, &sensor.MaxThreshold, &isActive, &sensor.CreatedAt, &sensor.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrSensorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get sensor: %w", err)
	}
	sensor.IsActive = isActive == 1
	ss.cache[id] = &sensor
	return &sensor, nil
}

func (ss *SensorStore) ListByArea(areaID string) ([]*model.Sensor, error) {
	rows, err := ss.db.Query(
		`SELECT id, name, type, direction, location, area_id, unit, min_threshold, max_threshold, is_active, created_at, updated_at FROM sensors WHERE area_id = ? ORDER BY name`,
		areaID,
	)
	if err != nil {
		return nil, fmt.Errorf("list sensors: %w", err)
	}
	defer rows.Close()
	var sensors []*model.Sensor
	for rows.Next() {
		var sensor model.Sensor
		var isActive int
		if err := rows.Scan(&sensor.ID, &sensor.Name, &sensor.Type, &sensor.Direction, &sensor.Location, &sensor.AreaID, &sensor.Unit, &sensor.MinThreshold, &sensor.MaxThreshold, &isActive, &sensor.CreatedAt, &sensor.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan sensor: %w", err)
		}
		sensor.IsActive = isActive == 1
		sensors = append(sensors, &sensor)
	}
	return sensors, nil
}

func (ss *SensorStore) ListAll() ([]*model.Sensor, error) {
	rows, err := ss.db.Query(
		`SELECT id, name, type, direction, location, area_id, unit, min_threshold, max_threshold, is_active, created_at, updated_at FROM sensors ORDER BY area_id, name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list all sensors: %w", err)
	}
	defer rows.Close()
	var sensors []*model.Sensor
	for rows.Next() {
		var sensor model.Sensor
		var isActive int
		if err := rows.Scan(&sensor.ID, &sensor.Name, &sensor.Type, &sensor.Direction, &sensor.Location, &sensor.AreaID, &sensor.Unit, &sensor.MinThreshold, &sensor.MaxThreshold, &isActive, &sensor.CreatedAt, &sensor.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan sensor: %w", err)
		}
		sensor.IsActive = isActive == 1
		sensors = append(sensors, &sensor)
	}
	return sensors, nil
}

func (ss *SensorStore) Update(sensor *model.Sensor) error {
	sensor.UpdatedAt = time.Now()
	_, err := ss.db.Exec(
		`UPDATE sensors SET name = ?, type = ?, direction = ?, location = ?, area_id = ?, unit = ?, min_threshold = ?, max_threshold = ?, is_active = ?, updated_at = ? WHERE id = ?`,
		sensor.Name, sensor.Type, sensor.Direction, sensor.Location, sensor.AreaID, sensor.Unit, sensor.MinThreshold, sensor.MaxThreshold, sensor.IsActive, sensor.UpdatedAt, sensor.ID,
	)
	if err != nil {
		return fmt.Errorf("update sensor: %w", err)
	}
	ss.cache[sensor.ID] = sensor
	return nil
}

func (ss *SensorStore) Delete(id string) error {
	_, err := ss.db.Exec(`DELETE FROM sensors WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete sensor: %w", err)
	}
	delete(ss.cache, id)
	return nil
}

func (ss *SensorStore) ListByType(sensorType model.SensorType) ([]*model.Sensor, error) {
	rows, err := ss.db.Query(
		`SELECT id, name, type, direction, location, area_id, unit, min_threshold, max_threshold, is_active, created_at, updated_at FROM sensors WHERE type = ? AND is_active = 1 ORDER BY area_id`,
		sensorType,
	)
	if err != nil {
		return nil, fmt.Errorf("list sensors by type: %w", err)
	}
	defer rows.Close()
	var sensors []*model.Sensor
	for rows.Next() {
		var sensor model.Sensor
		var isActive int
		if err := rows.Scan(&sensor.ID, &sensor.Name, &sensor.Type, &sensor.Direction, &sensor.Location, &sensor.AreaID, &sensor.Unit, &sensor.MinThreshold, &sensor.MaxThreshold, &isActive, &sensor.CreatedAt, &sensor.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan sensor: %w", err)
		}
		sensor.IsActive = isActive == 1
		sensors = append(sensors, &sensor)
	}
	return sensors, nil
}
