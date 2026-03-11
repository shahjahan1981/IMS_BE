package handlers

import "strings"

func isValidRole(role string) bool {
	return role == "admin" || role == "supervisor" || role == "engineer"
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}