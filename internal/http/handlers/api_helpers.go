package handlers

import "net/http"

func apiFail(w http.ResponseWriter, code int, errs map[string]string) {
	writeJSON(w, code, map[string]any{
		"status": false,
		"errors": errs,
	})
}

func apiOK(w http.ResponseWriter, payload map[string]any) {
	if payload == nil {
		payload = map[string]any{}
	}
	payload["status"] = true
	payload["errors"] = map[string]string{}
	writeJSON(w, http.StatusOK, payload)
}