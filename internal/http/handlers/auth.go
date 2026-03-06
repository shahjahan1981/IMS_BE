package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"myapp/internal/http/handlers/requests"
	"myapp/internal/http/handlers/responses"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			fail(w, http.StatusMethodNotAllowed, map[string]string{"method": "invalid"})
			return
		}

		var req requests.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			fail(w, http.StatusBadRequest, map[string]string{"body": "invalid"})
			return
		}

		// normalize
		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		req.Username = strings.TrimSpace(req.Username)
		req.FirstName = strings.TrimSpace(req.FirstName)
		req.LastName = strings.TrimSpace(req.LastName)
		req.RoleKey = strings.TrimSpace(strings.ToLower(req.RoleKey))
		req.Gender = strings.TrimSpace(strings.ToLower(req.Gender))
		req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)

		// Back-compat mapping
		if req.FirstName == "" && req.Name != "" {
			req.FirstName = strings.TrimSpace(req.Name)
		}
		if req.Username == "" && req.Email != "" {
			req.Username = strings.Split(req.Email, "@")[0]
		}

		// required (NO messages)
		errs := map[string]string{}
		if req.Username == "" {
			errs["username"] = "required"
		}
		if req.Email == "" {
			errs["email"] = "required"
		}
		if strings.TrimSpace(req.Password) == "" {
			errs["password"] = "required"
		}
		if len(errs) > 0 {
			fail(w, http.StatusBadRequest, errs)
			return
		}

		// role default + validation
		if req.RoleKey == "" {
			req.RoleKey = "engineer"
		}
		if !isValidRole(req.RoleKey) {
			fail(w, http.StatusBadRequest, map[string]string{"rolekey": "invalid"})
			return
		}

		// optional gender validation
		if req.Gender != "" && req.Gender != "male" && req.Gender != "female" && req.Gender != "other" {
			fail(w, http.StatusBadRequest, map[string]string{"gender": "invalid"})
			return
		}

		// hash password
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			fail(w, http.StatusInternalServerError, map[string]string{"general": "invalid"})
			return
		}

		var resp responses.RegisterResponse
		err = db.QueryRow(`
			INSERT INTO users
				(username, first_name, last_name, role_key, gender, email, phone_number, password_hash)
			VALUES
				($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id, username, first_name, last_name, role_key, gender, email, phone_number, is_online
		`,
			req.Username,
			nullIfEmpty(req.FirstName),
			nullIfEmpty(req.LastName),
			req.RoleKey,
			nullIfEmpty(req.Gender),
			req.Email,
			nullIfEmpty(req.PhoneNumber),
			string(hashBytes),
		).Scan(
			&resp.ID,
			&resp.Username,
			&resp.FirstName,
			&resp.LastName,
			&resp.RoleKey,
			&resp.Gender,
			&resp.Email,
			&resp.PhoneNumber,
			&resp.IsOnline,
		)

		if err != nil {
			// unique conflict -> return field properties only
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				c := strings.ToLower(pqErr.Constraint + " " + pqErr.Detail)

				switch {
				case strings.Contains(c, "username"):
					fail(w, http.StatusConflict, map[string]string{"username": "exists"})
				case strings.Contains(c, "email"):
					fail(w, http.StatusConflict, map[string]string{"email": "exists"})
				case strings.Contains(c, "phone"):
					fail(w, http.StatusConflict, map[string]string{"phonenumber": "exists"})
				default:
					fail(w, http.StatusConflict, map[string]string{"general": "exists"})
				}
				return
			}

			fail(w, http.StatusInternalServerError, map[string]string{"general": "invalid"})
			return
		}

		// success can stay as your normal response struct
		writeJSON(w, http.StatusOK, resp)
	}
}