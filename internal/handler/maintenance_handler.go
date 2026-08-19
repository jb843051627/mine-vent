package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"mine-vent/internal/model"
)

func (r *Router) handleMaintenance(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		pathID := strings.TrimPrefix(req.URL.Path, "/api/maintenance/")
		if pathID != "" && pathID != "/api/maintenance" {
			if pathID == "scheduled" {
				items, err := r.maintSvc.ListScheduled()
				if err != nil {
					writeError(w, http.StatusInternalServerError, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, items)
				return
			}
			if pathID == "overdue" {
				items, err := r.maintSvc.ListOverdue()
				if err != nil {
					writeError(w, http.StatusInternalServerError, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, items)
				return
			}
			if strings.HasPrefix(pathID, "fan/") {
				fanID := strings.TrimPrefix(pathID, "fan/")
				items, err := r.maintSvc.ListByFan(fanID)
				if err != nil {
					writeError(w, http.StatusInternalServerError, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, items)
				return
			}
			if strings.HasPrefix(pathID, "area/") {
				areaID := strings.TrimPrefix(pathID, "area/")
				items, err := r.maintSvc.ListByArea(areaID)
				if err != nil {
					writeError(w, http.StatusInternalServerError, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, items)
				return
			}
			m, err := r.maintSvc.GetByID(pathID)
			if err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, m)
			return
		}
		status := req.URL.Query().Get("status")
		if status != "" {
			items, err := r.maintSvc.ListByStatus(model.MaintenanceStatus(status))
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, items)
			return
		}
		items, err := r.maintSvc.ListScheduled()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, items)

	case http.MethodPost:
		var m model.Maintenance
		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := r.maintSvc.Schedule(&m); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, m)

	case http.MethodPut:
		parts := strings.Split(strings.TrimPrefix(req.URL.Path, "/api/maintenance/"), "/")
		if len(parts) < 2 || parts[0] == "" {
			writeError(w, http.StatusBadRequest, "maintenance id and action are required")
			return
		}
		maintID := parts[0]
		action := parts[1]
		switch action {
		case "start":
			if err := r.maintSvc.Start(maintID); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "in_progress"})
		case "complete":
			var body struct{ Notes string }
			json.NewDecoder(req.Body).Decode(&body)
			if err := r.maintSvc.Complete(maintID, body.Notes); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
		case "cancel":
			var body struct{ Reason string }
			json.NewDecoder(req.Body).Decode(&body)
			if err := r.maintSvc.Cancel(maintID, body.Reason); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
		default:
			writeError(w, http.StatusBadRequest, "unknown action: "+action)
		}

	case http.MethodDelete:
		id := strings.TrimPrefix(req.URL.Path, "/api/maintenance/")
		if id == "" {
			writeError(w, http.StatusBadRequest, "maintenance id is required")
			return
		}
		if err := r.maintSvc.Delete(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
