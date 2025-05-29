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

	var doctorID int
	err = db.QueryRow("SELECT id FROM doctors WHERE full_name = $1", "Марина Цветаева").Scan(&doctorID)
	if err != nil {
		log.Printf("Failed to find doctor 'Марина Цветаева': %v", err)
		if err == sql.ErrNoRows {
			log.Printf("Creating doctor 'Марина Цветаева'...")
			_, err = db.Exec(
				"INSERT INTO doctors (full_name, specialization_id) VALUES ($1, $2) RETURNING id",
				"Марина Цветаева", 1,
			)
			if err != nil {
				log.Printf("Failed to create doctor: %v", err)
			} else {
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

	// --- BEGIN New User Seeding Logic ---
	log.Println("Starting new user seeding...")

	randAlphanum := func(length int) string {
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		seededRand := rand.New(rand.NewSource(time.Now().UnixNano())) // Ensure new seed each call for variety in quick succession
		b := make([]byte, length)
		for i := range b {
			b[i] = charset[seededRand.Intn(len(charset))]
		}
		return string(b)
	}

	allDoctorNamesForUserSeeding := []string{
		"Марина Цветаева",
		"Айгуль Омарова",
		"Жанат Мусаев",
		"Григорьев Максим Валерьевич",
		"Фёдорова Алина Романовна",
		"Степанов Арсений Кириллович",
		"Беляева София Львовна",
		"Андреев Даниил Артёмович",
		"Виноградова Полина Глебовна",
		"Богданов Марк Денисович",
		"Комарова Ева Ярославовна",
		"Киселёв Лев Игоревич",
		"Абрамова Милана Эмировна",
		"Тихонов Руслан Альбертович",
		"Мельникова Вероника Макаровна",
		"Щербаков Глеб Робертович",
		"Кузьмина Ульяна Давидовна",
	}

	var iinCounter int64 = 100000000000

	userInsertQuery := `INSERT INTO users (iin, password_hash, full_name, role) 
	                    VALUES ($1, $2, $3, $4)
	                    ON CONFLICT (iin) DO NOTHING`

	log.Println("Seeding new doctor users...")
	for _, name := range allDoctorNamesForUserSeeding {
		var currentIIN string
		if name == "Марина Цветаева" {
			currentIIN = "040831650398"
		} else {
			currentIIN = fmt.Sprintf("%012d", iinCounter)
			iinCounter++
		}

		newPassword := randAlphanum(12)
		newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password for doctor %s: %v", name, err)
			continue
		}

		_, err = db.Exec(userInsertQuery, currentIIN, string(newHashedPassword), name, "doctor")
		if err != nil {
			log.Printf("Error inserting/verifying doctor user %s (IIN: %s): %v", name, currentIIN, err)
		} else {
			log.Printf("User entry for doctor %s (IIN: %s) processed.", name, currentIIN)
		}
	}

	log.Println("Seeding new admin users...")
	for i := 1; i <= 3; i++ {
		adminFullName := fmt.Sprintf("Admin User %d", i)
		newIIN := fmt.Sprintf("%012d", iinCounter)
		iinCounter++

		newPassword := randAlphanum(12)
		newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password for admin %s: %v", adminFullName, err)
			continue
		}

		_, err = db.Exec(userInsertQuery, newIIN, string(newHashedPassword), adminFullName, "admin")
		if err != nil {
			log.Printf("Error inserting admin user %s (IIN: %s): %v", adminFullName, newIIN, err)
		} else {
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
	schedSvc := usecase.NewScheduleService(scheduleRepo)

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

		botHandler = telegram.NewBotHandler(bot, patientService, doctorService, appointmentService, nil, openaiService)

		notificationService = usecase.NewNotificationService(
			appointmentRepo,
			patientRepo,
			botHandler,
		)

		botHandler.SetNotificationService(notificationService)
		notificationService.StartNotificationScheduler()
	}

	msgHandler := http1.NewMessageHandler(msgService, appointmentService)
	apptHandler := http1.NewAppointmentHandler(appointmentService, doctorService)
	apptHandler.SetBotHandler(botHandler)

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

	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1", "192.168.0.0/16", "10.0.0.0/8"})

	r.GET("/ws", video.SignalingHandler)

	r.Static("/static", "./templates/static")

	r.StaticFile("/webrtc/room.html", "./templates/appointment.html")

	r.StaticFile("/debug.html", "./templates/debug.html")

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
	r.GET("/api/appointments/:id/status", apptHandler.GetAppointmentStatus)
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
