package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"myapp/internal/http/handlers/requests"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendFail(w, http.StatusMethodNotAllowed, map[string]string{"method": "invalid"})
			return
		}

		var req requests.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendFail(w, http.StatusBadRequest, map[string]string{"body": "invalid"})
			return
		}

		// normalize + support both key styles
		username := strings.TrimSpace(req.Username)
		email := strings.TrimSpace(strings.ToLower(req.Email))
		password := strings.TrimSpace(req.Password)

		firstName := strings.TrimSpace(req.FirstName)
		if firstName == "" {
			firstName = strings.TrimSpace(req.Firstname)
		}

		lastName := strings.TrimSpace(req.LastName)
		if lastName == "" {
			lastName = strings.TrimSpace(req.Lastname)
		}

		roleKey := strings.TrimSpace(strings.ToLower(req.RoleKey))
		if roleKey == "" {
			roleKey = strings.TrimSpace(strings.ToLower(req.Rolekey))
		}

		gender := strings.TrimSpace(strings.ToLower(req.Gender))

		phone := strings.TrimSpace(req.PhoneNumber)
		if phone == "" {
			phone = strings.TrimSpace(req.Phonenumber)
		}

		// ✅ ALL REQUIRED
		errs := map[string]string{}
		if username == "" {
			errs["username"] = "required"
		}
		if password == "" {
			errs["password"] = "required"
		}
		if firstName == "" {
			errs["firstname"] = "required"
		}
		if lastName == "" {
			errs["lastname"] = "required"
		}
		if gender == "" {
			errs["gender"] = "required"
		}
		if roleKey == "" {
			errs["rolekey"] = "required"
		}
		if email == "" {
			errs["email"] = "required"
		}
		if phone == "" {
			errs["phonenumber"] = "required"
		}
		if len(errs) > 0 {
			sendFail(w, http.StatusBadRequest, errs)
			return
		}

		// validate values
		if !isValidRole(roleKey) {
			sendFail(w, http.StatusBadRequest, map[string]string{"rolekey": "invalid"})
			return
		}
		if gender != "male" && gender != "female" && gender != "other" {
			sendFail(w, http.StatusBadRequest, map[string]string{"gender": "invalid"})
			return
		}

		// hash password
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			sendFail(w, http.StatusInternalServerError, map[string]string{"general": "invalid"})
			return
		}
		passwordHash := string(hashBytes)

		// insert (return only id)
		var userID int64
		err = db.QueryRow(`
			INSERT INTO users (username, first_name, last_name, role_key, gender, email, phone_number, password_hash)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id
		`, username, firstName, lastName, roleKey, gender, email, phone, passwordHash).Scan(&userID)

		if err != nil {
			// unique conflict -> property only
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				c := strings.ToLower(pqErr.Constraint + " " + pqErr.Detail)
				switch {
				case strings.Contains(c, "username"):
					sendFail(w, http.StatusConflict, map[string]string{"username": "exists"})
				case strings.Contains(c, "email"):
					sendFail(w, http.StatusConflict, map[string]string{"email": "exists"})
				case strings.Contains(c, "phone"):
					sendFail(w, http.StatusConflict, map[string]string{"phonenumber": "exists"})
				default:
					sendFail(w, http.StatusConflict, map[string]string{"general": "exists"})
				}
				return
			}

			sendFail(w, http.StatusInternalServerError, map[string]string{"general": "invalid"})
			return
		}

		// ✅ SUCCESS: only created flag + user id + errors {}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":               true,
			"successfully_created": true,
			"user_id":              userID,
			"errors":               map[string]string{},
		})
	}
}