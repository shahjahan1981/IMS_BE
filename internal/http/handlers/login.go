package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"myapp/internal/http/dto"
	"myapp/internal/http/handlers/requests"

	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendFail(w, http.StatusMethodNotAllowed, map[string]string{
				"method": "method not allowed",
			})
			return
		}

		var req requests.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendFail(w, http.StatusBadRequest, map[string]string{
				"general": "invalid json",
			})
			return
		}

		req.Username = strings.TrimSpace(req.Username)
		req.Password = strings.TrimSpace(req.Password)

		errs := map[string]string{}
		if req.Username == "" {
			errs["username"] = "required"
		}
		if req.Password == "" {
			errs["password"] = "required"
		}
		if len(errs) > 0 {
			sendFail(w, http.StatusBadRequest, errs)
			return
		}

		var (
			u            dto.UserDTO
			passwordHash string
			firstName    sql.NullString
			lastName     sql.NullString
			gender       sql.NullString
			phoneNumber  sql.NullString
			lastLogin    sql.NullTime
		)

		err := db.QueryRow(`
			SELECT id, username, first_name, last_name, role_key, gender, email, phone_number,
			       last_login, is_online, created_at, password_hash
			FROM users
			WHERE username = $1
		`, req.Username).Scan(
			&u.ID, &u.Username, &firstName, &lastName, &u.RoleKey, &gender, &u.Email, &phoneNumber,
			&lastLogin, &u.IsOnline, &u.CreatedAt, &passwordHash,
		)

		if err == sql.ErrNoRows {
			sendFail(w, http.StatusUnauthorized, map[string]string{
				"username_or_password": "invalid",
			})
			return
		}
		if err != nil {
			sendFail(w, http.StatusInternalServerError, map[string]string{
				"general": "db error: " + err.Error(),
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
		if phoneNumber.Valid {
			u.PhoneNumber = &phoneNumber.String
		}
		if lastLogin.Valid {
			t := lastLogin.Time
			u.LastLogin = &t
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			sendFail(w, http.StatusUnauthorized, map[string]string{
				"username_or_password": "invalid",
			})
			return
		}

		var updatedLastLogin time.Time
		err = db.QueryRow(`
			UPDATE users
			SET last_login = now(), is_online = TRUE
			WHERE id = $1
			RETURNING last_login, is_online
		`, u.ID).Scan(&updatedLastLogin, &u.IsOnline)

		if err != nil {
			sendFail(w, http.StatusInternalServerError, map[string]string{
				"general": "db error: " + err.Error(),
			})
			return
		}
		u.LastLogin = &updatedLastLogin

		secret := os.Getenv("JWT_SECRET")
		if strings.TrimSpace(secret) == "" {
			sendFail(w, http.StatusInternalServerError, map[string]string{
				"general": "server missing JWT_SECRET",
			})
			return
		}

		// ✅ here we pass remember_me
		tokenStr, err := makeJWT(secret, u, req.RememberMe)
		if err != nil {
			sendFail(w, http.StatusInternalServerError, map[string]string{
				"general": "token error",
			})
			return
		}

		sendSuccess(w, tokenStr, &u)
	}
}