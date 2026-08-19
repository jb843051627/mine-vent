package model

import "time"

type ScheduleType string

const (
	ScheduleTypeContinuous  ScheduleType = "continuous"
	ScheduleTypePeriodic    ScheduleType = "periodic"
	ScheduleTypeEmergency   ScheduleType = "emergency"
)

type ScheduleStatus string

const (
	ScheduleStatusDraft     ScheduleStatus = "draft"
	ScheduleStatusActive    ScheduleStatus = "active"
	ScheduleStatusPaused    ScheduleStatus = "paused"
	ScheduleStatusCompleted ScheduleStatus = "completed"
)

type Schedule struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Type        ScheduleType  `json:"type"`
	Status      ScheduleStatus `json:"status"`
	FanID       string        `json:"fan_id"`
	AreaID      string        `json:"area_id"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Interval    time.Duration `json:"interval"`
	TargetRPM   int           `json:"target_rpm"`
	Priority    int           `json:"priority"`
	CreatedBy   string        `json:"created_by"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type ScheduleLog struct {
	ID          string    `json:"id"`
	ScheduleID  string    `json:"schedule_id"`
	Action      string    `json:"action"`
	Timestamp   time.Time `json:"timestamp"`
	Operator    string    `json:"operator"`
	Details     string    `json:"details"`
}

type MaintenanceWindow struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Reason    string    `json:"reason"`
	FanID     string    `json:"fan_id"`
}
