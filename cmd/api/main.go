package main

import (
    "log"
    "os"
    "fmt"
    stdhttp "net/http" 
    "database/sql"
    
    _ "github.com/jackc/pgx/v5/stdlib"

    "github.com/gorilla/mux"
    handler "be_customer/internal/delivery/http"
    "be_customer/internal/repository"
    "be_customer/internal/usecase"

     "github.com/joho/godotenv"

)

func main() {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using system environment variables")
    }

    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    name := os.Getenv("DB_NAME")
    ssl := os.Getenv("DB_SSLMODE")
    serverPort := os.Getenv("SERVER_PORT")

    dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
        user, password, host, port, name, ssl)

    // 3️⃣ Connect to PostgreSQL
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    r := mux.NewRouter()

    // DI (dependency injection)
    repo := repository.NewMerchantRepoPG(db)
    usecase := usecase.NewMerchantUsecase(repo)

    handler.NewMerchantHandler(r, usecase)

    log.Println("Server running at port", serverPort)
    stdhttp.ListenAndServe(":"+serverPort, r)
}