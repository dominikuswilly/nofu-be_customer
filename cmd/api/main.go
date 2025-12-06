package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	handler "be_customer/internal/delivery/http"
	"be_customer/internal/repository"
	"be_customer/internal/usecase"

	"be_customer/internal/middleware"

	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	// Load .env — optional
	_ = godotenv.Load()

	// Load required environment variables
	requiredVars := []string{
		"DB_USER", "DB_PASSWORD", "DB_HOST", "DB_PORT",
		"DB_NAME", "DB_SSLMODE", "SERVER_PORT",
	}

	log.SetOutput(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    20, // MB
		MaxBackups: 7,
		MaxAge:     1, // days (rotate daily)
		Compress:   true,
	})

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
	r.Use(middleware.Logging)

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.JSONResponse)

	// Dependency injection
	merchantRepo := repository.NewMerchantRepoPG(db)
	adminRepo := repository.NewAdminRepoPG(db)
	merchantUc := usecase.NewMerchantUsecase(merchantRepo)
	adminUc := usecase.NewAdminUsecase(adminRepo)

	// HTTP handlers
	handler.NewMerchantHandler(r, merchantUc)
	handler.NewAdminHandler(r, adminUc)

	port := os.Getenv("SERVER_PORT")
	log.Println("🚀 Server running on port", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

}
