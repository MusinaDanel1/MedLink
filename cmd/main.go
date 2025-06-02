package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

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
	_ = godotenv.Load()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN must be set in environment")
	}

	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Fatal("OPENAI_API_KEY must be set in environment")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL must be set in environment")
	}

	// 2) Подключаемся к БД
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// --- BEGIN New User Seeding Logic ---
	log.Println("Starting new user seeding...")

	// Seed math/rand for password generation (crypto/rand is better for production)
	// Using math/rand for simplicity as per typical mock data generation context
	// For crypto/rand, no seeding is needed.
	// For this exercise, let's assume math/rand is chosen for simplicity.
	// rand.Seed(time.Now().UnixNano()) // Usually done once, but if main is re-run, ensure it's effective.

	// Helper function to generate random alphanumeric password
	randAlphanum := func(length int) string {
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		seededRand := rand.New(rand.NewSource(time.Now().UnixNano())) // Ensure new seed each call for variety in quick succession
		b := make([]byte, length)
		for i := range b {
			b[i] = charset[seededRand.Intn(len(charset))]
		}
		return string(b)
	}

	// Doctor User Seeding
	// Note: The original list had 17 names. The request was for 16 *new* doctor users.
	// "Марина Цветаева" and "Айгуль Омарова", "Жанат Мусаев" are already in the doctors table from init.sql.
	// This seeding focuses on creating *user accounts* for doctors.
	// The provided list includes names already in the doctors table.
	// The logic below will attempt to create user accounts for all these names,
	// using ON CONFLICT for IINs.
	allDoctorNamesForUserSeeding := []string{
		"Марина Цветаева",
		"Айгуль Омарова",
		"Жанат Мусаев",
		"Алексей Смирнов",
		"Гаухар Сагынова",
		"Игорь Брагин",
		"Наталья Ким",
		"Бахытжан Ермеков",
		"Ольга Соколова",
		"Мурат Бейсеков",
		"Елена Жумагалиева",
		"Руслан Тлеулин",
	}

	// Start IINs for new users from a different range to avoid conflicts with manually specified ones.
	var iinCounter int64 = 100000000000 // Starting IIN for newly generated ones

	userInsertQuery := `INSERT INTO users (iin, password_hash, full_name, role) 
	                    VALUES ($1, $2, $3, $4)
	                    ON CONFLICT (iin) DO NOTHING`

	log.Println("Seeding new doctor users...")
	for _, name := range allDoctorNamesForUserSeeding {
		// Handle specific known IINs first
		currentIIN := fmt.Sprintf("%012d", iinCounter)
		iinCounter++

		newPassword := randAlphanum(12) // Generate a random password
		newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password for doctor %s: %v", name, err)
			continue
		}

		result, err := db.Exec(userInsertQuery, currentIIN, string(newHashedPassword), name, "doctor")
		if err != nil {
			log.Printf("Error inserting/verifying doctor user %s (IIN: %s): %v", name, currentIIN, err)
		} else {
			affected, _ := result.RowsAffected()
			if affected > 0 {
				log.Printf("CREATED Doctor User - IIN: %s, Name: %s, Password: %s", currentIIN, name, newPassword)
			}
			log.Printf("User entry for doctor %s (IIN: %s) processed.", name, currentIIN)
		}
	}

	// New Admin User Seeding
	log.Println("Seeding new admin users...")
	adminFirstNames := []string{"Бауржан", "Айдын", "Санжар", "Ермек", "Нурлан", "Динара", "Алия", "Гаухар"}
	adminLastNames := []string{"Ибраев", "Смагулов", "Касенов", "Ахметова", "Жуманова", "Ким", "Ли", "Сергеев"}

	for i := 0; i < 3; i++ { // Loop 3 times for 3 admins
		// Ensure we have enough names, or handle gracefully if not
		if i >= len(adminFirstNames) || i >= len(adminLastNames) {
			log.Printf("Warning: Not enough unique names in predefined lists to create admin user %d. Skipping.", i+1)
			continue
		}
		adminFullName := fmt.Sprintf("%s %s", adminFirstNames[i], adminLastNames[i])

		newIIN := fmt.Sprintf("%012d", iinCounter) // Continue IIN sequence
		iinCounter++

		newPassword := randAlphanum(12)
		newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password for admin %s: %v", adminFullName, err)
			continue
		}

		result, err := db.Exec(userInsertQuery, newIIN, string(newHashedPassword), adminFullName, "admin")
		if err != nil {
			log.Printf("Error inserting admin user %s (IIN: %s): %v", adminFullName, newIIN, err)
		} else {
			affected, _ := result.RowsAffected()
			if affected > 0 {
				log.Printf("CREATED Admin User - IIN: %s, Name: %s, Password: %s", newIIN, adminFullName, newPassword)
			}
			log.Printf("Admin user %s (IIN: %s) processed.", adminFullName, newIIN)
		}
	}

	log.Println("Finished new user seeding.")
	// --- END New User Seeding Logic ---

	// Initialize repositories
	authRepo := postgres.NewAuthRepository(db)
	adminRepo := postgres.NewAdminRepository(db)
	patientRepo := postgres.NewPatientRepository(db)
	doctorRepo := postgres.NewDoctorRepository(db)
	appointmentRepo := postgres.NewAppointmentRepository(db)
	msgRepo := postgres.NewMessageRepository(db)
	scheduleRepo := postgres.NewScheduleRepo(db)
	timeslotRepo := postgres.NewTimeslotRepo(db)
	// schedRepo := postgres.NewScheduleRepo(db) // Duplicate of scheduleRepo
	videoRepo := postgres.NewVideoSessionRepository(db)

	// Initialize services
	videoSvc := usecase.NewVideoService(videoRepo)
	authService := usecase.NewAuthService(authRepo)
	doctorService := usecase.NewDoctorService(doctorRepo)
	adminService := usecase.NewAdminService(adminRepo, doctorService)
	patientService := usecase.NewPatientService(patientRepo)
	appointmentService := usecase.NewAppointmentService(appointmentRepo, scheduleRepo, timeslotRepo, videoSvc)
	msgService := usecase.NewMessageService(msgRepo)
	openaiService := usecase.New(openaiKey)
	schedSvc := usecase.NewScheduleService(scheduleRepo) // Changed schedRepo to scheduleRepo

	// Initialize handlers
	authHandler := http1.NewAuthHandler(authService)
	authHandler.SetDoctorService(doctorService)
	adminHandler := http1.NewAdminHandler(adminService, doctorService)
	doctorHandler := http1.NewDoctorHandler(doctorService)
	doctorHandler.SetPatientService(patientService)
	schedH := http1.NewScheduleHandler(schedSvc, timeslotRepo)

	// Initialize Telegram bot
	var bot *tgbotapi.BotAPI
	var botHandler *telegram.BotHandler
	var notificationService *usecase.NotificationService

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Ошибка инициализации Telegram-бота: ", err)
	}
	bot.Debug = true
	log.Printf("Запущен бот: %s", bot.Self.UserName)

	botHandler = telegram.NewBotHandler(bot, patientService, doctorService, appointmentService, nil, openaiService)
	notificationService = usecase.NewNotificationService(
		appointmentRepo,
		patientRepo,
		botHandler,
	)
	botHandler.SetNotificationService(notificationService)
	notificationService.StartNotificationScheduler()

	msgHandler := http1.NewMessageHandler(msgService, appointmentService)
	apptHandler := http1.NewAppointmentHandler(appointmentService, doctorService)
	apptHandler.SetBotHandler(botHandler)

	// Run bot updates handling in a separate goroutine
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	go func() {
		for update := range updates {
			botHandler.HandleUpdate(update)
		}
	}()

	r := gin.Default()
	// Set trusted proxy to localhost and private IPs
	r.SetTrustedProxies([]string{"127.0.0.1", "192.168.0.0/16", "10.0.0.0/8"})

	// 2) WebSocket-сигналинг для вашего video.SignalingHandler
	r.GET("/ws", video.SignalingHandler)

	// 3) Статика
	//    Ваши CSS/JS лежат в templates/static/{css,js}
	r.Static("/static", "./templates/static")

	// 4) Страница WebRTC-рума
	//    URL: /webrtc/room.html?appointment_id=42&role=doctor
	r.StaticFile("/webrtc/room.html", "./templates/appointment.html")

	// 5) Отладочная страница, если нужна
	r.StaticFile("/debug.html", "./templates/debug.html")

	// Add login routes to Gin server
	r.GET("/login", func(c *gin.Context) {
		authHandler.ShowLoginForm(c.Writer, c.Request)
	})
	r.GET("/main", func(c *gin.Context) {
		authHandler.ShowMainForm(c.Writer, c.Request)
	})
	r.POST("/login", func(c *gin.Context) {
		authHandler.Login(c.Writer, c.Request)
	})

	// Add protected and admin routes to Gin
	r.GET("/protected", func(c *gin.Context) {
		http1.AuthMiddleware(http.HandlerFunc(authHandler.ProtectedRoute)).ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/admin-dashboard", func(c *gin.Context) {
		http1.AuthMiddleware(http.HandlerFunc(authHandler.AdminDashboard)).ServeHTTP(c.Writer, c.Request)
	})
	r.Any("/admin/register", func(c *gin.Context) {
		http1.AuthMiddleware(http.HandlerFunc(adminHandler.RegisterUser)).ServeHTTP(c.Writer, c.Request)
	})
	r.Any("/admin/block", func(c *gin.Context) {
		http1.AuthMiddleware(http.HandlerFunc(adminHandler.BlockUser)).ServeHTTP(c.Writer, c.Request)
	})
	r.Any("/admin/unblock", func(c *gin.Context) {
		http1.AuthMiddleware(http.HandlerFunc(adminHandler.UnblockUser)).ServeHTTP(c.Writer, c.Request)
	})
	r.Any("/admin/delete", func(c *gin.Context) {
		http1.AuthMiddleware(http.HandlerFunc(adminHandler.DeleteUser)).ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/admin/users", func(c *gin.Context) {
		http1.AuthMiddleware(http.HandlerFunc(adminHandler.GetAllUsers)).ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/admin/specializations", func(c *gin.Context) {
		http1.AuthMiddleware(http.HandlerFunc(adminHandler.GetSpecializations)).ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/doctor-dashboard", func(c *gin.Context) {
		http1.AuthMiddleware(http.HandlerFunc(authHandler.DoctorDashboard)).ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/api/appointments/:id/messages", msgHandler.List)
	r.POST("/api/appointments/:id/messages", msgHandler.Create)
	r.GET("/api/appointments/:id/details", apptHandler.GetAppointmentDetails)
	r.PUT("/api/appointments/:id/details", apptHandler.CompleteAppointment)
	r.POST("/api/appointments/:id/complete", apptHandler.CompleteAppointment)
	r.PUT("/api/appointments/:id/accept", apptHandler.AcceptAppointment)
	r.GET("/api/appointments/:id/status", apptHandler.GetAppointmentStatus) // Добавлен /api/appointments + правильный хендлер
	r.PUT("/api/appointments/:id/end-call", apptHandler.EndCall)
	r.POST("/api/appointments", apptHandler.BookAppointment)
	r.GET("/api/appointments", apptHandler.ListBySchedules)
	r.GET("/api/diagnoses", doctorHandler.GetAllDiagnoses)
	r.GET("/api/services", doctorHandler.GetAllServices)
	r.POST("/api/services", doctorHandler.CreateService)
	r.GET("/api/patients", doctorHandler.GetAllPatients)
	r.GET("/api/doctors/:id", doctorHandler.GetDoctorByID)
	r.GET("/api/doctors", doctorHandler.GetAllDoctors)
	r.GET("/api/schedules", schedH.GetSchedules)
	r.POST("/api/schedules", schedH.CreateSchedule)

	// Start Gin server on port 8080
	log.Printf("Server starting at http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
