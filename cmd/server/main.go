package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	http1 "telemed/internal/delivery/http"
	"telemed/internal/repository/postgres"
	"telemed/internal/usecase/auth"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env файл не найден. Читаем переменные из окружения")
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Fail to connect database", err)
	}
	defer db.Close()

	authRepo := postgres.NewAuthRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := http1.NewAuthHandler(authService)

	mux := http.NewServeMux()
	mux.HandleFunc("/login", authHandler.Login)

	handler := http1.AuthMiddleware(mux)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server starting at localhost " + port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
