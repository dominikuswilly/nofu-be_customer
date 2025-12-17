package main

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	handlerAdmin "be_customer/internal/delivery/http"
	handlerMerchant "be_customer/internal/delivery/http/merchant"
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

	// Setup rotating logger (daily) in current path and symlink to ./app.log
	rl, err := rotatelogs.New(
		"./app.log.%Y%m%d",
		rotatelogs.WithLinkName("./app.log"),      // create/keep ./app.log -> latest file
		rotatelogs.WithRotationTime(24*time.Hour), // rotate daily
		rotatelogs.WithMaxAge(7*24*time.Hour),     // keep 7 days
	)
	if err != nil {
		log.Fatalf("failed to setup rotatelogs: %v", err)
	}
	log.SetOutput(rl)
	// Optionally also send to stderr (useful for container logs)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// start compressor goroutine that gzips rotated files
	go compressRotatedFiles(".", "app.log.", 1*time.Minute)

	log.SetOutput(io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    20, // MB
		MaxBackups: 7,
		MaxAge:     1, // days (rotate daily)
		Compress:   true,
	}))

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
	r.Use(middleware.CORSMiddleware) // Add CORS middleware

	// Global OPTIONS handler to ensure CORS middleware intercepts preflight requests
	r.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.JSONResponse)

	// Dependency injection
	merchantRepo := repository.NewMerchantRepoPG(db)
	adminRepo := repository.NewAdminRepoPG(db)
	merchantUc := usecase.NewMerchantUsecase(merchantRepo)
	adminUc := usecase.NewAdminUsecase(adminRepo)

	// HTTP handlers
	handlerMerchant.NewMerchantHandler(r, merchantUc)
	handlerAdmin.NewAdminHandler(r, adminUc)

	port := os.Getenv("SERVER_PORT")
	log.Println("🚀 Server running on port", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

}

// compressRotatedFiles gzips any rotated files matching prefix that are older than minAge.
// It writes to a .tmp then renames to .gz for atomicity, then deletes original.
func compressRotatedFiles(dir, prefix string, minAge time.Duration) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		_ = compressOnce(dir, prefix, minAge)
		<-ticker.C
	}
}

func compressOnce(dir, prefix string, minAge time.Duration) error {
	now := time.Now()
	glob := filepath.Join(dir, prefix+"*")
	matches, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	for _, p := range matches {
		// skip already compressed files
		if strings.HasSuffix(p, ".gz") || strings.HasSuffix(p, ".tmp") {
			continue
		}
		// skip symlink (e.g. ./app.log) or directories
		if info, err := os.Lstat(p); err == nil {
			if info.Mode()&os.ModeSymlink != 0 || info.IsDir() {
				continue
			}
			// skip recent files
			if now.Sub(info.ModTime()) < minAge {
				continue
			}
		}

		tmpPath := p + ".gz.tmp"
		gzPath := p + ".gz"

		if err := compressFile(p, tmpPath); err != nil {
			log.Printf("compress error for %s: %v", p, err)
			_ = os.Remove(tmpPath)
			continue
		}
		if err := os.Rename(tmpPath, gzPath); err != nil {
			log.Printf("rename tmp gz failed: %v", err)
			_ = os.Remove(tmpPath)
			continue
		}
		if err := os.Remove(p); err != nil {
			log.Printf("remove original after compress: %v", err)
		} else {
			log.Printf("compressed %s -> %s", p, gzPath)
		}
	}
	return nil
}

func compressFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	gw.Name = filepath.Base(src)
	defer gw.Close()

	_, err = io.Copy(gw, in)
	return err
}
