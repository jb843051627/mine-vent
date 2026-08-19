package handler

import (
	"net/http"
	"os"
	"path/filepath"
)

func (r *Router) handleDashboard(w http.ResponseWriter, req *http.Request) {
	sensors, _ := r.sensorSvc.ListAll()
	fans, _ := r.fanSvc.ListAll()
	alerts, _ := r.alertSvc.ListActive()
	maints, _ := r.maintSvc.ListScheduled()
	fanCounts, _ := r.fanSvc.GetFanCountByStatus()
	schedCounts, _ := r.schedSvc.GetScheduleCountByStatus()
	alertSummary, _ := r.alertSvc.GetAlertSummary()

	dashboard := map[string]interface{}{
		"total_sensors":       len(sensors),
		"total_fans":          len(fans),
		"active_alerts":       len(alerts),
		"scheduled_maintenance": len(maints),
		"fan_status_counts":   fanCounts,
		"schedule_status_counts": schedCounts,
		"alert_summary":      alertSummary,
	}
	writeJSON(w, http.StatusOK, dashboard)
}

func (r *Router) handleWeb(w http.ResponseWriter, req *http.Request) {
	if len(req.URL.Path) > 5 && (req.URL.Path[len(req.URL.Path)-5:] == ".html" ||
		req.URL.Path[len(req.URL.Path)-3:] == ".js" ||
		req.URL.Path[len(req.URL.Path)-4:] == ".css") {
		webDir := "web"
		if _, err := os.Stat(webDir); os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "web directory not found")
			return
		}
		filePath := filepath.Join(webDir, req.URL.Path[1:])
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}
		http.ServeFile(w, req, filePath)
		return
	}
	http.ServeFile(w, req, "web/index.html")
}
