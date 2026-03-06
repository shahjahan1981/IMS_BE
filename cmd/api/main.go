package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"myapp/internal/database"
	"myapp/internal/http/router"
)

func corsMiddleware(next http.Handler) http.Handler {
	allowed := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	allowCreds := strings.TrimSpace(strings.ToLower(os.Getenv("CORS_ALLOW_CREDENTIALS"))) == "true"

	allowedMap := map[string]bool{}
	if allowed != "" {
		for _, o := range strings.Split(allowed, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				allowedMap[o] = true
			}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// only set allow-origin if origin is in allowed list
		if origin != "" && allowedMap[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin") // important for caching
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if allowCreds {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	_ = godotenv.Load()

	db, err := database.OpenDB()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close()

	mux := router.New(db)

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(mux)))
}