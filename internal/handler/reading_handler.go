package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mine-vent/internal/model"
)

func (r *Router) handleReadings(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		pathID := strings.TrimPrefix(req.URL.Path, "/api/readings/")
		if pathID != "" && pathID != "/api/readings" {
			if pathID == "latest" {
				sensorID := req.URL.Query().Get("sensor")
				if sensorID == "" {
					writeError(w, http.StatusBadRequest, "sensor parameter is required")
					return
				}
				reading, err := r.readingSvc.GetLatest(sensorID)
				if err != nil {
					writeError(w, http.StatusNotFound, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, reading)
				return
			}
			reading, err := r.readingSvc.GetByID(pathID)
			if err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, reading)
			return
		}
		areaID := req.URL.Query().Get("area")
		limitStr := req.URL.Query().Get("limit")
		limit, _ := strconv.Atoi(limitStr)
		if limit <= 0 {
			limit = 100
		}
		startStr := req.URL.Query().Get("start")
		endStr := req.URL.Query().Get("end")
		if startStr != "" && endStr != "" {
			start, err1 := time.Parse(time.RFC3339, startStr)
			end, err2 := time.Parse(time.RFC3339, endStr)
			if err1 != nil || err2 != nil {
				writeError(w, http.StatusBadRequest, "invalid time format")
				return
			}
			readings, err := r.readingSvc.ListByTimeRange(start, end, limit)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, readings)
			return
		}
		if areaID != "" {
			readings, err := r.readingSvc.ListByArea(areaID, limit)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, readings)
			return
		}
		sensorID := req.URL.Query().Get("sensor")
		if sensorID != "" {
			readings, err := r.readingSvc.ListBySensor(sensorID, limit)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, readings)
			return
		}
		writeError(w, http.StatusBadRequest, "sensor or area parameter is required")

	case http.MethodPost:
		body, err := io.ReadAll(req.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read body")
			return
		}
		var batch model.ReadingBatch
		if jsonErr := json.Unmarshal(body, &batch); jsonErr == nil && len(batch.Readings) > 0 {
			processed, err := r.readingSvc.BatchIngest(req.Context(), &batch)
			if err != nil {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"processed": processed,
					"error":     err.Error(),
				})
				return
			}
			writeJSON(w, http.StatusCreated, map[string]int{"processed": processed})
			return
		}
		var reading model.Reading
		if err := json.Unmarshal(body, &reading); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := r.readingSvc.RecordReading(&reading); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, reading)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
