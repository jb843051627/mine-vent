package model

import "time"

type SensorType string

const (
	SensorTypeGasCH4    SensorType = "CH4"
	SensorTypeGasCO     SensorType = "CO"
	SensorTypeAirflow   SensorType = "AIRFLOW"
	SensorTypePressure  SensorType = "PRESSURE"
	SensorTypeTemperature SensorType = "TEMP"
)

type SensorDirection string

const (
	DirectionIntake  SensorDirection = "intake"
	DirectionExhaust SensorDirection = "exhaust"
)

type Sensor struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Type      SensorType     `json:"type"`
	Direction SensorDirection `json:"direction"`
	Location  string         `json:"location"`
	AreaID    string         `json:"area_id"`
	Unit      string         `json:"unit"`
	MinThreshold float64     `json:"min_threshold"`
	MaxThreshold float64     `json:"max_threshold"`
	IsActive  bool           `json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type SensorArea struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Level    int    `json:"level"`
	Section  string `json:"section"`
}
