package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"myapp/internal/http/dto"
)

func isValidTicketType(value string) bool {
	allowed := map[string]bool{
		"service_down":           true,
		"high_cpu_usage":         true,
		"memory_warning":         true,
		"disk_space_alert":       true,
		"network_latency":        true,
		"ssl_certificate_expiry": true,
		"database_connection":    true,
		"host_unreachable":       true,
	}
	return allowed[value]
}

func isValidTicketSeverity(value string) bool {
	return value == "warning" || value == "critical"
}

func isValidTicketSource(value string) bool {
	return value == "email" || value == "portal" || value == "communication"
}

func CreateTicketHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			apiFail(w, http.StatusMethodNotAllowed, map[string]string{
				"method": "method not allowed",
			})
			return
		}

		var req dto.CreateTicketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apiFail(w, http.StatusBadRequest, map[string]string{
				"body": "invalid json body",
			})
			return
		}

		errors := map[string]string{}

		if strings.TrimSpace(req.TicketID) == "" {
			errors["ticket_id"] = "ticket_id is required"
		}
		if strings.TrimSpace(req.Title) == "" {
			errors["title"] = "title is required"
		}
		if !isValidTicketType(req.Type) {
			errors["type"] = "invalid type"
		}
		if !isValidTicketSeverity(req.Severity) {
			errors["severity"] = "invalid severity"
		}
		if !isValidTicketSource(req.Source) {
			errors["source"] = "invalid source"
		}
		if strings.TrimSpace(req.TicketTimeIn) == "" {
			errors["ticket_time_in"] = "ticket_time_in is required"
		}
		if strings.TrimSpace(req.FirstResponseAt) == "" {
			errors["first_response_at"] = "first_response_at is required"
		}
		if strings.TrimSpace(req.TicketTimeOut) == "" {
			errors["ticket_time_out"] = "ticket_time_out is required"
		}
		if strings.TrimSpace(req.DetailedDescription) == "" {
			errors["detailed_description"] = "detailed_description is required"
		}
		if req.CreatedByUserID <= 0 {
			errors["created_by_user_id"] = "created_by_user_id is required"
		}
		if req.HandledByUserID <= 0 {
			errors["handled_by_user_id"] = "handled_by_user_id is required"
		}

		if len(errors) > 0 {
			apiFail(w, http.StatusBadRequest, errors)
			return
		}

		query := `
			INSERT INTO tickets (
				ticket_id,
				title,
				type,
				severity,
				source,
				is_escalated,
				auto_resolved,
				detailed_description,
				created_by_user_id,
				handled_by_user_id,
				ticket_time_in,
				first_response_at,
				ticket_time_out,
				created_at,
				updated_at
			)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				date_trunc('day', CURRENT_TIMESTAMP) + $11::time,
				date_trunc('day', CURRENT_TIMESTAMP) + $12::time,
				date_trunc('day', CURRENT_TIMESTAMP) + $13::time,
				NOW(),
				NOW()
			)
			RETURNING
				id,
				ticket_id,
				title,
				type,
				severity,
				source,
				is_escalated,
				auto_resolved,
				detailed_description,
				created_by_user_id,
				handled_by_user_id,
				ticket_time_in,
				first_response_at,
				ticket_time_out,
				created_at,
				updated_at
		`

		var ticket dto.TicketResponse

		err := db.QueryRow(
			query,
			req.TicketID,
			req.Title,
			req.Type,
			req.Severity,
			req.Source,
			req.IsEscalated,
			req.AutoResolved,
			req.DetailedDescription,
			req.CreatedByUserID,
			req.HandledByUserID,
			req.TicketTimeIn,
			req.FirstResponseAt,
			req.TicketTimeOut,
		).Scan(
			&ticket.ID,
			&ticket.TicketID,
			&ticket.Title,
			&ticket.Type,
			&ticket.Severity,
			&ticket.Source,
			&ticket.IsEscalated,
			&ticket.AutoResolved,
			&ticket.DetailedDescription,
			&ticket.CreatedByUserID,
			&ticket.HandledByUserID,
			&ticket.TicketTimeIn,
			&ticket.FirstResponseAt,
			&ticket.TicketTimeOut,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			apiFail(w, http.StatusInternalServerError, map[string]string{
				"database": err.Error(),
			})
			return
		}

		apiOK(w, map[string]any{
			"ticket": ticket,
		})
	})
}

func ListTicketsHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			apiFail(w, http.StatusMethodNotAllowed, map[string]string{
				"method": "method not allowed",
			})
			return
		}

		query := `
			SELECT
				id,
				ticket_id,
				title,
				type,
				severity,
				source,
				is_escalated,
				auto_resolved,
				detailed_description,
				created_by_user_id,
				handled_by_user_id,
				ticket_time_in,
				first_response_at,
				ticket_time_out,
				created_at,
				updated_at
			FROM tickets
			ORDER BY id DESC
		`

		rows, err := db.Query(query)
		if err != nil {
			apiFail(w, http.StatusInternalServerError, map[string]string{
				"database": err.Error(),
			})
			return
		}
		defer rows.Close()

		var tickets []dto.TicketResponse

		for rows.Next() {
			var t dto.TicketResponse

			err := rows.Scan(
				&t.ID,
				&t.TicketID,
				&t.Title,
				&t.Type,
				&t.Severity,
				&t.Source,
				&t.IsEscalated,
				&t.AutoResolved,
				&t.DetailedDescription,
				&t.CreatedByUserID,
				&t.HandledByUserID,
				&t.TicketTimeIn,
				&t.FirstResponseAt,
				&t.TicketTimeOut,
				&t.CreatedAt,
				&t.UpdatedAt,
			)
			if err != nil {
				apiFail(w, http.StatusInternalServerError, map[string]string{
					"database": err.Error(),
				})
				return
			}

			tickets = append(tickets, t)
		}

		apiOK(w, map[string]any{
			"tickets": tickets,
		})
	})
}

func TicketDetailByTicketIDHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			apiFail(w, http.StatusMethodNotAllowed, map[string]string{
				"method": "method not allowed",
			})
			return
		}

		ticketID := strings.TrimPrefix(r.URL.Path, "/ticket/detail/")
		if strings.TrimSpace(ticketID) == "" {
			apiFail(w, http.StatusBadRequest, map[string]string{
				"ticket_id": "ticket_id is required",
			})
			return
		}

		query := `
			SELECT
				id,
				ticket_id,
				title,
				type,
				severity,
				source,
				is_escalated,
				auto_resolved,
				detailed_description,
				created_by_user_id,
				handled_by_user_id,
				ticket_time_in,
				first_response_at,
				ticket_time_out,
				created_at,
				updated_at
			FROM tickets
			WHERE ticket_id = $1
		`

		var ticket dto.TicketResponse

		err := db.QueryRow(query, ticketID).Scan(
			&ticket.ID,
			&ticket.TicketID,
			&ticket.Title,
			&ticket.Type,
			&ticket.Severity,
			&ticket.Source,
			&ticket.IsEscalated,
			&ticket.AutoResolved,
			&ticket.DetailedDescription,
			&ticket.CreatedByUserID,
			&ticket.HandledByUserID,
			&ticket.TicketTimeIn,
			&ticket.FirstResponseAt,
			&ticket.TicketTimeOut,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err == sql.ErrNoRows {
			apiFail(w, http.StatusNotFound, map[string]string{
				"ticket": "ticket not found",
			})
			return
		}
		if err != nil {
			apiFail(w, http.StatusInternalServerError, map[string]string{
				"database": err.Error(),
			})
			return
		}

		apiOK(w, map[string]any{
			"ticket": ticket,
		})
	})
}

func ResolveTicketHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			apiFail(w, http.StatusMethodNotAllowed, map[string]string{
				"method": "method not allowed",
			})
			return
		}

		var req dto.ResolveTicketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apiFail(w, http.StatusBadRequest, map[string]string{
				"body": "invalid json body",
			})
			return
		}

		if strings.TrimSpace(req.TicketID) == "" {
			apiFail(w, http.StatusBadRequest, map[string]string{
				"ticket_id": "ticket_id is required",
			})
			return
		}

		query := `
			UPDATE tickets
			SET updated_at = NOW()
			WHERE ticket_id = $1
			RETURNING
				id,
				ticket_id,
				title,
				type,
				severity,
				source,
				is_escalated,
				auto_resolved,
				detailed_description,
				created_by_user_id,
				handled_by_user_id,
				ticket_time_in,
				first_response_at,
				ticket_time_out,
				created_at,
				updated_at
		`

		var ticket dto.TicketResponse

		err := db.QueryRow(query, req.TicketID).Scan(
			&ticket.ID,
			&ticket.TicketID,
			&ticket.Title,
			&ticket.Type,
			&ticket.Severity,
			&ticket.Source,
			&ticket.IsEscalated,
			&ticket.AutoResolved,
			&ticket.DetailedDescription,
			&ticket.CreatedByUserID,
			&ticket.HandledByUserID,
			&ticket.TicketTimeIn,
			&ticket.FirstResponseAt,
			&ticket.TicketTimeOut,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err == sql.ErrNoRows {
			apiFail(w, http.StatusNotFound, map[string]string{
				"ticket": "ticket not found",
			})
			return
		}
		if err != nil {
			apiFail(w, http.StatusInternalServerError, map[string]string{
				"database": err.Error(),
			})
			return
		}

		apiOK(w, map[string]any{
			"ticket": ticket,
		})
	})
}