package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func main() {
	var err error
	r := chi.NewRouter()

	sqlite, err := sql.Open(sqliteshim.ShimName, "data.db")
	if err != nil {
		log.Fatalf("failed to initialise database %v", err)
	}

	db := bun.NewDB(sqlite, sqlitedialect.New())

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	} else {
		fmt.Println("SQLite database initialized successfully with Bun!")
	}

	http.ListenAndServe(":"+os.Getenv("PORT"), r)
}
