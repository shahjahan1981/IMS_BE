package handlers

import "net/http"

// Standard error shape (NO messages)
func fail(w http.ResponseWriter, code int, errs map[string]string) {
	writeJSON(w, code, map[string]any{
		"status": false,
		"errors": errs,
	})
}