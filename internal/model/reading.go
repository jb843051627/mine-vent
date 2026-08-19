package model

import "time"

type Reading struct {
	ID        string    `json:"id"`
	SensorID  string    `json:"sensor_id"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
	RecordedAt time.Time `json:"recorded_at"`
	Quality   float64   `json:"quality"`
}

type ReadingBatch struct {
	BatchID   string    `json:"batch_id"`
	Readings  []Reading `json:"readings"`
	Source    string    `json:"source"`
	IngestedAt time.Time `json:"ingested_at"`
}

type AggregatedReading struct {
	SensorID   string  `json:"sensor_id"`
	SensorName string  `json:"sensor_name"`
	Type       SensorType `json:"type"`
	Count      int     `json:"count"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	Avg        float64 `json:"avg"`
	Latest     float64 `json:"latest"`
}

type ReadingStats struct {
	AreaID    string  `json:"area_id"`
	TotalReadings int `json:"total_readings"`
	ActiveSensors int `json:"active_sensors"`
	LastUpdate time.Time `json:"last_update"`
}
