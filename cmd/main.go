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
	"telemed/internal/usecase"
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Ошибка при хешировании пароля:", err)
	}
	hashedPassword1, err := bcrypt.GenerateFromPassword([]byte(password1), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Ошибка при хешировании пароля:", err)
	}

	query := `INSERT INTO users (iin, password_hash, full_name, role) 
	          VALUES ($1, $2, $3, $4)
	          ON CONFLICT (iin) DO NOTHING`

	_, err = db.Exec(query, "123456789012", string(hashedPassword), "Иван Иванов", "admin")
	if err != nil {
		log.Fatal("Ошибка при вставке админа:", err)
	}
	_, err = db.Exec(query, "040831650398", string(hashedPassword1), "Марина Цветаева", "doctor")
	if err != nil {
		log.Fatal("Ошибка при вставке врача:", err)
	}

	fmt.Println("Админ и врач успешно добавлены в базу данных!")

	// Initialize repositories
	authRepo := postgres.NewAuthRepository(db)
	adminRepo := postgres.NewAdminRepository(db)

	// Initialize services
	authService := usecase.NewAuthService(authRepo)
	adminService := usecase.NewAdminService(adminRepo)

	// Initialize handlers
	authHandler := http1.NewAuthHandler(authService)
	adminHandler := http1.NewAdminHandler(adminService)

	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authHandler.ShowLoginForm(w, r)
			return
		}
		if r.Method == http.MethodPost {
			authHandler.Login(w, r)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})

	// Protected routes
	protectedHandler := http1.AuthMiddleware(http.HandlerFunc(authHandler.ProtectedRoute))
	mux.Handle("/protected", protectedHandler)

	// Admin routes
	adminMiddleware := http1.AuthMiddleware
	mux.Handle("/admin-dashboard", adminMiddleware(http.HandlerFunc(authHandler.AdminDashboard)))
	mux.Handle("/admin/register", adminMiddleware(http.HandlerFunc(adminHandler.RegisterUser)))
	mux.Handle("/admin/block", adminMiddleware(http.HandlerFunc(adminHandler.BlockUser)))
	mux.Handle("/admin/unblock", adminMiddleware(http.HandlerFunc(adminHandler.UnblockUser)))
	mux.Handle("/admin/delete", adminMiddleware(http.HandlerFunc(adminHandler.DeleteUser)))
	mux.Handle("/admin/users", adminMiddleware(http.HandlerFunc(adminHandler.GetAllUsers)))

	// Doctor routes
	mux.Handle("/doctor-dashboard", http1.AuthMiddleware(http.HandlerFunc(authHandler.DoctorDashboard)))

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting at http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
