package handlers

import (
	"net/http"
	"time"

	"myapp/internal/http/dto"
	"myapp/internal/http/handlers/responses"

	"github.com/golang-jwt/jwt/v5"
)

func makeJWT(secret string, u dto.UserDTO, rememberMe bool) (string, error) {
	now := time.Now()

	// ✅ Your requirement:
	// if remember_me = true => 30 minutes
	// else => keep normal 24 hours (you can change if you want)
	exp := now.Add(24 * time.Hour)
	if rememberMe {
		exp = now.Add(30 * time.Minute)
	}

	claims := jwt.MapClaims{
		"sub":      u.ID,
		"username": u.Username,
		"role":     u.RoleKey,
		"iat":      now.Unix(),
		"exp":      exp.Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func sendFail(w http.ResponseWriter, code int, errs map[string]string) {
	writeJSON(w, code, responses.LoginResponse{
		Status: false,
		Errors: errs,
	})
}

func sendSuccess(w http.ResponseWriter, token string, user *dto.UserDTO) {
	writeJSON(w, http.StatusOK, responses.LoginResponse{
		Status: true,
		Token:  token,
		User:   user,
		Errors: map[string]string{},
	})
}