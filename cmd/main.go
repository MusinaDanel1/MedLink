package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	http1 "medlink/internal/delivery/http"
	"medlink/internal/delivery/telegram"
	"medlink/internal/delivery/video"
	"medlink/internal/repository/postgres"
	"medlink/internal/usecase"
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

	// New Doctor User Seeding
	log.Println("Seeding doctor users with specific IINs and passwords...")

	type doctorSeed struct {
		FullName string
		IIN      string
		Password string
	}

	doctors := []doctorSeed{
		{"Марина Цветаева", "870102300001", "cvetaeva123"},
		{"Айгуль Омарова", "860315400002", "omarova123"},
		{"Жанат Мусаев", "850610500003", "musaev123"},
		{"Алексей Смирнов", "840920600004", "smirnov123"},
		{"Гаухар Сагынова", "830207700005", "sagynova123"},
		{"Игорь Брагин", "820813800006", "bragin123"},
		{"Наталья Ким", "811225900007", "kim123"},
		{"Бахытжан Ермеков", "800501000008", "ermekov123"},
		{"Ольга Соколова", "790310100009", "sokolova123"},
		{"Мурат Бейсеков", "780624200010", "beisekov123"},
		{"Елена Жумагалиева", "770809300011", "zhumagalieva123"},
		{"Руслан Тлеулин", "760430400012", "tleulin123"},
	}

	insertQuery := `
	INSERT INTO users (iin, password_hash, full_name, role)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (iin) DO NOTHING
`

	for _, doc := range doctors {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(doc.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password for %s: %v", doc.FullName, err)
			continue
		}

		result, err := db.Exec(insertQuery, doc.IIN, string(hashedPassword), doc.FullName, "doctor")
		if err != nil {
			log.Printf("Error inserting doctor user %s (IIN: %s): %v", doc.FullName, doc.IIN, err)
			continue
		}

		rows, _ := result.RowsAffected()
		if rows > 0 {
			log.Printf("✅ CREATED: %s | IIN: %s | Password: %s", doc.FullName, doc.IIN, doc.Password)
		} else {
			log.Printf("ℹ️ Skipped (already exists): %s | IIN: %s", doc.FullName, doc.IIN)
		}
	}

	// New Admin User Seeding
	log.Println("Seeding admin users with specific IINs and passwords...")

	type adminSeed struct {
		FullName string
		IIN      string
		Password string
	}

	admins := []adminSeed{
		{"Асель Нурмагамбетова", "890215350013", "aseladmin123"},
		{"Тимур Садыков", "870908460014", "timuradmin123"},
		{"Дарья Белова", "860124570015", "daryaadmin123"},
	}

	insertAdminQuery := `
	INSERT INTO users (iin, password_hash, full_name, role)
	VALUES ($1, $2, $3, 'admin')
	ON CONFLICT (iin) DO NOTHING
`

	for _, admin := range admins {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password for admin %s: %v", admin.FullName, err)
			continue
		}

		result, err := db.Exec(insertAdminQuery, admin.IIN, string(hashedPassword), admin.FullName)
		if err != nil {
			log.Printf("Error inserting admin user %s (IIN: %s): %v", admin.FullName, admin.IIN, err)
			continue
		}

		rows, _ := result.RowsAffected()
		if rows > 0 {
			log.Printf("✅ CREATED Admin: %s | IIN: %s | Password: %s", admin.FullName, admin.IIN, admin.Password)
		} else {
			log.Printf("ℹ️ Skipped (already exists): %s | IIN: %s", admin.FullName, admin.IIN)
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
