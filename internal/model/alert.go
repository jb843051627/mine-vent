package model

import "time"

type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
	AlertLevelFatal    AlertLevel = "fatal"
)

type AlertStatus string

const (
	AlertStatusActive    AlertStatus = "active"
	AlertStatusAck       AlertStatus = "acknowledged"
	AlertStatusResolved  AlertStatus = "resolved"
	AlertStatusSuppressed AlertStatus = "suppressed"
)

type Alert struct {
	ID          string      `json:"id"`
	SensorID    string      `json:"sensor_id"`
	FanID       string      `json:"fan_id"`
	Level       AlertLevel  `json:"level"`
	Status      AlertStatus `json:"status"`
	Title       string      `json:"title"`
	Message     string      `json:"message"`
	Value       float64     `json:"value"`
	Threshold   float64     `json:"threshold"`
	TriggeredAt time.Time   `json:"triggered_at"`
	AckedAt     *time.Time  `json:"acked_at"`
	AckedBy     string      `json:"acked_by"`
	ResolvedAt  *time.Time  `json:"resolved_at"`
	AreaID      string      `json:"area_id"`
}

type AlertThreshold struct {
	SensorType   SensorType  `json:"sensor_type"`
	MinValue     float64     `json:"min_value"`
	MaxValue     float64     `json:"max_value"`
	WarningLevel AlertLevel  `json:"warning_level"`
	CriticalLevel AlertLevel `json:"critical_level"`
}

type AlertSummary struct {
	TotalAlerts   int            `json:"total_alerts"`
	ActiveAlerts  int            `json:"active_alerts"`
	ByLevel       map[AlertLevel]int `json:"by_level"`
	ByArea        map[string]int `json:"by_area"`
	LastAlertTime time.Time      `json:"last_alert_time"`
}
