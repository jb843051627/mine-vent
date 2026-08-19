package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"mine-vent/internal/cache"
	"mine-vent/internal/service"
	"mine-vent/internal/store"
)

type Router struct {
	mux        *http.ServeMux
	store      *store.Store
	sensorSvc  *service.SensorService
	readingSvc *service.ReadingService
	fanSvc     *service.FanService
	schedSvc   *service.ScheduleService
	alertSvc   *service.AlertService
	maintSvc   *service.MaintenanceService
}

func NewRouter(
	dbPath string,
) (*Router, error) {
	s, err := store.NewStore(dbPath)
	if err != nil {
		return nil, err
	}
	sensorStore := store.NewSensorStore(s)
	readingStore := store.NewReadingStore(s)
	fanStore := store.NewFanStore(s)
	scheduleStore := store.NewScheduleStore(s)
	alertStore := store.NewAlertStore(s)
	maintStore := store.NewMaintenanceStore(s)

	readingCache := cache.NewReadingCache()

	r := &Router{
		mux:        http.NewServeMux(),
		store:      s,
		sensorSvc:  service.NewSensorService(sensorStore),
		readingSvc: service.NewReadingService(readingStore, sensorStore, readingCache),
		fanSvc:     service.NewFanService(fanStore, maintStore, readingStore),
		schedSvc:   service.NewScheduleService(scheduleStore, fanStore, maintStore),
		alertSvc:   service.NewAlertService(alertStore, sensorStore),
		maintSvc:   service.NewMaintenanceService(maintStore, fanStore),
	}
	r.registerRoutes()
	return r, nil
}

func (r *Router) registerRoutes() {
	r.mux.HandleFunc("/api/sensors", r.handleSensors)
	r.mux.HandleFunc("/api/sensors/", r.handleSensors)
	r.mux.HandleFunc("/api/readings", r.handleReadings)
	r.mux.HandleFunc("/api/readings/", r.handleReadings)
	r.mux.HandleFunc("/api/fans", r.handleFans)
	r.mux.HandleFunc("/api/fans/", r.handleFans)
	r.mux.HandleFunc("/api/schedules", r.handleSchedules)
	r.mux.HandleFunc("/api/schedules/", r.handleSchedules)
	r.mux.HandleFunc("/api/alerts", r.handleAlerts)
	r.mux.HandleFunc("/api/alerts/", r.handleAlerts)
	r.mux.HandleFunc("/api/maintenance", r.handleMaintenance)
	r.mux.HandleFunc("/api/maintenance/", r.handleMaintenance)
	r.mux.HandleFunc("/api/dashboard", r.handleDashboard)
	r.mux.HandleFunc("/", r.handleWeb)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) Close() error {
	return r.store.Close()
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func parseID(path, prefix string) string {
	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.Split(trimmed, "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return ""
}
