package store

import (
	"database/sql"
	"fmt"
	"time"

	"mine-vent/internal/model"
)

type ScheduleStore struct {
	db *sql.DB
}

func NewScheduleStore(s *Store) *ScheduleStore {
	return &ScheduleStore{db: s.db}
}

func (ss *ScheduleStore) Create(schedule *model.Schedule) error {
	now := time.Now()
	schedule.CreatedAt = now
	schedule.UpdatedAt = now
	_, err := ss.db.Exec(
		`INSERT INTO schedules (id, name, type, status, fan_id, area_id, start_time, end_time, interval_ns, target_rpm, priority, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		schedule.ID, schedule.Name, schedule.Type, schedule.Status, schedule.FanID, schedule.AreaID, schedule.StartTime, schedule.EndTime, int64(schedule.Interval), schedule.TargetRPM, schedule.Priority, schedule.CreatedBy, schedule.CreatedAt, schedule.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create schedule: %w", err)
	}
	return nil
}

func (ss *ScheduleStore) GetByID(id string) (*model.Schedule, error) {
	row := ss.db.QueryRow(
		`SELECT id, name, type, status, fan_id, area_id, start_time, end_time, interval_ns, target_rpm, priority, created_by, created_at, updated_at FROM schedules WHERE id = ?`,
		id,
	)
	var s model.Schedule
	var intervalNS int64
	err := row.Scan(&s.ID, &s.Name, &s.Type, &s.Status, &s.FanID, &s.AreaID, &s.StartTime, &s.EndTime, &intervalNS, &s.TargetRPM, &s.Priority, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("schedule not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get schedule: %w", err)
	}
	s.Interval = time.Duration(intervalNS)
	return &s, nil
}

func (ss *ScheduleStore) ListByStatus(status model.ScheduleStatus) ([]*model.Schedule, error) {
	rows, err := ss.db.Query(
		`SELECT id, name, type, status, fan_id, area_id, start_time, end_time, interval_ns, target_rpm, priority, created_by, created_at, updated_at FROM schedules WHERE status = ? ORDER BY priority DESC, start_time`,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf("list schedules: %w", err)
	}
	defer rows.Close()
	var schedules []*model.Schedule
	for rows.Next() {
		var s model.Schedule
		var intervalNS int64
		if err := rows.Scan(&s.ID, &s.Name, &s.Type, &s.Status, &s.FanID, &s.AreaID, &s.StartTime, &s.EndTime, &intervalNS, &s.TargetRPM, &s.Priority, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan schedule: %w", err)
		}
		s.Interval = time.Duration(intervalNS)
		schedules = append(schedules, &s)
	}
	return schedules, nil
}

func (ss *ScheduleStore) ListByFan(fanID string) ([]*model.Schedule, error) {
	rows, err := ss.db.Query(
		`SELECT id, name, type, status, fan_id, area_id, start_time, end_time, interval_ns, target_rpm, priority, created_by, created_at, updated_at FROM schedules WHERE fan_id = ? ORDER BY start_time DESC`,
		fanID,
	)
	if err != nil {
		return nil, fmt.Errorf("list schedules by fan: %w", err)
	}
	defer rows.Close()
	var schedules []*model.Schedule
	for rows.Next() {
		var s model.Schedule
		var intervalNS int64
		if err := rows.Scan(&s.ID, &s.Name, &s.Type, &s.Status, &s.FanID, &s.AreaID, &s.StartTime, &s.EndTime, &intervalNS, &s.TargetRPM, &s.Priority, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan schedule: %w", err)
		}
		s.Interval = time.Duration(intervalNS)
		schedules = append(schedules, &s)
	}
	return schedules, nil
}

func (ss *ScheduleStore) Update(schedule *model.Schedule) error {
	schedule.UpdatedAt = time.Now()
	_, err := ss.db.Exec(
		`UPDATE schedules SET name = ?, type = ?, status = ?, fan_id = ?, area_id = ?, start_time = ?, end_time = ?, interval_ns = ?, target_rpm = ?, priority = ?, updated_at = ? WHERE id = ?`,
		schedule.Name, schedule.Type, schedule.Status, schedule.FanID, schedule.AreaID, schedule.StartTime, schedule.EndTime, int64(schedule.Interval), schedule.TargetRPM, schedule.Priority, schedule.UpdatedAt, schedule.ID,
	)
	if err != nil {
		return fmt.Errorf("update schedule: %w", err)
	}
	return nil
}

func (ss *ScheduleStore) UpdateStatus(id string, status model.ScheduleStatus) error {
	_, err := ss.db.Exec(
		`UPDATE schedules SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("update schedule status: %w", err)
	}
	return nil
}

func (ss *ScheduleStore) ListActive() ([]*model.Schedule, error) {
	return ss.ListByStatus(model.ScheduleStatusActive)
}

func (ss *ScheduleStore) ListByArea(areaID string) ([]*model.Schedule, error) {
	rows, err := ss.db.Query(
		`SELECT id, name, type, status, fan_id, area_id, start_time, end_time, interval_ns, target_rpm, priority, created_by, created_at, updated_at FROM schedules WHERE area_id = ? ORDER BY priority DESC, start_time`,
		areaID,
	)
	if err != nil {
		return nil, fmt.Errorf("list schedules by area: %w", err)
	}
	defer rows.Close()
	var schedules []*model.Schedule
	for rows.Next() {
		var s model.Schedule
		var intervalNS int64
		if err := rows.Scan(&s.ID, &s.Name, &s.Type, &s.Status, &s.FanID, &s.AreaID, &s.StartTime, &s.EndTime, &intervalNS, &s.TargetRPM, &s.Priority, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan schedule: %w", err)
		}
		s.Interval = time.Duration(intervalNS)
		schedules = append(schedules, &s)
	}
	return schedules, nil
}

func (ss *ScheduleStore) Delete(id string) error {
	_, err := ss.db.Exec(`DELETE FROM schedules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	return nil
}

func (ss *ScheduleStore) CountByStatus() (map[model.ScheduleStatus]int, error) {
	rows, err := ss.db.Query(`SELECT status, COUNT(*) FROM schedules GROUP BY status`)
	if err != nil {
		return nil, fmt.Errorf("count schedules: %w", err)
	}
	defer rows.Close()
	result := make(map[model.ScheduleStatus]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan schedule count: %w", err)
		}
		result[model.ScheduleStatus(status)] = count
	}
	return result, nil
}
