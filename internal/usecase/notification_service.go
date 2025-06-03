package usecase

import (
	"fmt"
	"log"
	"telemed/internal/domain"
	"time"
)

type NotificationService struct {
	appointmentRepo   domain.AppointmentRepository
	patientRepo       domain.PatientRepository
	telegramBot       TelegramNotifier
	sentNotifications map[string]bool
}

type TelegramNotifier interface {
	SendNotification(chatID int64, message string) error
	SendVideoLink(chatID int64, appointmentID int) error
	GetUserLanguage(chatID int64) (string, bool)
}

func NewNotificationService(
	appointmentRepo domain.AppointmentRepository,
	patientRepo domain.PatientRepository,
	telegramBot TelegramNotifier,
) *NotificationService {
	return &NotificationService{
		appointmentRepo:   appointmentRepo,
		patientRepo:       patientRepo,
		telegramBot:       telegramBot,
		sentNotifications: make(map[string]bool), // инициализируем
	}
}

func (ns *NotificationService) StartNotificationScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			ns.checkAndSendNotifications()
		}
	}()
}

func (ns *NotificationService) checkAndSendNotifications() {
	// Создаем фиксированную таймзону GMT+5
	location := time.FixedZone("GMT+5", 5*3600)

	// Получаем текущее время в фиксированной таймзоне
	now := time.Now().In(location)
	log.Printf("Local time (GMT+5): %s", now.Format("2006-01-02 15:04 MST"))

	// Для работы с БД используем время в UTC
	nowUTC := now.UTC()
	log.Printf("UTC time: %s", nowUTC.Format("2006-01-02 15:04 MST"))

	// Получаем приемы из БД в интервале [nowUTC, nowUTC+25h]
	appts, err := ns.appointmentRepo.GetUpcomingAppointments(nowUTC, nowUTC.Add(25*time.Hour))
	if err != nil {
		log.Printf("Error fetching appointments: %v", err)
		return
	}
	log.Printf("Found %d appointments", len(appts))

	for _, appt := range appts {
		// Преобразуем время приема в локальное время (GMT+5)
		startLocal := appt.StartTime.In(location)
		timeUntil := startLocal.Sub(now)

		log.Printf(
			"APPOINTMENT %d | startLocal: %s | timeUntil: %v",
			appt.AppointmentID,
			startLocal.Format("2006-01-02 15:04 MST"),
			timeUntil,
		)

		if ns.shouldSendNotification(timeUntil, appt.AppointmentID) {
			ns.sendAppointmentNotification(appt, timeUntil)
		}
	}
}

func (ns *NotificationService) shouldSendNotification(timeUntil time.Duration, apptID int) bool {
	var notifType string

	// Точные интервалы для уведомлений
	switch {
	case isInTimeRange(timeUntil, 24*time.Hour):
		notifType = fmt.Sprintf("%d_24h", apptID)
	case isInTimeRange(timeUntil, 6*time.Hour):
		notifType = fmt.Sprintf("%d_6h", apptID)
	case isInTimeRange(timeUntil, 1*time.Hour):
		notifType = fmt.Sprintf("%d_1h", apptID)
	case isInTimeRange(timeUntil, 30*time.Minute):
		notifType = fmt.Sprintf("%d_30m", apptID)
	default:
		return false
	}

	if ns.sentNotifications[notifType] {
		return false
	}

	ns.sentNotifications[notifType] = true
	return true
}

// Вспомогательная функция для проверки временного интервала
func isInTimeRange(timeUntil, targetDuration time.Duration) bool {
	margin := 5 * time.Minute
	return timeUntil >= targetDuration-margin && timeUntil <= targetDuration+margin
}

func (ns *NotificationService) sendAppointmentNotification(appt domain.NotificationData, timeUntil time.Duration) {
	finalLangCode := appt.Language // Default from DB (e.g., "ru")
	// specificLangFound := false // Optional, for more complex logic or logging if needed

	if ns.telegramBot != nil { // Check if telegramBot is not nil
		langCode, found := ns.telegramBot.GetUserLanguage(appt.PatientChatID)
		if found {
			finalLangCode = langCode
			// specificLangFound = true
		}
	}

	// Check if it's time for the 30-minute notification (video link)
	// This window should be consistent with shouldSendNotification
	if timeUntil >= 25*time.Minute && timeUntil <= 35*time.Minute {
		log.Printf("Attempting to send 30-minute video link notification for AppointmentID: %d to PatientChatID: %d (Language context: %s)", appt.AppointmentID, appt.PatientChatID, finalLangCode)
		err := ns.telegramBot.SendVideoLink(appt.PatientChatID, appt.AppointmentID)
		if err != nil {
			log.Printf("Error sending 30-minute video link for AppointmentID: %d to PatientChatID: %d: %v", appt.AppointmentID, appt.PatientChatID, err)
		} else {
			log.Printf("Successfully sent 30-minute video link for AppointmentID: %d to PatientChatID: %d", appt.AppointmentID, appt.PatientChatID)
		}
		// Video link is sent, no further text notification for this specific 30-min window
		return
	}

	// For other notification windows (1h, 6h, 24h), send a formatted text message
	message := ns.formatNotificationMessage(appt, timeUntil, finalLangCode)

	// Determine notification type for logging
	var notifTypeStr string
	// These windows should match shouldSendNotification logic for non-30m cases
	if timeUntil >= 23*time.Hour+55*time.Minute && timeUntil <= 24*time.Hour+5*time.Minute {
		notifTypeStr = "24-hour"
	} else if timeUntil >= 5*time.Hour+55*time.Minute && timeUntil <= 6*time.Hour+5*time.Minute {
		notifTypeStr = "6-hour"
	} else if timeUntil >= 55*time.Minute && timeUntil <= 1*time.Hour+5*time.Minute {
		notifTypeStr = "1-hour"
	} else {
		notifTypeStr = fmt.Sprintf("unknown type (timeUntil: %v)", timeUntil)
	}

	log.Printf("Attempting to send %s text notification for AppointmentID: %d to PatientChatID: %d (Language: %s)", notifTypeStr, appt.AppointmentID, appt.PatientChatID, finalLangCode)
	err := ns.telegramBot.SendNotification(appt.PatientChatID, message)
	if err != nil {
		log.Printf("Error sending %s text notification for AppointmentID: %d to PatientChatID: %d (Language: %s): %v", notifTypeStr, appt.AppointmentID, appt.PatientChatID, finalLangCode, err)
	} else {
		log.Printf("Successfully sent %s text notification for AppointmentID: %d to PatientChatID: %d (Language: %s)", notifTypeStr, appt.AppointmentID, appt.PatientChatID, finalLangCode)
	}
}

func (ns *NotificationService) formatNotificationMessage(appt domain.NotificationData, timeUntil time.Duration, languageCode string) string {
	timeStr := appt.StartTime.Format("02.01.2006 15:04")

	var timeLeft string
	var message string

	if languageCode == "kz" {
		if timeUntil >= 23*time.Hour {
			timeLeft = "24 сағат"
		} else if timeUntil >= 5*time.Hour {
			timeLeft = "6 сағат"
		} else { // Covers 1-hour notification window
			timeLeft = "1 сағат"
		}

		message = fmt.Sprintf(
			"🔔 Дәрігерге жазылу туралы еске салу\n\n"+
				"Дәрігер: %s\n"+
				"Қызмет: %s\n"+
				"Уақыт: %s\n\n"+
				"Қабылдауға дейін: %s",
			appt.DoctorName,
			appt.ServiceName,
			timeStr,
			timeLeft,
		)
	} else { // Default to Russian or other main language
		if timeUntil >= 23*time.Hour {
			timeLeft = "24 часа"
		} else if timeUntil >= 5*time.Hour {
			timeLeft = "6 часов"
		} else { // Covers 1-hour notification window
			timeLeft = "1 час"
		}

		message = fmt.Sprintf(
			"🔔 Напоминание о записи к врачу\n\n"+
				"Врач: %s\n"+
				"Услуга: %s\n"+
				"Время: %s\n\n"+
				"До приема осталось: %s",
			appt.DoctorName,
			appt.ServiceName,
			timeStr,
			timeLeft,
		)
	}

	return message
}
