package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/uptrace/bun"
)

var RegisterRoutes = func(hub *Hub, r chi.Router, database *bun.DB) {
	db = database
	h = hub

	r.Get("/", IndexHandler)
	r.Get("/register", RegisterHandlerPage)
	r.Get("/login", LoginHandlerPage)

	r.Post("/register", RegisterHandler)
	r.Post("/login", LoginHandler)
	r.Post("/logout", LogoutHandler)

	r.Post("/create", CreateRoomHandler)

	r.Group(func(r chi.Router) {
		r.Use(CheckCookieAuth)
		r.Get("/dashboard", DashboardHandlerPage)
		r.Get("/room/{id}", RoomHandlerPage)
		r.Get("/ws", WebSocketHandler)
	})
}
