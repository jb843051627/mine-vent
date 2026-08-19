package service

import (
	"fmt"
	"math"
	"sort"
	"time"

	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

type FanService struct {
	fanStore       *store.FanStore
	maintStore     *store.MaintenanceStore
	readingStore   *store.ReadingStore
}

func NewFanService(fs *store.FanStore, ms *store.MaintenanceStore, rs *store.ReadingStore) *FanService {
	return &FanService{
		fanStore:     fs,
		maintStore:   ms,
		readingStore: rs,
	}
}

func (svc *FanService) Register(fan *model.Fan) error {
	if fan.ID == "" {
		fan.ID = fmt.Sprintf("fan-%d", time.Now().UnixNano())
	}
	if fan.Name == "" {
		return fmt.Errorf("fan name is required")
	}
	if fan.AreaID == "" {
		return fmt.Errorf("area id is required")
	}
	if fan.Capacity <= 0 {
		return fmt.Errorf("capacity must be positive")
	}
	fan.Status = model.FanStatusStopped
	fan.IsActive = true
	if fan.InstalledAt.IsZero() {
		fan.InstalledAt = time.Now()
	}
	fan.LastServiceAt = fan.InstalledAt
	return svc.fanStore.Create(fan)
}

func (svc *FanService) GetByID(id string) (*model.Fan, error) {
	fan, err := svc.fanStore.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get fan: %w", err)
	}
	return fan, nil
}

func (svc *FanService) ListAll() ([]*model.Fan, error) {
	fans, err := svc.fanStore.ListAll()
	if err != nil {
		return nil, fmt.Errorf("list fans: %w", err)
	}
	return fans, nil
}

func (svc *FanService) ListByArea(areaID string) ([]*model.Fan, error) {
	fans, err := svc.fanStore.ListByArea(areaID)
	if err != nil {
		return nil, fmt.Errorf("list fans by area: %w", err)
	}
	return fans, nil
}

func (svc *FanService) StartFan(id string, targetRPM int) error {
	fan, err := svc.fanStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("fan not found: %w", err)
	}
	if fan.Status == model.FanStatusFault {
		return fmt.Errorf("fan %s is in fault state, cannot start", id)
	}
	if targetRPM <= 0 {
		targetRPM = int(fan.Capacity * 0.8)
	}
	return svc.fanStore.UpdateStatus(id, model.FanStatusRunning, targetRPM)
}

func (svc *FanService) StopFan(id string) error {
	return svc.fanStore.UpdateStatus(id, model.FanStatusStopped, 0)
}

func (svc *FanService) SetFault(id string) error {
	return svc.fanStore.UpdateStatus(id, model.FanStatusFault, 0)
}

func (svc *FanService) Update(fan *model.Fan) error {
	existing, err := svc.fanStore.GetByID(fan.ID)
	if err != nil {
		return fmt.Errorf("fan not found: %w", err)
	}
	if fan.Name == "" {
		fan.Name = existing.Name
	}
	if fan.AreaID == "" {
		fan.AreaID = existing.AreaID
	}
	if fan.Capacity <= 0 {
		fan.Capacity = existing.Capacity
	}
	return svc.fanStore.Update(fan)
}

func (svc *FanService) Delete(id string) error {
	return svc.fanStore.Delete(id)
}

func (svc *FanService) EvaluateFanHealth(fanID string) (*model.FanHealth, error) {
	fan, err := svc.fanStore.GetByID(fanID)
	if err != nil {
		return nil, fmt.Errorf("fan not found: %w", err)
	}
	health := &model.FanHealth{
		FanID:   fan.ID,
		FanName: fan.Name,
	}
	latestReading, err := svc.readingStore.GetLatestBySensor(fanID)
	if err == nil && latestReading != nil {
		health.LastReading = latestReading.Value
	}
	if fan.Capacity > 0 {
		health.Efficiency = float64(fan.CurrentRPM) / fan.Capacity
	}
	lastMaint, err := svc.maintStore.GetLastMaintenanceDate(fanID)
	if err == nil && !lastMaint.IsZero() {
		hours := time.Since(lastMaint).Hours()
		health.UptimeHours = hours
		if hours > 720 {
			health.NeedsService = true
			health.Warnings = append(health.Warnings, "overdue for service")
		}
	} else {
		hours := time.Since(fan.InstalledAt).Hours()
		health.UptimeHours = hours
		if hours > 2160 {
			health.NeedsService = true
			health.Warnings = append(health.Warnings, "never serviced")
		}
	}
	if fan.Status == model.FanStatusFault {
		health.HealthScore = 0
		health.Warnings = append(health.Warnings, "fan in fault state")
	} else if health.Efficiency > 0 {
		health.HealthScore = math.Max(0, 100-(100-health.Efficiency*100))
	} else {
		health.HealthScore = 100
	}
	if !health.NeedsService && len(health.Warnings) == 0 {
		health.HealthScore = 100
	}
	return health, nil
}

func (svc *FanService) GenerateHealthReport(areaID string) ([]*model.FanHealth, error) {
	fans, err := svc.fanStore.ListByArea(areaID)
	if err != nil {
		return nil, fmt.Errorf("list fans: %w", err)
	}
	report := make([]*model.FanHealth, 0, len(fans))
	for _, fan := range fans {
		health, err := svc.EvaluateFanHealth(fan.ID)
		if err != nil {
			continue
		}
		report = append(report, health)
	}
	sort.Slice(report, func(i, j int) bool {
		return report[i].HealthScore < report[j].HealthScore
	})
	return report, nil
}

func (svc *FanService) GetFanCountByStatus() (map[model.FanStatus]int, error) {
	return svc.fanStore.CountByStatus()
}

func (svc *FanService) RotateFans(activeFanID, standbyFanID string) error {
	activeFan, err := svc.fanStore.GetByID(activeFanID)
	if err != nil {
		return fmt.Errorf("active fan not found: %w", err)
	}
	standbyFan, err := svc.fanStore.GetByID(standbyFanID)
	if err != nil {
		return fmt.Errorf("standby fan not found: %w", err)
	}
	if activeFan.Status != model.FanStatusRunning {
		return fmt.Errorf("active fan is not running")
	}
	if standbyFan.Status != model.FanStatusStopped {
		return fmt.Errorf("standby fan is not stopped")
	}
	if err := svc.fanStore.UpdateStatus(standbyFan.ID, model.FanStatusRunning, int(standbyFan.Capacity*0.8)); err != nil {
		return fmt.Errorf("start standby: %w", err)
	}
	if err := svc.fanStore.UpdateStatus(activeFan.ID, model.FanStatusStopped, 0); err != nil {
		return fmt.Errorf("stop active: %w", err)
	}
	return nil
}
