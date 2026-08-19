package service

import (
	"fmt"
	"time"

	"mine-vent/internal/model"
	"mine-vent/internal/store"
)

type ScheduleService struct {
	scheduleStore *store.ScheduleStore
	fanStore      *store.FanStore
	maintStore    *store.MaintenanceStore
}

func NewScheduleService(ss *store.ScheduleStore, fs *store.FanStore, ms *store.MaintenanceStore) *ScheduleService {
	return &ScheduleService{
		scheduleStore: ss,
		fanStore:      fs,
		maintStore:    ms,
	}
}

func (svc *ScheduleService) Create(schedule *model.Schedule) error {
	if schedule.ID == "" {
		schedule.ID = fmt.Sprintf("sch-%d", time.Now().UnixNano())
	}
	if schedule.Name == "" {
		return fmt.Errorf("schedule name is required")
	}
	if schedule.FanID == "" {
		return fmt.Errorf("fan id is required")
	}
	if schedule.AreaID == "" {
		return fmt.Errorf("area id is required")
	}
	if err := svc.validateSchedule(schedule); err != nil {
		return fmt.Errorf("validate schedule: %w", err)
	}
	schedule.Status = model.ScheduleStatusDraft
	return svc.scheduleStore.Create(schedule)
}

func (svc *ScheduleService) validateSchedule(s *model.Schedule) error {
	if s.StartTime.IsZero() {
		return fmt.Errorf("start time is required")
	}
	if !s.EndTime.IsZero() && s.EndTime.Before(s.StartTime) {
		return fmt.Errorf("end time cannot be before start time")
	}
	if s.TargetRPM < 0 {
		return fmt.Errorf("target rpm cannot be negative")
	}
	if s.Priority < 0 || s.Priority > 10 {
		return fmt.Errorf("priority must be between 0 and 10")
	}
	if s.Type == "" {
		return fmt.Errorf("schedule type is required")
	}
	return nil
}

func (svc *ScheduleService) ActivateSchedule(id string) error {
	schedule, err := svc.scheduleStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("schedule not found: %w", err)
	}
	if schedule.Status != model.ScheduleStatusDraft && schedule.Status != model.ScheduleStatusPaused {
		return fmt.Errorf("cannot activate schedule in status %s", schedule.Status)
	}
	if err := svc.validateSchedule(schedule); err != nil {
		return fmt.Errorf("validate schedule: %w", err)
	}
	fan, err := svc.fanStore.GetByID(schedule.FanID)
	if err != nil {
		return fmt.Errorf("fan not found: %w", err)
	}
	if fan.Status == model.FanStatusFault {
		return fmt.Errorf("cannot assign schedule to fan in fault state")
	}
	schedule.Status = model.ScheduleStatusActive
	return svc.scheduleStore.Update(schedule)
}

func (svc *ScheduleService) AssignFan(scheduleID, fanID string) error {
	schedule, err := svc.scheduleStore.GetByID(scheduleID)
	if err != nil {
		return fmt.Errorf("schedule not found: %w", err)
	}
	fan, err := svc.fanStore.GetByID(fanID)
	if err != nil {
		return fmt.Errorf("fan not found: %w", err)
	}
	if fan == nil {
		return fmt.Errorf("fan %s not found", fanID)
	}
	if fan.Status == model.FanStatusFault {
		return fmt.Errorf("cannot assign fault fan to schedule")
	}
	schedule.FanID = fanID
	return svc.scheduleStore.Update(schedule)
}

func (svc *ScheduleService) Pause(id string) error {
	return svc.scheduleStore.UpdateStatus(id, model.ScheduleStatusPaused)
}

func (svc *ScheduleService) Complete(id string) error {
	return svc.scheduleStore.UpdateStatus(id, model.ScheduleStatusCompleted)
}

func (svc *ScheduleService) GetByID(id string) (*model.Schedule, error) {
	schedule, err := svc.scheduleStore.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("schedule not found: %w", err)
	}
	return schedule, nil
}

func (svc *ScheduleService) ListActive() ([]*model.Schedule, error) {
	return svc.scheduleStore.ListActive()
}

func (svc *ScheduleService) ListByStatus(status model.ScheduleStatus) ([]*model.Schedule, error) {
	return svc.scheduleStore.ListByStatus(status)
}

func (svc *ScheduleService) ListByFan(fanID string) ([]*model.Schedule, error) {
	return svc.scheduleStore.ListByFan(fanID)
}

func (svc *ScheduleService) ListByArea(areaID string) ([]*model.Schedule, error) {
	return svc.scheduleStore.ListByArea(areaID)
}

func (svc *ScheduleService) Delete(id string) error {
	return svc.scheduleStore.Delete(id)
}

func (svc *ScheduleService) GetNextMaintenanceWindow(fanID string) (*model.MaintenanceWindow, error) {
	fan, err := svc.fanStore.GetByID(fanID)
	if err != nil {
		return nil, fmt.Errorf("fan not found: %w", err)
	}
	lastMaint, err := svc.maintStore.GetLastMaintenanceDate(fanID)
	if err != nil {
		return nil, fmt.Errorf("get maintenance history: %w", err)
	}
	now := time.Now()
	interval := 720 * time.Hour
	if lastMaint.IsZero() {
		lastMaint = fan.InstalledAt
	}
	nextDate := lastMaint.Add(interval)
	if nextDate.Before(now) {
		nextDate = now.Add(24 * time.Hour)
	}
	window := &model.MaintenanceWindow{
		StartTime: nextDate,
		EndTime:   nextDate.Add(4 * time.Hour),
		Duration:  4 * time.Hour,
		Reason:    "routine maintenance",
		FanID:     fanID,
	}
	return window, nil
}

func (svc *ScheduleService) AutoRotateFans(areaID string) (*model.FanRotationPlan, error) {
	fans, err := svc.fanStore.ListByArea(areaID)
	if err != nil {
		return nil, fmt.Errorf("list fans: %w", err)
	}
	if len(fans) < 2 {
		return nil, fmt.Errorf("need at least 2 fans for rotation")
	}
	var activeFan, standbyFan *model.Fan
	for _, f := range fans {
		if f.Status == model.FanStatusRunning && activeFan == nil {
			activeFan = f
		}
		if f.Status == model.FanStatusStopped && standbyFan == nil {
			standbyFan = f
		}
	}
	if activeFan == nil || standbyFan == nil {
		return nil, fmt.Errorf("no valid fan rotation pair found")
	}
	plan := &model.FanRotationPlan{
		ActiveFanID:    activeFan.ID,
		StandbyFanID:   standbyFan.ID,
		NextRotationAt: time.Now().Add(168 * time.Hour),
		Reason:         "weekly rotation",
	}
	return plan, nil
}

func (svc *ScheduleService) GetScheduleCountByStatus() (map[model.ScheduleStatus]int, error) {
	return svc.scheduleStore.CountByStatus()
}
