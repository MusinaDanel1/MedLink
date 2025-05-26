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

	// Add sample services for the doctor
	var doctorID int
	err = db.QueryRow("SELECT id FROM doctors WHERE full_name = $1", "Марина Цветаева").Scan(&doctorID)
	if err != nil {
		log.Printf("Failed to find doctor 'Марина Цветаева': %v", err)
		// Let's create the doctor if not found
		if err == sql.ErrNoRows {
			log.Printf("Creating doctor 'Марина Цветаева'...")
			_, err = db.Exec(
				"INSERT INTO doctors (full_name, specialization_id) VALUES ($1, $2) RETURNING id",
				"Марина Цветаева", 1,
			)
			if err != nil {
				log.Printf("Failed to create doctor: %v", err)
			} else {
				// Try again to get the ID
				err = db.QueryRow("SELECT id FROM doctors WHERE full_name = $1", "Марина Цветаева").Scan(&doctorID)
				if err != nil {
					log.Printf("Still couldn't get doctor ID: %v", err)
				} else {
					log.Printf("Successfully created doctor with ID: %d", doctorID)
				}
			}
		}
	}

	if doctorID > 0 {
		log.Printf("Found doctor with ID: %d", doctorID)

		// Check existing services first
		rows, err := db.Query("SELECT id, name FROM services WHERE doctor_id = $1", doctorID)
		if err != nil {
			log.Printf("Error checking existing services: %v", err)
		} else {
			var existingServices []string
			for rows.Next() {
				var id int
				var name string
				if err := rows.Scan(&id, &name); err != nil {
					log.Printf("Error scanning service: %v", err)
				} else {
					existingServices = append(existingServices, fmt.Sprintf("%d: %s", id, name))
				}
			}
			rows.Close()

			if len(existingServices) > 0 {
				log.Printf("Doctor already has services: %v", existingServices)
			} else {
				log.Printf("No existing services found for doctor ID %d, adding new ones", doctorID)
			}
		}

		serviceQuery := `INSERT INTO services (doctor_id, name) 
		                 VALUES ($1, $2)
		                 ON CONFLICT (doctor_id, name) DO NOTHING`

		// Add some sample services for this doctor
		services := []string{
			"Консультация",
			"Осмотр",
			"Диагностика",
			"Плановый прием",
		}

		for _, service := range services {
			result, err := db.Exec(serviceQuery, doctorID, service)
			if err != nil {
				log.Printf("Ошибка при добавлении услуги '%s': %v", service, err)
			} else {
				affected, _ := result.RowsAffected()
				if affected > 0 {
					log.Printf("Услуга '%s' добавлена для врача ID %d", service, doctorID)
				} else {
					log.Printf("Услуга '%s' уже существует для врача ID %d", service, doctorID)
				}
			}
		}
	} else {
		log.Printf("Failed to get valid doctor ID")
	}

	fmt.Println("Админ и врач успешно добавлены в базу данных!")

	// Initialize repositories
	authRepo := postgres.NewAuthRepository(db)
	adminRepo := postgres.NewAdminRepository(db)
	patientRepo := postgres.NewPatientRepository(db)
	doctorRepo := postgres.NewDoctorRepository(db)
	appointmentRepo := postgres.NewAppointmentRepository(db)
	msgRepo := postgres.NewMessageRepository(db)
	scheduleRepo := postgres.NewScheduleRepo(db)
	timeslotRepo := postgres.NewTimeslotRepo(db)
	schedRepo := postgres.NewScheduleRepo(db)
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
	schedSvc := usecase.NewScheduleService(schedRepo)

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

	if token != "dummy_token" {
		bot, err = tgbotapi.NewBotAPI(token)
		if err != nil {
			log.Fatal("Ошибка инициализации Telegram-бота: ", err)
		}
		bot.Debug = true
		log.Printf("Запущен бот: %s", bot.Self.UserName)

		// Создаем bot handler БЕЗ notification service (пока nil)
		botHandler = telegram.NewBotHandler(bot, patientService, doctorService, appointmentService, nil, openaiService)

		// ТЕПЕРЬ создаем notification service с bot handler
		notificationService = usecase.NewNotificationService(
			appointmentRepo,
			patientRepo,
			botHandler,
		)

		// Запускаем планировщик уведомлений
		notificationService.StartNotificationScheduler()
	}

	msgHandler := http1.NewMessageHandler(msgService, appointmentService)
	apptHandler := http1.NewAppointmentHandler(appointmentService, doctorService)
	apptHandler.SetBotHandler(botHandler)

	// Run bot updates handling in a separate goroutine
	if bot != nil {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates := bot.GetUpdatesChan(u)

		go func() {
			for update := range updates {
				botHandler.HandleUpdate(update)
			}
		}()
	}

	// Auth routes
	mux := http.NewServeMux()
	mux.HandleFunc("/main", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authHandler.ShowMainForm(w, r)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})
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

	// This code is now unreachable since we're running everything on the Gin server
	// port := "8081" // Changed from 8080 to 8081
	// log.Printf("HTTP Server starting at http://localhost:%s", port)
	// if err := http.ListenAndServe(":"+port, mux); err != nil {
	// 	log.Fatal(err)
	// }
}
