package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"

    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/joho/godotenv"
    "github.com/gorilla/mux"

    handler "be_customer/internal/delivery/http"
    "be_customer/internal/repository"
    "be_customer/internal/usecase"
)

func main() {
    // Load .env — optional
    _ = godotenv.Load()

    // Load required environment variables
    requiredVars := []string{
        "DB_USER", "DB_PASSWORD", "DB_HOST", "DB_PORT",
        "DB_NAME", "DB_SSLMODE", "SERVER_PORT",
    }

    for _, key := range requiredVars {
        if os.Getenv(key) == "" {
            log.Fatalf("Missing required environment variable: %s", key)
        }
    }

    // Build DSN string
    dsn := fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=%s",
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_NAME"),
        os.Getenv("DB_SSLMODE"),
    )

    // Connect to PostgreSQL
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        log.Fatalf("Failed to connect database: %v", err)
    }
    defer db.Close()

    // Check actual DB connection
    if err := db.Ping(); err != nil {
        log.Fatalf("Database ping failed: %v", err)
    }

    // Router
    r := mux.NewRouter()

    // Dependency injection
    repo := repository.NewMerchantRepoPG(db)
    uc := usecase.NewMerchantUsecase(repo)

    // HTTP handlers
    handler.NewMerchantHandler(r, uc)

    port := os.Getenv("SERVER_PORT")
    log.Println("🚀 Server running on port", port)

    if err := http.ListenAndServe(":"+port, r); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
