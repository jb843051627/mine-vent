package store

import (
	"database/sql"
	"fmt"
	"time"

	"mine-vent/internal/model"
)

type FanStore struct {
	db    *sql.DB
	cache map[string]*model.Fan
}

func NewFanStore(s *Store) *FanStore {
	return &FanStore{
		db:    s.db,
		cache: make(map[string]*model.Fan),
	}
}

var ErrFanNotFound = fmt.Errorf("fan not found")

func (fs *FanStore) Create(fan *model.Fan) error {
	_, err := fs.db.Exec(
		`INSERT INTO fans (id, name, area_id, capacity, current_rpm, status, power_kw, installed_at, last_service_at, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fan.ID, fan.Name, fan.AreaID, fan.Capacity, fan.CurrentRPM, fan.Status, fan.PowerKW, fan.InstalledAt, fan.LastServiceAt, fan.IsActive,
	)
	if err != nil {
		return fmt.Errorf("create fan: %w", err)
	}
	fs.cache[fan.ID] = fan
	return nil
}

func (fs *FanStore) GetByID(id string) (*model.Fan, error) {
	if cached, ok := fs.cache[id]; ok {
		return cached, nil
	}
	row := fs.db.QueryRow(
		`SELECT id, name, area_id, capacity, current_rpm, status, power_kw, installed_at, last_service_at, is_active FROM fans WHERE id = ?`,
		id,
	)
	var fan model.Fan
	var isActive int
	err := row.Scan(&fan.ID, &fan.Name, &fan.AreaID, &fan.Capacity, &fan.CurrentRPM, &fan.Status, &fan.PowerKW, &fan.InstalledAt, &fan.LastServiceAt, &isActive)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get fan: %w", err)
	}
	fan.IsActive = isActive == 1
	fs.cache[id] = &fan
	return &fan, nil
}

func (fs *FanStore) ListAll() ([]*model.Fan, error) {
	rows, err := fs.db.Query(
		`SELECT id, name, area_id, capacity, current_rpm, status, power_kw, installed_at, last_service_at, is_active FROM fans WHERE is_active = 1 ORDER BY area_id, name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list fans: %w", err)
	}
	defer rows.Close()
	var fans []*model.Fan
	for rows.Next() {
		var fan model.Fan
		var isActive int
		if err := rows.Scan(&fan.ID, &fan.Name, &fan.AreaID, &fan.Capacity, &fan.CurrentRPM, &fan.Status, &fan.PowerKW, &fan.InstalledAt, &fan.LastServiceAt, &isActive); err != nil {
			return nil, fmt.Errorf("scan fan: %w", err)
		}
		fan.IsActive = isActive == 1
		fans = append(fans, &fan)
	}
	return fans, nil
}

func (fs *FanStore) ListByArea(areaID string) ([]*model.Fan, error) {
	rows, err := fs.db.Query(
		`SELECT id, name, area_id, capacity, current_rpm, status, power_kw, installed_at, last_service_at, is_active FROM fans WHERE area_id = ? AND is_active = 1 ORDER BY name`,
		areaID,
	)
	if err != nil {
		return nil, fmt.Errorf("list fans by area: %w", err)
	}
	defer rows.Close()
	var fans []*model.Fan
	for rows.Next() {
		var fan model.Fan
		var isActive int
		if err := rows.Scan(&fan.ID, &fan.Name, &fan.AreaID, &fan.Capacity, &fan.CurrentRPM, &fan.Status, &fan.PowerKW, &fan.InstalledAt, &fan.LastServiceAt, &isActive); err != nil {
			return nil, fmt.Errorf("scan fan: %w", err)
		}
		fan.IsActive = isActive == 1
		fans = append(fans, &fan)
	}
	return fans, nil
}

func (fs *FanStore) Update(fan *model.Fan) error {
	_, err := fs.db.Exec(
		`UPDATE fans SET name = ?, area_id = ?, capacity = ?, current_rpm = ?, status = ?, power_kw = ?, last_service_at = ?, is_active = ? WHERE id = ?`,
		fan.Name, fan.AreaID, fan.Capacity, fan.CurrentRPM, fan.Status, fan.PowerKW, fan.LastServiceAt, fan.IsActive, fan.ID,
	)
	if err != nil {
		return fmt.Errorf("update fan: %w", err)
	}
	fs.cache[fan.ID] = fan
	return nil
}

func (fs *FanStore) UpdateStatus(id string, status model.FanStatus, rpm int) error {
	_, err := fs.db.Exec(
		`UPDATE fans SET status = ?, current_rpm = ? WHERE id = ?`,
		status, rpm, id,
	)
	if err != nil {
		return fmt.Errorf("update fan status: %w", err)
	}
	if cached, ok := fs.cache[id]; ok {
		cached.Status = status
		cached.CurrentRPM = rpm
	}
	return nil
}

func (fs *FanStore) Delete(id string) error {
	_, err := fs.db.Exec(`DELETE FROM fans WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete fan: %w", err)
	}
	delete(fs.cache, id)
	return nil
}

func (fs *FanStore) GetByStatus(status model.FanStatus) ([]*model.Fan, error) {
	rows, err := fs.db.Query(
		`SELECT id, name, area_id, capacity, current_rpm, status, power_kw, installed_at, last_service_at, is_active FROM fans WHERE status = ? AND is_active = 1`,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf("get fans by status: %w", err)
	}
	defer rows.Close()
	var fans []*model.Fan
	for rows.Next() {
		var fan model.Fan
		var isActive int
		if err := rows.Scan(&fan.ID, &fan.Name, &fan.AreaID, &fan.Capacity, &fan.CurrentRPM, &fan.Status, &fan.PowerKW, &fan.InstalledAt, &fan.LastServiceAt, &isActive); err != nil {
			return nil, fmt.Errorf("scan fan: %w", err)
		}
		fan.IsActive = isActive == 1
		fans = append(fans, &fan)
	}
	return fans, nil
}

func (fs *FanStore) CountByStatus() (map[model.FanStatus]int, error) {
	rows, err := fs.db.Query(
		`SELECT status, COUNT(*) FROM fans WHERE is_active = 1 GROUP BY status`,
	)
	if err != nil {
		return nil, fmt.Errorf("count fans by status: %w", err)
	}
	defer rows.Close()
	result := make(map[model.FanStatus]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan fan count: %w", err)
		}
		result[model.FanStatus(status)] = count
	}
	return result, nil
}

func (fs *FanStore) TouchServiceDate(id string, when time.Time) error {
	_, err := fs.db.Exec(`UPDATE fans SET last_service_at = ? WHERE id = ?`, when, id)
	if err != nil {
		return fmt.Errorf("touch service date: %w", err)
	}
	if cached, ok := fs.cache[id]; ok {
		cached.LastServiceAt = when
	}
	return nil
}
