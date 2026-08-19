package model

import "time"

type FanStatus string

const (
	FanStatusRunning   FanStatus = "running"
	FanStatusStopped   FanStatus = "stopped"
	FanStatusFault     FanStatus = "fault"
	FanStatusMaintenance FanStatus = "maintenance"
)

type Fan struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	AreaID       string    `json:"area_id"`
	Capacity     float64   `json:"capacity"`
	CurrentRPM   int       `json:"current_rpm"`
	Status       FanStatus `json:"status"`
	PowerKW      float64   `json:"power_kw"`
	InstalledAt  time.Time `json:"installed_at"`
	LastServiceAt time.Time `json:"last_service_at"`
	IsActive     bool      `json:"is_active"`
}

type FanHealth struct {
	FanID         string  `json:"fan_id"`
	FanName       string  `json:"fan_name"`
	HealthScore   float64 `json:"health_score"`
	Efficiency    float64 `json:"efficiency"`
	UptimeHours   float64 `json:"uptime_hours"`
	LastReading   float64 `json:"last_reading"`
	NeedsService  bool    `json:"needs_service"`
	Warnings      []string `json:"warnings"`
}

type FanRotationPlan struct {
	ActiveFanID    string `json:"active_fan_id"`
	StandbyFanID   string `json:"standby_fan_id"`
	NextRotationAt time.Time `json:"next_rotation_at"`
	Reason         string `json:"reason"`
}
