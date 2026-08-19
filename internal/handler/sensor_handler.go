package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"mine-vent/internal/model"
)

func (r *Router) handleSensors(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		areaID := req.URL.Query().Get("area")
		sensorType := req.URL.Query().Get("type")
		if sensorType != "" {
			sensors, err := r.sensorSvc.ListByType(model.SensorType(sensorType))
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, sensors)
			return
		}
		if areaID != "" {
			sensors, err := r.sensorSvc.ListByArea(areaID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, sensors)
			return
		}
		sensors, err := r.sensorSvc.ListAll()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, sensors)

	case http.MethodPost:
		var sensor model.Sensor
		if err := json.NewDecoder(req.Body).Decode(&sensor); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := r.sensorSvc.Register(&sensor); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, sensor)

	case http.MethodPut:
		id := strings.TrimPrefix(req.URL.Path, "/api/sensors/")
		if id == "" {
			writeError(w, http.StatusBadRequest, "sensor id is required")
			return
		}
		var sensor model.Sensor
		if err := json.NewDecoder(req.Body).Decode(&sensor); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		sensor.ID = id
		if err := r.sensorSvc.Update(&sensor); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, sensor)

	case http.MethodDelete:
		id := strings.TrimPrefix(req.URL.Path, "/api/sensors/")
		if id == "" {
			writeError(w, http.StatusBadRequest, "sensor id is required")
			return
		}
		if err := r.sensorSvc.Delete(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
