package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"mine-vent/internal/model"
)

func (r *Router) handleFans(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		areaID := req.URL.Query().Get("area")
		if areaID != "" {
			fans, err := r.fanSvc.ListByArea(areaID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, fans)
			return
		}
		fans, err := r.fanSvc.ListAll()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, fans)

	case http.MethodPost:
		var fan model.Fan
		if err := json.NewDecoder(req.Body).Decode(&fan); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := r.fanSvc.Register(&fan); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, fan)

	case http.MethodPut:
		parts := strings.Split(strings.TrimPrefix(req.URL.Path, "/api/fans/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			writeError(w, http.StatusBadRequest, "fan id is required")
			return
		}
		fanID := parts[0]
		if len(parts) > 1 {
			action := parts[1]
			switch action {
			case "start":
				var body struct{ RPM int }
				json.NewDecoder(req.Body).Decode(&body)
				if err := r.fanSvc.StartFan(fanID, body.RPM); err != nil {
					writeError(w, http.StatusBadRequest, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
			case "stop":
				if err := r.fanSvc.StopFan(fanID); err != nil {
					writeError(w, http.StatusBadRequest, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
			case "fault":
				if err := r.fanSvc.SetFault(fanID); err != nil {
					writeError(w, http.StatusBadRequest, err.Error())
					return
				}
				writeJSON(w, http.StatusOK, map[string]string{"status": "fault"})
			default:
				writeError(w, http.StatusBadRequest, "unknown action: "+action)
			}
			return
		}
		var fan model.Fan
		if err := json.NewDecoder(req.Body).Decode(&fan); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		fan.ID = fanID
		if err := r.fanSvc.Update(&fan); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, fan)

	case http.MethodDelete:
		id := strings.TrimPrefix(req.URL.Path, "/api/fans/")
		if id == "" {
			writeError(w, http.StatusBadRequest, "fan id is required")
			return
		}
		if err := r.fanSvc.Delete(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
