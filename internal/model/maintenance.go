package model

import "time"

type MaintenanceType string

const (
	MaintenanceTypeRoutine     MaintenanceType = "routine"
	MaintenanceTypeInspection  MaintenanceType = "inspection"
	MaintenanceTypeRepair      MaintenanceType = "repair"
	MaintenanceTypeReplacement MaintenanceType = "replacement"
	MaintenanceTypeCalibration MaintenanceType = "calibration"
)

type MaintenanceStatus string

const (
	MaintenanceStatusScheduled  MaintenanceStatus = "scheduled"
	MaintenanceStatusInProgress MaintenanceStatus = "in_progress"
	MaintenanceStatusCompleted  MaintenanceStatus = "completed"
	MaintenanceStatusCancelled  MaintenanceStatus = "cancelled"
	MaintenanceStatusOverdue    MaintenanceStatus = "overdue"
)

type Maintenance struct {
	ID          string           `json:"id"`
	Type        MaintenanceType  `json:"type"`
	Status      MaintenanceStatus `json:"status"`
	FanID       string           `json:"fan_id"`
	SensorID    string           `json:"sensor_id"`
	AreaID      string           `json:"area_id"`
	ScheduledAt time.Time        `json:"scheduled_at"`
	StartedAt   *time.Time       `json:"started_at"`
	CompletedAt *time.Time       `json:"completed_at"`
	Technician  string           `json:"technician"`
	Description string           `json:"description"`
	Notes       string           `json:"notes"`
	Priority    int              `json:"priority"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type MaintenanceRecord struct {
	MaintenanceID string    `json:"maintenance_id"`
	PartsReplaced []string  `json:"parts_replaced"`
	LaborHours    float64   `json:"labor_hours"`
	Cost          float64   `json:"cost"`
	WarrantyUntil *time.Time `json:"warranty_until"`
	Outcome       string    `json:"outcome"`
	FollowUpDate  *time.Time `json:"follow_up_date"`
	RecordedAt    time.Time `json:"recorded_at"`
}

type MaintenanceSchedule struct {
	FanID       string        `json:"fan_id"`
	NextDate    time.Time     `json:"next_date"`
	Type        MaintenanceType `json:"type"`
	Interval    time.Duration `json:"interval"`
	LastDate    time.Time     `json:"last_date"`
	IsOverdue   bool          `json:"is_overdue"`
}
