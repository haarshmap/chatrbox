package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/haarshmap/chatrbox/internal/server"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func main() {
	var err error
	ctx := context.Background()
	r := chi.NewRouter()

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	sqlite, err := sql.Open(sqliteshim.ShimName, "/var/data/data.db")
	if err != nil {
		log.Fatalf("failed to initialise database %v", err)
	}

	conn, err := server.InitClickhouse()
	if err != nil {
		log.Fatalf("Failed to initialise Clickhouse", err)
	}
	if err == nil {
		fmt.Printf("\ninitialised clickhouse")
	}

	err = conn.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS message_logs (Col1 UInt8, Col2 String, Col3 String)
		`)
	if err != nil {
		log.Fatalf("Failed to create clickhouse db", err)
	}

	db := bun.NewDB(sqlite, sqlitedialect.New())

	_, err = db.NewCreateTable().Model((*server.Users)(nil)).IfNotExists().Exec(ctx)
	_, err = db.NewCreateTable().Model((*server.Rooms)(nil)).IfNotExists().Exec(ctx)
	_, err = db.NewCreateTable().Model((*server.RoomMembers)(nil)).IfNotExists().Exec(ctx)

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	} else {
		fmt.Println("SQLite database initialized successfully with Bun!")
	}

	if err := server.InitTemplates(); err != nil {
		log.Fatal(err)
	}
	Hub := server.NewHub()
	go Hub.Run()

	server.RegisterRoutes(Hub, r, db)

	port := os.Getenv("PORT")
	_, err = server.InitRedis(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting HTTP server on :%s", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
