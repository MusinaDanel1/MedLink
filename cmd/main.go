package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	http1 "telemed/internal/delivery/http"
	"telemed/internal/delivery/telegram"
	"telemed/internal/delivery/video"
	"telemed/internal/repository/postgres"
	"telemed/internal/usecase"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("⚠️  .env файл не найден. Читаем переменные из окружения", err)
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не найден в .env")
	}

	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Fatal("OPENAI_API_KEY не установлен в .env")
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

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Ошибка инициализации Telegram-бота: ", err)
	}

	bot.Debug = true

	log.Printf("Запущен бот: %s", bot.Self.UserName)

	// Initialize repositories
	authRepo := postgres.NewAuthRepository(db)
	adminRepo := postgres.NewAdminRepository(db)
	patientRepo := postgres.NewPatientRepository(db)
	doctorRepo := postgres.NewDoctorRepository(db)
	appointmentRepo := postgres.NewAppointmentRepository(db)
	msgRepo := postgres.NewMessageRepository(db)

	// Initialize services
	authService := usecase.NewAuthService(authRepo)
	doctorService := usecase.NewDoctorService(doctorRepo)
	adminService := usecase.NewAdminService(adminRepo, doctorService)
	patientService := usecase.NewPatientService(patientRepo)
	appointmentService := usecase.NewAppointmentService(appointmentRepo)
	msgService := usecase.NewMessageService(msgRepo)
	openaiService := usecase.New(openaiKey)
	// Initialize handlers
	authHandler := http1.NewAuthHandler(authService)
	adminHandler := http1.NewAdminHandler(adminService, doctorService)
	botHandler := telegram.NewBotHandler(bot, patientService, doctorService, appointmentService, openaiService)

	msgHandler := http1.NewMessageHandler(msgService)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	mux := http.NewServeMux()

	// Run bot updates handling in a separate goroutine
	go func() {
		for update := range updates {
			botHandler.HandleUpdate(update)
		}
	}()

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
	mux.Handle("/admin/specializations", adminMiddleware(http.HandlerFunc(adminHandler.GetSpecializations)))

	// Doctor routes
	mux.Handle("/doctor-dashboard", http1.AuthMiddleware(http.HandlerFunc(authHandler.DoctorDashboard)))

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("templates"))))

	r := gin.Default()
	// Set trusted proxy to localhost and private IPs
	r.SetTrustedProxies([]string{"127.0.0.1", "192.168.0.0/16", "10.0.0.0/8"})

	// 1) Сигнальный WebSocket
	r.GET("/ws", video.SignalingHandler)
	// 2) Отдаём конкретную страницу приёма
	r.GET("/appointment.html", func(c *gin.Context) {
		c.File("./templates/appointment.html")
	})
	// 3) Раздаём остальную статику из под /static
	r.Static("/static", "./static")
	r.GET("/api/appointments/:id/messages", msgHandler.List)
	r.POST("/api/appointments/:id/messages", msgHandler.Create)

	// Start Gin server in a goroutine on port 8080
	go func() {
		if err := r.Run(":8080"); err != nil {
			log.Fatal(err)
		}
	}()

	// Start HTTP server on port 8081
	port := "8081" // Changed from 8080 to 8081
	log.Printf("HTTP Server starting at http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
