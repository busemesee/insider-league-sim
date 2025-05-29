package main

import (
    "database/sql"
    "log"
    "net/http"
    "os"

    _ "github.com/lib/pq"
    "github.com/gorilla/mux"
    "github.com/yourusername/insider-league-simulation/pkg/server"
    "github.com/yourusername/insider-league-simulation/pkg/simulation"
)

func main() {
    dbURL := os.Getenv("DB_CONN")
    if dbURL == "" {
        log.Fatal("DB_CONN env var is required")
    }
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    app := &server.AppContext{
        DB:        db,
        Simulator: simulation.NewSimpleSimulator(),
    }

    r := mux.NewRouter()

    // Register API routes first
    app.RegisterRoutes(r)

    // Serve static files from frontend directory (index.html at root)
    fs := http.FileServer(http.Dir("./frontend"))
    r.PathPrefix("/").Handler(fs)

    log.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
