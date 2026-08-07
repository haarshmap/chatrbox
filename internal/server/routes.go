package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/uptrace/bun"
)

var RegisterRoutes = func(r chi.Router, database *bun.DB) {
	db = database
	r.Get("/", IndexHandler)
	r.Get("/register", RegisterHandlerPage)
	r.Get("/login", LoginHandlerPage)
	r.Get("/dashboard", DashboardHandlerPage)

	r.Post("/register", RegisterHandler)
	r.Post("/login", LoginHandler)
	r.Post("/logout", LogoutHandler)

	r.Group(func(r chi.Router) {
		r.Use(CheckCookieAuth)
		r.Post("/dashboard", DashboardHandler)
	})
}
