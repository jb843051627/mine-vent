package service

import (
	"fmt"
	"time"

	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

type MaintenanceService struct {
	maintStore *store.MaintenanceStore
	fanStore   *store.FanStore
}

func NewMaintenanceService(ms *store.MaintenanceStore, fs *store.FanStore) *MaintenanceService {
	return &MaintenanceService{
		maintStore: ms,
		fanStore:   fs,
	}
}

func (svc *MaintenanceService) Schedule(m *model.Maintenance) error {
	if m.FanID == "" {
		return fmt.Errorf("fan id is required")
	}
	if m.AreaID == "" {
		return fmt.Errorf("area id is required")
	}
	if m.Type == "" {
		m.Type = model.MaintenanceTypeRoutine
	}
	if m.Technician == "" {
		return fmt.Errorf("technician is required")
	}
	if m.ScheduledAt.IsZero() {
		m.ScheduledAt = time.Now().Add(72 * time.Hour)
	}
	if m.Priority < 0 || m.Priority > 10 {
		return fmt.Errorf("priority must be between 0 and 10")
	}
	return svc.maintStore.Create(m)
}

func (svc *MaintenanceService) Start(id string) error {
	m, err := svc.maintStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("maintenance not found: %w", err)
	}
	if m.Status != model.MaintenanceStatusScheduled {
		return fmt.Errorf("cannot start maintenance in status %s", m.Status)
	}
	fan, err := svc.fanStore.GetByID(m.FanID)
	if err != nil {
		return fmt.Errorf("fan not found: %w", err)
	}
	if fan.Status == model.FanStatusRunning {
		return fmt.Errorf("cannot perform maintenance on running fan")
	}
	return svc.maintStore.UpdateStatus(id, model.MaintenanceStatusInProgress)
}

func (svc *MaintenanceService) Complete(id, notes string) error {
	m, err := svc.maintStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("maintenance not found: %w", err)
	}
	if m.Status != model.MaintenanceStatusInProgress {
		return fmt.Errorf("cannot complete maintenance in status %s", m.Status)
	}
	m.Notes = notes
	m.Status = model.MaintenanceStatusCompleted
	if err := svc.maintStore.Update(m); err != nil {
		return fmt.Errorf("update maintenance: %w", err)
	}
	if err := svc.fanStore.TouchServiceDate(m.FanID, time.Now()); err != nil {
		return fmt.Errorf("touch service date: %w", err)
	}
	return nil
}

func (svc *MaintenanceService) Cancel(id, reason string) error {
	m, err := svc.maintStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("maintenance not found: %w", err)
	}
	if m.Status == model.MaintenanceStatusCompleted {
		return fmt.Errorf("cannot cancel completed maintenance")
	}
	m.Status = model.MaintenanceStatusCancelled
	m.Notes = fmt.Sprintf("%s [cancelled: %s]", m.Notes, reason)
	return svc.maintStore.Update(m)
}

func (svc *MaintenanceService) GetByID(id string) (*model.Maintenance, error) {
	return svc.maintStore.GetByID(id)
}

func (svc *MaintenanceService) ListScheduled() ([]*model.Maintenance, error) {
	return svc.maintStore.ListScheduled()
}

func (svc *MaintenanceService) ListByFan(fanID string) ([]*model.Maintenance, error) {
	return svc.maintStore.ListByFan(fanID)
}

func (svc *MaintenanceService) ListByArea(areaID string) ([]*model.Maintenance, error) {
	return svc.maintStore.ListByArea(areaID)
}

func (svc *MaintenanceService) ListByStatus(status model.MaintenanceStatus) ([]*model.Maintenance, error) {
	return svc.maintStore.ListByStatus(status)
}

func (svc *MaintenanceService) ListOverdue() ([]*model.Maintenance, error) {
	return svc.maintStore.ListOverdue(time.Now())
}

func (svc *MaintenanceService) BatchSchedule(items []*model.Maintenance) (int, error) {
	for _, m := range items {
		if m.FanID == "" {
			return 0, fmt.Errorf("fan id is required")
		}
		if m.AreaID == "" {
			return 0, fmt.Errorf("area id is required")
		}
		if m.Technician == "" {
			return 0, fmt.Errorf("technician is required")
		}
		if m.Type == "" {
			m.Type = model.MaintenanceTypeRoutine
		}
		if m.ScheduledAt.IsZero() {
			m.ScheduledAt = time.Now().Add(72 * time.Hour)
		}
	}
	if err := svc.maintStore.BatchCreate(items); err != nil {
		return 0, fmt.Errorf("batch create: %w", err)
	}
	return len(items), nil
}

func (svc *MaintenanceService) Delete(id string) error {
	return svc.maintStore.Delete(id)
}

func (svc *MaintenanceService) GetMaintenanceSchedule(fanID string) (*model.MaintenanceSchedule, error) {
	fan, err := svc.fanStore.GetByID(fanID)
	if err != nil {
		return nil, fmt.Errorf("fan not found: %w", err)
	}
	prevDate, err := svc.maintStore.GetLastMaintenanceDate(fanID)
	if err != nil {
		return nil, fmt.Errorf("get last maintenance: %w", err)
	}
	interval := 720 * time.Hour
	if prevDate.IsZero() {
		prevDate = fan.InstalledAt
	}
	nextDate := prevDate.Add(interval)
	return &model.MaintenanceSchedule{
		FanID:     fanID,
		NextDate:  nextDate,
		Type:      model.MaintenanceTypeRoutine,
		Interval:  interval,
		LastDate:  prevDate,
		IsOverdue: nextDate.Before(time.Now()),
	}, nil
}
