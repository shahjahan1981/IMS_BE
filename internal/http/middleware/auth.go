package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
	CtxUserID   ctxKey = "user_id"
	CtxUsername ctxKey = "username"
	CtxRole     ctxKey = "role"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func fail(w http.ResponseWriter, status int, errs map[string]string) {
	writeJSON(w, status, map[string]any{
		"status": false,
		"errors": errs,
	})
}

// Auth requires Authorization: Bearer <token>
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allow preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
		if secret == "" {
			fail(w, http.StatusInternalServerError, map[string]string{"jwt_secret": "missing"})
			return
		}

		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		if auth == "" {
			fail(w, http.StatusUnauthorized, map[string]string{"token": "required"})
			return
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			fail(w, http.StatusUnauthorized, map[string]string{"token": "invalid"})
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			fail(w, http.StatusUnauthorized, map[string]string{"token": "required"})
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			// only allow HMAC
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || token == nil || !token.Valid {
			fail(w, http.StatusUnauthorized, map[string]string{"token": "invalid"})
			return
		}

		// Extract claims (your login token has: sub, username, role)
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fail(w, http.StatusUnauthorized, map[string]string{"token": "invalid"})
			return
		}

		// sub is user id (often float64 in MapClaims)
		var userID int64
		if v, ok := claims["sub"]; ok {
			switch t := v.(type) {
			case float64:
				userID = int64(t)
			case int64:
				userID = t
			case string:
				// optional: if you ever send as string
				// ignore if parse fails
			}
		}

		username, _ := claims["username"].(string)
		role, _ := claims["role"].(string)

		ctx := r.Context()
		ctx = context.WithValue(ctx, CtxUserID, userID)
		ctx = context.WithValue(ctx, CtxUsername, username)
		ctx = context.WithValue(ctx, CtxRole, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}