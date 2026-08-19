package service

import (
	"fmt"
	"time"

	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

type SensorService struct {
	sensorStore *store.SensorStore
}

func NewSensorService(ss *store.SensorStore) *SensorService {
	return &SensorService{sensorStore: ss}
}

func (svc *SensorService) Register(sensor *model.Sensor) error {
	if sensor.ID == "" {
		sensor.ID = fmt.Sprintf("sensor-%d", time.Now().UnixNano())
	}
	if sensor.Name == "" {
		return fmt.Errorf("sensor name is required")
	}
	if sensor.Type == "" {
		return fmt.Errorf("sensor type is required")
	}
	if sensor.AreaID == "" {
		return fmt.Errorf("area id is required")
	}
	if sensor.MinThreshold > sensor.MaxThreshold {
		return fmt.Errorf("min threshold cannot be greater than max threshold")
	}
	return svc.sensorStore.Create(sensor)
}

func (svc *SensorService) GetByID(id string) (*model.Sensor, error) {
	sensor, err := svc.sensorStore.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get sensor: %w", err)
	}
	return sensor, nil
}

func (svc *SensorService) ListByArea(areaID string) ([]*model.Sensor, error) {
	sensors, err := svc.sensorStore.ListByArea(areaID)
	if err != nil {
		return nil, fmt.Errorf("list sensors: %w", err)
	}
	return sensors, nil
}

func (svc *SensorService) ListAll() ([]*model.Sensor, error) {
	sensors, err := svc.sensorStore.ListAll()
	if err != nil {
		return nil, fmt.Errorf("list all sensors: %w", err)
	}
	return sensors, nil
}

func (svc *SensorService) ListByType(sensorType model.SensorType) ([]*model.Sensor, error) {
	sensors, err := svc.sensorStore.ListByType(sensorType)
	if err != nil {
		return nil, fmt.Errorf("list sensors by type: %w", err)
	}
	return sensors, nil
}

func (svc *SensorService) Update(sensor *model.Sensor) error {
	existing, err := svc.sensorStore.GetByID(sensor.ID)
	if err != nil {
		return fmt.Errorf("sensor not found: %w", err)
	}
	if sensor.MinThreshold > sensor.MaxThreshold {
		return fmt.Errorf("min threshold cannot be greater than max threshold")
	}
	if sensor.Name == "" {
		sensor.Name = existing.Name
	}
	if sensor.Type == "" {
		sensor.Type = existing.Type
	}
	return svc.sensorStore.Update(sensor)
}

func (svc *SensorService) Deactivate(id string) error {
	sensor, err := svc.sensorStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("sensor not found: %w", err)
	}
	sensor.IsActive = false
	return svc.sensorStore.Update(sensor)
}

func (svc *SensorService) Activate(id string) error {
	sensor, err := svc.sensorStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("sensor not found: %w", err)
	}
	sensor.IsActive = true
	return svc.sensorStore.Update(sensor)
}

func (svc *SensorService) Delete(id string) error {
	return svc.sensorStore.Delete(id)
}

func (svc *SensorService) GetSensorIDsByArea(areaID string) ([]string, error) {
	sensors, err := svc.sensorStore.ListByArea(areaID)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, s := range sensors {
		ids = append(ids, s.ID)
	}
	return ids, nil
}
