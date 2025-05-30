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
func withCORS(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        h.ServeHTTP(w, r)
    })
}
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
    handler := withCORS(r)
log.Println("Server listening on :8080 with CORS enabled")
log.Fatal(http.ListenAndServe(":8080", handler))
}
