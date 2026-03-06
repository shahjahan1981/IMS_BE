package router

import (
	"database/sql"
	"fmt"
	"net/http"

	"myapp/internal/http/handlers"
	"myapp/internal/http/middleware"
)

func New(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	// Public
	mux.HandleFunc("/login", handlers.LoginHandler(db))

	// Everything else requires token
	protected := func(h http.Handler) http.Handler { return middleware.Auth(h) }

	mux.Handle("/", protected(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Server is running 🚀")
	})))

	mux.Handle("/db-check", protected(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var now string
		if err := db.QueryRow("SELECT NOW()").Scan(&now); err != nil {
			http.Error(w, "DB query failed", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "DB OK ✅ Time:", now)
	})))

	// Users
	mux.Handle("/register", protected(handlers.RegisterHandler(db)))
	mux.Handle("/users", protected(handlers.ListUsersHandler(db)))
	mux.Handle("/user/create", protected(handlers.CreateUserHandler(db)))

	// Tickets
	mux.Handle("/ticket/create", protected(handlers.CreateTicketHandler(db)))
	mux.Handle("/tickets/list", protected(handlers.ListTicketsHandler(db)))
	mux.Handle("/ticket/detail/", protected(handlers.TicketDetailByTicketIDHandler(db)))
	mux.Handle("/ticket/resolve", protected(handlers.ResolveTicketHandler(db)))

	return mux
}