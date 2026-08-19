package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"mine-vent/internal/model"
)

func (r *Router) handleSchedules(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		pathID := strings.TrimPrefix(req.URL.Path, "/api/schedules/")
		if pathID != "" && pathID != "/api/schedules" {
			if strings.HasPrefix(pathID, "fan/") {
				fanID := strings.TrimPrefix(pathID, "fan/")
				schedules, err := r.schedSvc.ListByFan(fanID)
				if err != nil {
					writeError(w, http.StatusInternalServerError, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, schedules)
				return
			}
			if strings.HasPrefix(pathID, "area/") {
				areaID := strings.TrimPrefix(pathID, "area/")
				schedules, err := r.schedSvc.ListByArea(areaID)
				if err != nil {
					writeError(w, http.StatusInternalServerError, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, schedules)
				return
			}
			schedule, err := r.schedSvc.GetByID(pathID)
			if err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, schedule)
			return
		}
		status := req.URL.Query().Get("status")
		if status != "" {
			schedules, err := r.schedSvc.ListByStatus(model.ScheduleStatus(status))
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, schedules)
			return
		}
		schedules, err := r.schedSvc.ListActive()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, schedules)

	case http.MethodPost:
		var schedule model.Schedule
		if err := json.NewDecoder(req.Body).Decode(&schedule); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := r.schedSvc.Create(&schedule); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, schedule)

	case http.MethodPut:
		parts := strings.Split(strings.TrimPrefix(req.URL.Path, "/api/schedules/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			writeError(w, http.StatusBadRequest, "schedule id is required")
			return
		}
		scheduleID := parts[0]
		if len(parts) > 1 {
			action := parts[1]
			switch action {
			case "activate":
				if err := r.schedSvc.ActivateSchedule(scheduleID); err != nil {
					writeError(w, http.StatusBadRequest, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, map[string]string{"status": "active"})
			case "pause":
				if err := r.schedSvc.Pause(scheduleID); err != nil {
					writeError(w, http.StatusBadRequest, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, map[string]string{"status": "paused"})
			case "complete":
				if err := r.schedSvc.Complete(scheduleID); err != nil {
					writeError(w, http.StatusBadRequest, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
			case "assign":
				fanID := req.URL.Query().Get("fan")
				if fanID == "" {
					writeError(w, http.StatusBadRequest, "fan parameter is required")
					return
				}
				if err := r.schedSvc.AssignFan(scheduleID, fanID); err != nil {
					writeError(w, http.StatusBadRequest, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, map[string]string{"status": "assigned"})
			default:
				writeError(w, http.StatusBadRequest, "unknown action: "+action)
			}
			return
		}
		var schedule model.Schedule
		if err := json.NewDecoder(req.Body).Decode(&schedule); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		schedule.ID = scheduleID
		if err := r.schedSvc.Create(&schedule); err != nil {
			writeJSON(w, http.StatusOK, schedule)
			return
		}
		writeJSON(w, http.StatusOK, schedule)

	case http.MethodDelete:
		id := strings.TrimPrefix(req.URL.Path, "/api/schedules/")
		if id == "" {
			writeError(w, http.StatusBadRequest, "schedule id is required")
			return
		}
		if err := r.schedSvc.Delete(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
