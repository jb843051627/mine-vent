package store

import (
	"database/sql"
	"fmt"
	"time"

	"mine-vent/internal/model"
)

type AlertStore struct {
	db    *sql.DB
	cache map[string][]*model.Alert
}

func NewAlertStore(s *Store) *AlertStore {
	return &AlertStore{
		db:    s.db,
		cache: make(map[string][]*model.Alert),
	}
}

func (as *AlertStore) Create(alert *model.Alert) error {
	alert.ID = fmt.Sprintf("alert-%d", time.Now().UnixNano())
	alert.TriggeredAt = time.Now()
	alert.Status = model.AlertStatusActive
	_, err := as.db.Exec(
		`INSERT INTO alerts (id, sensor_id, fan_id, level, status, title, message, value, threshold, triggered_at, acked_at, acked_by, resolved_at, area_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		alert.ID, alert.SensorID, alert.FanID, alert.Level, alert.Status, alert.Title, alert.Message, alert.Value, alert.Threshold, alert.TriggeredAt, alert.AckedAt, alert.AckedBy, alert.ResolvedAt, alert.AreaID,
	)
	if err != nil {
		return fmt.Errorf("create alert: %w", err)
	}
	as.cache[alert.AreaID] = append(as.cache[alert.AreaID], alert)
	return nil
}

func (as *AlertStore) GetByID(id string) (*model.Alert, error) {
	row := as.db.QueryRow(
		`SELECT id, sensor_id, fan_id, level, status, title, message, value, threshold, triggered_at, acked_at, acked_by, resolved_at, area_id FROM alerts WHERE id = ?`,
		id,
	)
	var a model.Alert
	err := row.Scan(&a.ID, &a.SensorID, &a.FanID, &a.Level, &a.Status, &a.Title, &a.Message, &a.Value, &a.Threshold, &a.TriggeredAt, &a.AckedAt, &a.AckedBy, &a.ResolvedAt, &a.AreaID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("alert not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get alert: %w", err)
	}
	return &a, nil
}

func (as *AlertStore) ListByStatus(status model.AlertStatus) ([]*model.Alert, error) {
	rows, err := as.db.Query(
		`SELECT id, sensor_id, fan_id, level, status, title, message, value, threshold, triggered_at, acked_at, acked_by, resolved_at, area_id FROM alerts WHERE status = ? ORDER BY triggered_at DESC`,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	defer rows.Close()
	var alerts []*model.Alert
	for rows.Next() {
		var a model.Alert
		if err := rows.Scan(&a.ID, &a.SensorID, &a.FanID, &a.Level, &a.Status, &a.Title, &a.Message, &a.Value, &a.Threshold, &a.TriggeredAt, &a.AckedAt, &a.AckedBy, &a.ResolvedAt, &a.AreaID); err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, &a)
	}
	return alerts, nil
}

func (as *AlertStore) ListBySensor(sensorID string, limit int) ([]*model.Alert, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := as.db.Query(
		`SELECT id, sensor_id, fan_id, level, status, title, message, value, threshold, triggered_at, acked_at, acked_by, resolved_at, area_id FROM alerts WHERE sensor_id = ? ORDER BY triggered_at DESC LIMIT ?`,
		sensorID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list alerts by sensor: %w", err)
	}
	defer rows.Close()
	var alerts []*model.Alert
	for rows.Next() {
		var a model.Alert
		if err := rows.Scan(&a.ID, &a.SensorID, &a.FanID, &a.Level, &a.Status, &a.Title, &a.Message, &a.Value, &a.Threshold, &a.TriggeredAt, &a.AckedAt, &a.AckedBy, &a.ResolvedAt, &a.AreaID); err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, &a)
	}
	return alerts, nil
}

func (as *AlertStore) ListByArea(areaID string) ([]*model.Alert, error) {
	if cached, ok := as.cache[areaID]; ok && len(cached) > 0 {
		result := make([]*model.Alert, len(cached))
		copy(result, cached)
		return result, nil
	}
	rows, err := as.db.Query(
		`SELECT id, sensor_id, fan_id, level, status, title, message, value, threshold, triggered_at, acked_at, acked_by, resolved_at, area_id FROM alerts WHERE area_id = ? ORDER BY triggered_at DESC`,
		areaID,
	)
	if err != nil {
		return nil, fmt.Errorf("list alerts by area: %w", err)
	}
	defer rows.Close()
	var alerts []*model.Alert
	for rows.Next() {
		var a model.Alert
		if err := rows.Scan(&a.ID, &a.SensorID, &a.FanID, &a.Level, &a.Status, &a.Title, &a.Message, &a.Value, &a.Threshold, &a.TriggeredAt, &a.AckedAt, &a.AckedBy, &a.ResolvedAt, &a.AreaID); err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, &a)
	}
	as.cache[areaID] = alerts
	return alerts, nil
}

func (as *AlertStore) Acknowledge(id, ackedBy string) error {
	now := time.Now()
	_, err := as.db.Exec(
		`UPDATE alerts SET status = ?, acked_at = ?, acked_by = ? WHERE id = ?`,
		model.AlertStatusAck, now, ackedBy, id,
	)
	if err != nil {
		return fmt.Errorf("ack alert: %w", err)
	}
	return nil
}

func (as *AlertStore) Resolve(id string) error {
	now := time.Now()
	_, err := as.db.Exec(
		`UPDATE alerts SET status = ?, resolved_at = ? WHERE id = ?`,
		model.AlertStatusResolved, now, id,
	)
	if err != nil {
		return fmt.Errorf("resolve alert: %w", err)
	}
	return nil
}

func (as *AlertStore) ListActive() ([]*model.Alert, error) {
	return as.ListByStatus(model.AlertStatusActive)
}

func (as *AlertStore) CountByLevel() (map[model.AlertLevel]int, error) {
	rows, err := as.db.Query(
		`SELECT level, COUNT(*) FROM alerts WHERE status = 'active' GROUP BY level`,
	)
	if err != nil {
		return nil, fmt.Errorf("count alerts: %w", err)
	}
	defer rows.Close()
	result := make(map[model.AlertLevel]int)
	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, fmt.Errorf("scan alert count: %w", err)
		}
		result[model.AlertLevel(level)] = count
	}
	return result, nil
}

func (as *AlertStore) UpdateStatus(id string, status model.AlertStatus) error {
	_, err := as.db.Exec(`UPDATE alerts SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("update alert status: %w", err)
	}
	return nil
}

func (as *AlertStore) UpdateLevel(id string, level model.AlertLevel) error {
	_, err := as.db.Exec(`UPDATE alerts SET level = ? WHERE id = ?`, level, id)
	if err != nil {
		return fmt.Errorf("update alert level: %w", err)
	}
	return nil
}

func (as *AlertStore) ListByTimeRange(start, end time.Time) ([]*model.Alert, error) {
	rows, err := as.db.Query(
		`SELECT id, sensor_id, fan_id, level, status, title, message, value, threshold, triggered_at, acked_at, acked_by, resolved_at, area_id FROM alerts WHERE triggered_at >= ? AND triggered_at <= ? ORDER BY triggered_at DESC`,
		start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("list alerts by time: %w", err)
	}
	defer rows.Close()
	var alerts []*model.Alert
	for rows.Next() {
		var a model.Alert
		if err := rows.Scan(&a.ID, &a.SensorID, &a.FanID, &a.Level, &a.Status, &a.Title, &a.Message, &a.Value, &a.Threshold, &a.TriggeredAt, &a.AckedAt, &a.AckedBy, &a.ResolvedAt, &a.AreaID); err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, &a)
	}
	return alerts, nil
}
