package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"myapp/internal/http/dto"
	"myapp/internal/http/handlers/requests"

	"github.com/lib/pq"
)

var allowedTicketTypes = map[string]bool{
	"service_down":           true,
	"high_cpu_usage":         true,
	"memory_warning":         true,
	"disk_space_alert":       true,
	"network_latency":        true,
	"ssl_certificate_expiry": true,
	"database_connection":    true,
	"host_unreachable":       true,
}

var allowedSeverities = map[string]bool{
	"warning":  true,
	"critical": true,
}

var allowedSources = map[string]bool{
	"email":         true,
	"portal":        true,
	"communication": true,
}

// POST /ticket/create
func CreateTicketHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			apiFail(w, http.StatusMethodNotAllowed, map[string]string{"method": "invalid"})
			return
		}

		var req requests.CreateTicketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apiFail(w, http.StatusBadRequest, map[string]string{"body": "invalid"})
			return
		}

		// normalize
		req.TicketID = strings.TrimSpace(req.TicketID)
		req.Title = strings.TrimSpace(req.Title)
		req.Type = strings.TrimSpace(strings.ToLower(req.Type))
		req.Severity = strings.TrimSpace(strings.ToLower(req.Severity))
		req.Source = strings.TrimSpace(strings.ToLower(req.Source))
		req.DetailedDesc = strings.TrimSpace(req.DetailedDesc)

		// required validations (property-only)
		errs := map[string]string{}
		if req.TicketID == "" {
			errs["ticket_id"] = "required"
		}
		if req.Title == "" {
			errs["title"] = "required"
		}
		if req.Type == "" {
			errs["type"] = "required"
		}
		if req.Severity == "" {
			errs["severity"] = "required"
		}
		if req.Source == "" {
			errs["source"] = "required"
		}
		if req.CreatedByUserID <= 0 {
			errs["created_by_user_id"] = "required"
		}
		if len(errs) > 0 {
			apiFail(w, http.StatusBadRequest, errs)
			return
		}

		// title length (DB is varchar(30))
		if len(req.Title) > 30 {
			apiFail(w, http.StatusBadRequest, map[string]string{"title": "max_length"})
			return
		}

		// allowed values
		if !allowedTicketTypes[req.Type] {
			apiFail(w, http.StatusBadRequest, map[string]string{"type": "invalid"})
			return
		}
		if !allowedSeverities[req.Severity] {
			apiFail(w, http.StatusBadRequest, map[string]string{"severity": "invalid"})
			return
		}
		if !allowedSources[req.Source] {
			apiFail(w, http.StatusBadRequest, map[string]string{"source": "invalid"})
			return
		}

		var detail any = nil
		if req.DetailedDesc != "" {
			detail = req.DetailedDesc
		}

		var handled any = nil
		if req.HandledByUserID != nil && *req.HandledByUserID > 0 {
			handled = *req.HandledByUserID
		}

		// insert
		var newID int64
		err := db.QueryRow(`
			INSERT INTO tickets
				(ticket_id, title, type, severity, source, is_escalated, auto_resolved, detailed_description, created_by_user_id, handled_by_user_id)
			VALUES
				($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			RETURNING id
		`,
			req.TicketID, req.Title, req.Type, req.Severity, req.Source,
			req.IsEscalated, req.AutoResolved, detail, req.CreatedByUserID, handled,
		).Scan(&newID)

		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) {
				// unique violation
				if pqErr.Code == "23505" {
					c := strings.ToLower(pqErr.Constraint + " " + pqErr.Detail)
					if strings.Contains(c, "ticket_id") {
						apiFail(w, http.StatusConflict, map[string]string{"ticket_id": "exists"})
						return
					}
					apiFail(w, http.StatusConflict, map[string]string{"general": "exists"})
					return
				}
				// fk violation
				if pqErr.Code == "23503" {
					c := strings.ToLower(pqErr.Constraint + " " + pqErr.Detail)
					if strings.Contains(c, "created_by") {
						apiFail(w, http.StatusBadRequest, map[string]string{"created_by_user_id": "invalid"})
						return
					}
					if strings.Contains(c, "handled_by") {
						apiFail(w, http.StatusBadRequest, map[string]string{"handled_by_user_id": "invalid"})
						return
					}
					apiFail(w, http.StatusBadRequest, map[string]string{"general": "invalid"})
					return
				}
			}

			apiFail(w, http.StatusInternalServerError, map[string]string{"general": "invalid"})
			return
		}

		// success (no ticket data)
		apiOK(w, map[string]any{
			"successfully_created": true,
			"id":                   newID,
			"ticket_id":            req.TicketID,
		})
	}
}

// GET /tickets/list
func ListTicketsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			apiFail(w, http.StatusMethodNotAllowed, map[string]string{"method": "invalid"})
			return
		}

		rows, err := db.Query(`
			SELECT id, ticket_id, title, type, severity, source,
			       is_escalated, auto_resolved, detailed_description,
			       created_by_user_id, handled_by_user_id,
			       ticket_time_in, ticket_time_out, first_response_at,
			       created_at, updated_at
			FROM tickets
			ORDER BY id DESC
			LIMIT 200
		`)
		if err != nil {
			apiFail(w, http.StatusInternalServerError, map[string]string{"general": "invalid"})
			return
		}
		defer rows.Close()

		out := make([]dto.TicketDTO, 0)
		for rows.Next() {
			var t dto.TicketDTO
			var desc sql.NullString
			var handled sql.NullInt64
			var tout sql.NullTime
			var first sql.NullTime

			if err := rows.Scan(
				&t.ID, &t.TicketID, &t.Title, &t.Type, &t.Severity, &t.Source,
				&t.IsEscalated, &t.AutoResolved, &desc,
				&t.CreatedByUserID, &handled,
				&t.TicketTimeIn, &tout, &first,
				&t.CreatedAt, &t.UpdatedAt,
			); err != nil {
				apiFail(w, http.StatusInternalServerError, map[string]string{"general": "invalid"})
				return
			}

			if desc.Valid {
				v := desc.String
				t.DetailedDesc = &v
			}
			if handled.Valid {
				v := handled.Int64
				t.HandledByUserID = &v
			}
			if tout.Valid {
				v := tout.Time
				t.TicketTimeOut = &v
			}
			if first.Valid {
				v := first.Time
				t.FirstResponseAt = &v
			}

			out = append(out, t)
		}

		apiOK(w, map[string]any{"tickets": out})
	}
}

// GET /ticket/detail/{ticket_id}
// Example: /ticket/detail/TCK-0001
func TicketDetailByTicketIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			apiFail(w, http.StatusMethodNotAllowed, map[string]string{"method": "invalid"})
			return
		}

		ticketID := strings.TrimPrefix(r.URL.Path, "/ticket/detail/")
		ticketID = strings.TrimSpace(ticketID)

		if ticketID == "" {
			apiFail(w, http.StatusBadRequest, map[string]string{"ticket_id": "required"})
			return
		}

		var (
			t       dto.TicketDTO
			desc    sql.NullString
			handled sql.NullInt64
			tout    sql.NullTime
			first   sql.NullTime
		)

		err := db.QueryRow(`
			SELECT id, ticket_id, title, type, severity, source,
			       is_escalated, auto_resolved, detailed_description,
			       created_by_user_id, handled_by_user_id,
			       ticket_time_in, ticket_time_out, first_response_at,
			       created_at, updated_at
			FROM tickets
			WHERE ticket_id = $1
		`, ticketID).Scan(
			&t.ID, &t.TicketID, &t.Title, &t.Type, &t.Severity, &t.Source,
			&t.IsEscalated, &t.AutoResolved, &desc,
			&t.CreatedByUserID, &handled,
			&t.TicketTimeIn, &tout, &first,
			&t.CreatedAt, &t.UpdatedAt,
		)

		if errors.Is(err, sql.ErrNoRows) {
			apiFail(w, http.StatusNotFound, map[string]string{"ticket": "not_found"})
			return
		}
		if err != nil {
			apiFail(w, http.StatusInternalServerError, map[string]string{"general": "invalid"})
			return
		}

		if desc.Valid {
			v := desc.String
			t.DetailedDesc = &v
		}
		if handled.Valid {
			v := handled.Int64
			t.HandledByUserID = &v
		}
		if tout.Valid {
			v := tout.Time
			t.TicketTimeOut = &v
		}
		if first.Valid {
			v := first.Time
			t.FirstResponseAt = &v
		}

		apiOK(w, map[string]any{"ticket": t})
	}
}

type resolveTicketRequest struct {
	TicketID string `json:"ticket_id"`
}

// PATCH /ticket/resolve  (by ticket_id)
// Body: {"ticket_id":"TCK-0001"}
func ResolveTicketHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			apiFail(w, http.StatusMethodNotAllowed, map[string]string{"method": "invalid"})
			return
		}

		var req resolveTicketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apiFail(w, http.StatusBadRequest, map[string]string{"body": "invalid"})
			return
		}

		req.TicketID = strings.TrimSpace(req.TicketID)
		if req.TicketID == "" {
			apiFail(w, http.StatusBadRequest, map[string]string{"ticket_id": "required"})
			return
		}

		var updated int64
		err := db.QueryRow(`
			UPDATE tickets
			SET auto_resolved = TRUE
			WHERE ticket_id = $1
			RETURNING id
		`, req.TicketID).Scan(&updated)

		if errors.Is(err, sql.ErrNoRows) {
			apiFail(w, http.StatusNotFound, map[string]string{"ticket": "not_found"})
			return
		}
		if err != nil {
			apiFail(w, http.StatusInternalServerError, map[string]string{"general": "invalid"})
			return
		}

		apiOK(w, map[string]any{"resolved": true})
	}
}