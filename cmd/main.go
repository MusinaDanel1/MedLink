package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

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

	password := "mypassword"
	password1 := "mypassword1"

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Error hashing password:", err)
	}

	hashedPassword1, err := bcrypt.GenerateFromPassword([]byte(password1), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Error hashing password:", err)
	}

	// Печатаем хеш для проверки
	fmt.Println("Hashed password:", string(hashedPassword))
	fmt.Println("Hashed password:", string(hashedPassword1))

	// Вставляем пользователя с хешированным паролем в таблицу users
	query := `INSERT INTO users (iin, password_hash, full_name, role) 
	          VALUES ($1, $2, $3, $4)`

	_, err = db.Exec(query, "123456789012", string(hashedPassword), "Иван Иванов", "admin")
	if err != nil {
		log.Fatal("Error inserting user into database:", err)
	}

	fmt.Println("User successfully inserted into the database!")

	query1 := `INSERT INTO users (iin, password_hash, full_name, role) 
	          VALUES ($1, $2, $3, $4)`

	_, err = db.Exec(query1, "040831650398", string(hashedPassword1), "Марина Цветаева", "doctor")
	if err != nil {
		log.Fatal("Error inserting user into database:", err)
	}

	fmt.Println("User successfully inserted into the database!")

	authRepo := postgres.NewAuthRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := http1.NewAuthHandler(authService)

	mux := http.NewServeMux()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			log.Println("THIS IS SHOW LOGIN FORM")
			authHandler.ShowLoginForm(w, r)
			return
		}
		if r.Method == http.MethodPost {
			log.Println("THIS IS SHOW LOGIN")
			authHandler.Login(w, r)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})

	protectedHandler := http1.AuthMiddleware(http.HandlerFunc(authHandler.ProtectedRoute))
	mux.Handle("/protected", protectedHandler)
	mux.Handle("/admin-dashboard", http1.AuthMiddleware(http.HandlerFunc(authHandler.AdminDashboard)))
	mux.Handle("/doctor-dashboard", http1.AuthMiddleware(http.HandlerFunc(authHandler.DoctorDashboard)))

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server starting at localhost " + port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
