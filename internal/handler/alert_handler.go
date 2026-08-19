package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"mine-vent/internal/model"
)

func (r *Router) handleAlerts(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		pathID := strings.TrimPrefix(req.URL.Path, "/api/alerts/")
		if pathID != "" && pathID != "/api/alerts" {
			if pathID == "summary" {
				summary, err := r.alertSvc.GetAlertSummary()
				if err != nil {
					writeError(w, http.StatusInternalServerError, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, summary)
				return
			}
			if pathID == "active" {
				alerts, err := r.alertSvc.ListActive()
				if err != nil {
					writeError(w, http.StatusInternalServerError, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, alerts)
				return
			}
			alert, err := r.alertSvc.GetByID(pathID)
			if err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, alert)
			return
		}
		areaID := req.URL.Query().Get("area")
		sensorID := req.URL.Query().Get("sensor")
		if areaID != "" {
			alerts, err := r.alertSvc.ListByArea(areaID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, alerts)
			return
		}
		if sensorID != "" {
			alerts, err := r.alertSvc.ListBySensor(sensorID, 50)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, alerts)
			return
		}
		alerts, err := r.alertSvc.ListActive()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, alerts)

	case http.MethodPost:
		var alert model.Alert
		if err := json.NewDecoder(req.Body).Decode(&alert); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := r.alertSvc.Create(&alert); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, alert)

	case http.MethodPut:
		parts := strings.Split(strings.TrimPrefix(req.URL.Path, "/api/alerts/"), "/")
		if len(parts) < 2 || parts[0] == "" {
			writeError(w, http.StatusBadRequest, "alert id and action are required")
			return
		}
		alertID := parts[0]
		action := parts[1]
		switch action {
		case "ack":
			var body struct{ By string }
			json.NewDecoder(req.Body).Decode(&body)
			if body.By == "" {
				body.By = "system"
			}
			if err := r.alertSvc.Acknowledge(alertID, body.By); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "acknowledged"})
		case "resolve":
			if err := r.alertSvc.Resolve(alertID); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
		case "suppress":
			var body struct{ Reason string }
			json.NewDecoder(req.Body).Decode(&body)
			if err := r.alertSvc.Suppress(alertID, body.Reason); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "suppressed"})
		default:
			writeError(w, http.StatusBadRequest, "unknown action: "+action)
		}

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
