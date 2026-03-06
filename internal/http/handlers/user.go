package handlers

import (
	"database/sql"
	"net/http"

	"myapp/internal/http/dto"
)

func ListUsersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
				"status": false,
				"errors": map[string]string{"method": "method not allowed"},
			})
			return
		}

		rows, err := db.Query(`
			SELECT id, username, first_name, last_name, role_key, gender, email, phone_number,
			       last_login, is_online, created_at
			FROM users
			ORDER BY id DESC
		`)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"status": false,
				"errors": map[string]string{"general": "db error: " + err.Error()},
			})
			return
		}
		defer rows.Close()

		out := make([]dto.UserDTO, 0)
		for rows.Next() {
			var u dto.UserDTO
			var firstName, lastName, gender, phone sql.NullString
			var lastLogin sql.NullTime

			if err := rows.Scan(
				&u.ID, &u.Username, &firstName, &lastName, &u.RoleKey, &gender, &u.Email, &phone,
				&lastLogin, &u.IsOnline, &u.CreatedAt,
			); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{
					"status": false,
					"errors": map[string]string{"general": "scan error: " + err.Error()},
				})
				return
			}

			if firstName.Valid {
				u.FirstName = &firstName.String
			}
			if lastName.Valid {
				u.LastName = &lastName.String
			}
			if gender.Valid {
				u.Gender = &gender.String
			}
			if phone.Valid {
				u.PhoneNumber = &phone.String
			}
			if lastLogin.Valid {
				t := lastLogin.Time
				u.LastLogin = &t
			}

			out = append(out, u)
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"status": true,
			"users":  out,
			"errors": map[string]string{},
		})
	}
}