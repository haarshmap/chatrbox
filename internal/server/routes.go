package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/uptrace/bun"
)

var RegisterRoutes = func(r chi.Router, db *bun.DB) {

	r.Post("/register", registerHandler)
	r.Post("/login", LoginHandler)
	r.Post("/logout", LogoutHandler)

	r.Group(func(r chi.Router) {
		r.Use(CheckCookieAuth)
		r.Post("/dashboard", DashboardHandler)
	})
}
