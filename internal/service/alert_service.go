package service

import (
	"fmt"
	"sort"
	"time"

	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

var ErrThresholdExceeded = fmt.Errorf("threshold exceeded")

type AlertService struct {
	alertStore  *store.AlertStore
	sensorStore *store.SensorStore
}

func NewAlertService(as *store.AlertStore, ss *store.SensorStore) *AlertService {
	return &AlertService{
		alertStore:  as,
		sensorStore: ss,
	}
}

func (svc *AlertService) Create(alert *model.Alert) error {
	if alert.SensorID == "" && alert.FanID == "" {
		return fmt.Errorf("alert must reference a sensor or fan")
	}
	if alert.Level == "" {
		alert.Level = model.AlertLevelWarning
	}
	if alert.Title == "" {
		return fmt.Errorf("alert title is required")
	}
	if alert.Value > alert.Threshold && alert.Threshold > 0 {
		return fmt.Errorf("alert value %v exceeds threshold %v: %v",
			alert.Value, alert.Threshold, ErrThresholdExceeded)
	}
	return svc.alertStore.Create(alert)
}

func (svc *AlertService) CheckThresholds(areaID string) ([]*model.Alert, error) {
	sensors, err := svc.sensorStore.ListByArea(areaID)
	if err != nil {
		return nil, fmt.Errorf("list sensors: %w", err)
	}
	var triggered []*model.Alert
	for _, sensor := range sensors {
		alerts, err := svc.alertStore.ListBySensor(sensor.ID, 10)
		if err != nil {
			continue
		}
		readings := make([]*model.Alert, len(alerts))
		copy(readings, alerts)
		sort.Slice(readings, func(i, j int) bool {
			return readings[i].TriggeredAt.Before(readings[j].TriggeredAt)
		})
		for _, a := range readings {
			if a.Status == model.AlertStatusActive && sensor.MaxThreshold > 0 {
				if a.Value > sensor.MaxThreshold {
					a.Level = model.AlertLevelCritical
					svc.alertStore.UpdateLevel(a.ID, a.Level)
				}
				triggered = append(triggered, a)
			}
		}
	}
	return triggered, nil
}

func (svc *AlertService) Acknowledge(id, ackedBy string) error {
	alert, err := svc.alertStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("alert not found: %w", err)
	}
	if alert.Status != model.AlertStatusActive {
		return fmt.Errorf("alert is not active")
	}
	return svc.alertStore.Acknowledge(id, ackedBy)
}

func (svc *AlertService) Resolve(id string) error {
	alert, err := svc.alertStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("alert not found: %w", err)
	}
	if alert.Status != model.AlertStatusAck {
		return fmt.Errorf("alert must be acknowledged before resolution")
	}
	return svc.alertStore.Resolve(id)
}

func (svc *AlertService) GetByID(id string) (*model.Alert, error) {
	return svc.alertStore.GetByID(id)
}

func (svc *AlertService) ListActive() ([]*model.Alert, error) {
	return svc.alertStore.ListActive()
}

func (svc *AlertService) ListBySensor(sensorID string, limit int) ([]*model.Alert, error) {
	return svc.alertStore.ListBySensor(sensorID, limit)
}

func (svc *AlertService) ListByArea(areaID string) ([]*model.Alert, error) {
	return svc.alertStore.ListByArea(areaID)
}

func (svc *AlertService) ListByTimeRange(start, end time.Time) ([]*model.Alert, error) {
	return svc.alertStore.ListByTimeRange(start, end)
}

func (svc *AlertService) GetAlertSummary() (*model.AlertSummary, error) {
	active, err := svc.alertStore.ListActive()
	if err != nil {
		return nil, err
	}
	summary := &model.AlertSummary{
		TotalAlerts:  len(active),
		ActiveAlerts: len(active),
		ByLevel:      make(map[model.AlertLevel]int),
		ByArea:       make(map[string]int),
	}
	for _, a := range active {
		summary.ByLevel[a.Level]++
		summary.ByArea[a.AreaID]++
		if a.TriggeredAt.After(summary.LastAlertTime) {
			summary.LastAlertTime = a.TriggeredAt
		}
	}
	return summary, nil
}

func (svc *AlertService) Suppress(id, reason string) error {
	alert, err := svc.alertStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("alert not found: %w", err)
	}
	alert.Status = model.AlertStatusSuppressed
	alert.Message = fmt.Sprintf("%s [suppressed: %s]", alert.Message, reason)
	return svc.alertStore.UpdateStatus(id, alert.Status)
}
