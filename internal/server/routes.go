package server

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/uptrace/bun"
)

var RegisterRoutes = func(hub *Hub, r chi.Router, database *bun.DB, duckdb *sql.DB, msg chan Message) {
	db = database
	ddb = duckdb
	h = hub

	r.Group(func(r chi.Router) {
		r.Use(RateLimiter)
		r.Get("/", IndexHandler)
		r.Get("/register", RegisterHandlerPage)
		r.Get("/login", LoginHandlerPage)
		r.Post("/register", RegisterHandler)
		r.Post("/login", LoginHandler)
	})

	r.Post("/logout", LogoutHandler)
	r.Post("/leave", LeaveRoomHandler)
	r.Post("/create", CreateRoomHandler)
	r.Get("/room/{id}/msg", GetMessageLogs)

	r.Group(func(r chi.Router) {
		r.Use(CheckCookieAuth)
		r.Get("/dashboard", DashboardHandlerPage)
		r.Post("/dashboard", JoinHandler)
		r.Get("/room/{id}", RoomHandlerPage)
		r.Get("/ws/{id}", func(w http.ResponseWriter, req *http.Request) {
			WebSocketHandler(w, req, msg)
		})
	})
}
