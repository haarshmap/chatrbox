package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type Users struct {
	ID       int64  `bun:"id,pk,autoincrement"`
	Username string `bun:"username,unique,notnull"`
	Password string `bun:"password,notnull"`
}

func main() {
	var err error
	ctx := context.Background()
	r := chi.NewRouter()

	sqlite, err := sql.Open(sqliteshim.ShimName, "data.db")
	if err != nil {
		log.Fatalf("failed to initialise database %v", err)
	}

	db := bun.NewDB(sqlite, sqlitedialect.New())

	_, err = db.NewCreateTable().Model((*Users)(nil)).IfNotExists().Exec(ctx)

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	} else {
		fmt.Println("SQLite database initialized successfully with Bun!")
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	//routes
	r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("currently at register"))
		name := r.FormValue("username")
		password := r.FormValue("password")
		user := &Users{Username: name, Password: password}
		_, err = db.NewInsert().Model(user).Exec(ctx)
		if err != nil {
			slog.Log(r.Context(), 0, "failed to initialise row", err)
		}
	})
	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("currently at login"))
		name := r.FormValue("username")
		password := r.FormValue("password")
		user := new(Users)
		slog.Log(r.Context(), 0, "just checking out user", user)
		err = db.NewSelect().Model(user).Where("Username = ?", name).Scan(ctx)

		if password == user.Password {
			fmt.Println("User is logged in")
		} else {
			fmt.Println("User aint logged in twin")
		}

	})

	http.ListenAndServe(":"+os.Getenv("PORT"), r)

}
