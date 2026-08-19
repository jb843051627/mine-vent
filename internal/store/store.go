package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(30 * time.Minute)

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sensors (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			direction TEXT NOT NULL,
			location TEXT NOT NULL,
			area_id TEXT NOT NULL,
			unit TEXT NOT NULL,
			min_threshold REAL NOT NULL,
			max_threshold REAL NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS readings (
			id TEXT PRIMARY KEY,
			sensor_id TEXT NOT NULL,
			value REAL NOT NULL,
			unit TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			recorded_at DATETIME NOT NULL,
			quality REAL NOT NULL DEFAULT 1.0,
			FOREIGN KEY (sensor_id) REFERENCES sensors(id)
		)`,
		`CREATE TABLE IF NOT EXISTS fans (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			area_id TEXT NOT NULL,
			capacity REAL NOT NULL,
			current_rpm INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'stopped',
			power_kw REAL NOT NULL DEFAULT 0,
			installed_at DATETIME NOT NULL,
			last_service_at DATETIME NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS schedules (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'draft',
			fan_id TEXT NOT NULL,
			area_id TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			interval_ns INTEGER NOT NULL,
			target_rpm INTEGER NOT NULL,
			priority INTEGER NOT NULL DEFAULT 0,
			created_by TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (fan_id) REFERENCES fans(id)
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			sensor_id TEXT,
			fan_id TEXT,
			level TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			title TEXT NOT NULL,
			message TEXT NOT NULL,
			value REAL NOT NULL,
			threshold REAL NOT NULL,
			triggered_at DATETIME NOT NULL,
			acked_at DATETIME,
			acked_by TEXT,
			resolved_at DATETIME,
			area_id TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS maintenances (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'scheduled',
			fan_id TEXT NOT NULL,
			sensor_id TEXT,
			area_id TEXT NOT NULL,
			scheduled_at DATETIME NOT NULL,
			started_at DATETIME,
			completed_at DATETIME,
			technician TEXT NOT NULL,
			description TEXT,
			notes TEXT,
			priority INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS maintenance_records (
			maintenance_id TEXT PRIMARY KEY,
			parts_replaced TEXT,
			labor_hours REAL NOT NULL DEFAULT 0,
			cost REAL NOT NULL DEFAULT 0,
			warranty_until DATETIME,
			outcome TEXT,
			follow_up_date DATETIME,
			recorded_at DATETIME NOT NULL,
			FOREIGN KEY (maintenance_id) REFERENCES maintenances(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_readings_sensor ON readings(sensor_id, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status, level)`,
		`CREATE INDEX IF NOT EXISTS idx_schedules_status ON schedules(status, priority)`,
		`CREATE INDEX IF NOT EXISTS idx_maintenances_status ON maintenances(status, scheduled_at)`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("exec schema: %w", err)
		}
	}
	return nil
}

func (s *Store) BeginTx() (*sql.Tx, error) {
	return s.db.Begin()
}

func (s *Store) generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
