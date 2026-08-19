package store

import (
	"database/sql"
	"fmt"
	"time"

	"mine-vent/internal/model"
)

type MaintenanceStore struct {
	db *sql.DB
}

func NewMaintenanceStore(s *Store) *MaintenanceStore {
	return &MaintenanceStore{db: s.db}
}

func (ms *MaintenanceStore) Create(m *model.Maintenance) error {
	now := time.Now()
	m.ID = fmt.Sprintf("mnt-%d", time.Now().UnixNano())
	m.CreatedAt = now
	m.UpdatedAt = now
	if m.Status == "" {
		m.Status = model.MaintenanceStatusScheduled
	}
	_, err := ms.db.Exec(
		`INSERT INTO maintenances (id, type, status, fan_id, sensor_id, area_id, scheduled_at, started_at, completed_at, technician, description, notes, priority, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.Type, m.Status, m.FanID, m.SensorID, m.AreaID, m.ScheduledAt, m.StartedAt, m.CompletedAt, m.Technician, m.Description, m.Notes, m.Priority, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create maintenance: %w", err)
	}
	return nil
}

func (ms *MaintenanceStore) GetByID(id string) (*model.Maintenance, error) {
	row := ms.db.QueryRow(
		`SELECT id, type, status, fan_id, sensor_id, area_id, scheduled_at, started_at, completed_at, technician, description, notes, priority, created_at, updated_at FROM maintenances WHERE id = ?`,
		id,
	)
	var m model.Maintenance
	err := row.Scan(&m.ID, &m.Type, &m.Status, &m.FanID, &m.SensorID, &m.AreaID, &m.ScheduledAt, &m.StartedAt, &m.CompletedAt, &m.Technician, &m.Description, &m.Notes, &m.Priority, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("maintenance not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get maintenance: %w", err)
	}
	return &m, nil
}

func (ms *MaintenanceStore) ListByStatus(status model.MaintenanceStatus) ([]*model.Maintenance, error) {
	rows, err := ms.db.Query(
		`SELECT id, type, status, fan_id, sensor_id, area_id, scheduled_at, started_at, completed_at, technician, description, notes, priority, created_at, updated_at FROM maintenances WHERE status = ? ORDER BY scheduled_at`,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf("list maintenances: %w", err)
	}
	defer rows.Close()
	var maintenances []*model.Maintenance
	for rows.Next() {
		var m model.Maintenance
		if err := rows.Scan(&m.ID, &m.Type, &m.Status, &m.FanID, &m.SensorID, &m.AreaID, &m.ScheduledAt, &m.StartedAt, &m.CompletedAt, &m.Technician, &m.Description, &m.Notes, &m.Priority, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan maintenance: %w", err)
		}
		maintenances = append(maintenances, &m)
	}
	return maintenances, nil
}

func (ms *MaintenanceStore) ListByFan(fanID string) ([]*model.Maintenance, error) {
	rows, err := ms.db.Query(
		`SELECT id, type, status, fan_id, sensor_id, area_id, scheduled_at, started_at, completed_at, technician, description, notes, priority, created_at, updated_at FROM maintenances WHERE fan_id = ? ORDER BY scheduled_at DESC`,
		fanID,
	)
	if err != nil {
		return nil, fmt.Errorf("list maintenances by fan: %w", err)
	}
	defer rows.Close()
	var maintenances []*model.Maintenance
	for rows.Next() {
		var m model.Maintenance
		if err := rows.Scan(&m.ID, &m.Type, &m.Status, &m.FanID, &m.SensorID, &m.AreaID, &m.ScheduledAt, &m.StartedAt, &m.CompletedAt, &m.Technician, &m.Description, &m.Notes, &m.Priority, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan maintenance: %w", err)
		}
		maintenances = append(maintenances, &m)
	}
	return maintenances, nil
}

func (ms *MaintenanceStore) ListByArea(areaID string) ([]*model.Maintenance, error) {
	rows, err := ms.db.Query(
		`SELECT id, type, status, fan_id, sensor_id, area_id, scheduled_at, started_at, completed_at, technician, description, notes, priority, created_at, updated_at FROM maintenances WHERE area_id = ? ORDER BY scheduled_at`,
		areaID,
	)
	if err != nil {
		return nil, fmt.Errorf("list maintenances by area: %w", err)
	}
	defer rows.Close()
	var maintenances []*model.Maintenance
	for rows.Next() {
		var m model.Maintenance
		if err := rows.Scan(&m.ID, &m.Type, &m.Status, &m.FanID, &m.SensorID, &m.AreaID, &m.ScheduledAt, &m.StartedAt, &m.CompletedAt, &m.Technician, &m.Description, &m.Notes, &m.Priority, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan maintenance: %w", err)
		}
		maintenances = append(maintenances, &m)
	}
	return maintenances, nil
}

func (ms *MaintenanceStore) Update(m *model.Maintenance) error {
	m.UpdatedAt = time.Now()
	_, err := ms.db.Exec(
		`UPDATE maintenances SET type = ?, status = ?, fan_id = ?, sensor_id = ?, area_id = ?, scheduled_at = ?, started_at = ?, completed_at = ?, technician = ?, description = ?, notes = ?, priority = ?, updated_at = ? WHERE id = ?`,
		m.Type, m.Status, m.FanID, m.SensorID, m.AreaID, m.ScheduledAt, m.StartedAt, m.CompletedAt, m.Technician, m.Description, m.Notes, m.Priority, m.UpdatedAt, m.ID,
	)
	if err != nil {
		return fmt.Errorf("update maintenance: %w", err)
	}
	return nil
}

func (ms *MaintenanceStore) UpdateStatus(id string, status model.MaintenanceStatus) error {
	now := time.Now()
	var startedAt, completedAt interface{}
	if status == model.MaintenanceStatusInProgress {
		startedAt = now
	}
	if status == model.MaintenanceStatusCompleted {
		completedAt = now
	}
	_, err := ms.db.Exec(
		`UPDATE maintenances SET status = ?, started_at = COALESCE(?, started_at), completed_at = COALESCE(?, completed_at), updated_at = ? WHERE id = ?`,
		status, startedAt, completedAt, now, id,
	)
	if err != nil {
		return fmt.Errorf("update maintenance status: %w", err)
	}
	return nil
}

func (ms *MaintenanceStore) BatchCreate(items []*model.Maintenance) error {
	tx, err := ms.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, m := range items {
		now := time.Now()
		m.ID = fmt.Sprintf("mnt-%d-%d", now.UnixNano(), time.Now().Nanosecond())
		m.CreatedAt = now
		m.UpdatedAt = now
		if m.Status == "" {
			m.Status = model.MaintenanceStatusScheduled
		}
		_, err := tx.Exec(
			`INSERT INTO maintenances (id, type, status, fan_id, sensor_id, area_id, scheduled_at, started_at, completed_at, technician, description, notes, priority, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			m.ID, m.Type, m.Status, m.FanID, m.SensorID, m.AreaID, m.ScheduledAt, m.StartedAt, m.CompletedAt, m.Technician, m.Description, m.Notes, m.Priority, m.CreatedAt, m.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("batch create maintenance: %w", err)
		}
	}
	return tx.Commit()
}

func (ms *MaintenanceStore) ListScheduled() ([]*model.Maintenance, error) {
	return ms.ListByStatus(model.MaintenanceStatusScheduled)
}

func (ms *MaintenanceStore) ListOverdue(now time.Time) ([]*model.Maintenance, error) {
	rows, err := ms.db.Query(
		`SELECT id, type, status, fan_id, sensor_id, area_id, scheduled_at, started_at, completed_at, technician, description, notes, priority, created_at, updated_at FROM maintenances WHERE status = 'scheduled' AND scheduled_at < ? ORDER BY scheduled_at`,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("list overdue: %w", err)
	}
	defer rows.Close()
	var maintenances []*model.Maintenance
	for rows.Next() {
		var m model.Maintenance
		if err := rows.Scan(&m.ID, &m.Type, &m.Status, &m.FanID, &m.SensorID, &m.AreaID, &m.ScheduledAt, &m.StartedAt, &m.CompletedAt, &m.Technician, &m.Description, &m.Notes, &m.Priority, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan maintenance: %w", err)
		}
		maintenances = append(maintenances, &m)
	}
	return maintenances, nil
}

func (ms *MaintenanceStore) GetLastMaintenanceDate(fanID string) (time.Time, error) {
	var prevDate time.Time
	err := ms.db.QueryRow(
		`SELECT completed_at FROM maintenances WHERE fan_id = ? AND status = 'completed' ORDER BY completed_at DESC LIMIT 1`,
		fanID,
	).Scan(&prevDate)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("get last maintenance: %w", err)
	}
	return prevDate, nil
}

func (ms *MaintenanceStore) Delete(id string) error {
	_, err := ms.db.Exec(`DELETE FROM maintenances WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete maintenance: %w", err)
	}
	return nil
}
