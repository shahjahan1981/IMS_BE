package router

import (
	"database/sql"
	"net/http"

	"myapp/internal/http/handlers"
)

func SetupRouter(db *sql.DB, protected func(http.Handler) http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", handlers.LoginHandler(db))

	mux.Handle("/register", protected(handlers.RegisterHandler(db)))
	mux.Handle("/users", protected(handlers.ListUsersHandler(db)))
	mux.Handle("/user/create", protected(handlers.CreateUserHandler(db)))

	mux.Handle("/ticket/create", protected(handlers.CreateTicketHandler(db)))
	mux.Handle("/tickets/list", protected(handlers.ListTicketsHandler(db)))
	mux.Handle("/ticket/detail/", protected(handlers.TicketDetailByTicketIDHandler(db)))
	mux.Handle("/ticket/resolve", protected(handlers.ResolveTicketHandler(db)))

	return mux
}