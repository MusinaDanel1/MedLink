package usecase

import (
	"fmt"
	"log"
	"telemed/internal/domain"
	"time"
)

// NotificationService отвечает за отправку уведомлений
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
		sentNotifications: make(map[string]bool),
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

	// Получаем текущее локальное время в этой зоне
	now := time.Now().In(location)
	log.Printf("Local time (GMT+5): %s", now.Format("2006-01-02 15:04 MST"))

	// Для выборки из базы используем время в UTC
	nowUTC := now.UTC()
	log.Printf("UTC time: %s", nowUTC.Format("2006-01-02 15:04 MST"))

	// Получаем приемы из БД (предполагается, что время в БД хранится как локальное, т.е. 12:48)
	appts, err := ns.appointmentRepo.GetUpcomingAppointments(nowUTC, nowUTC.Add(25*time.Hour))
	if err != nil {
		log.Printf("Error fetching appointments: %v", err)
		return
	}
	log.Printf("Found %d appointments", len(appts))

	for _, appt := range appts {
		startLocal := time.Date(
			appt.StartTime.Year(),
			appt.StartTime.Month(),
			appt.StartTime.Day(),
			appt.StartTime.Hour(),
			appt.StartTime.Minute(),
			appt.StartTime.Second(),
			appt.StartTime.Nanosecond(),
			location,
		)

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

	// Определяем допустимые интервалы для уведомлений с допуском ±5 минут
	switch {
	case isInTimeRange(timeUntil, 24*time.Hour):
		notifType = fmt.Sprintf("%d_24h", apptID)
	case isInTimeRange(timeUntil, 6*time.Hour):
		notifType = fmt.Sprintf("%d_6h", apptID)
	case isInTimeRange(timeUntil, 1*time.Hour):
		notifType = fmt.Sprintf("%d_1h", apptID)
	case isInTimeRange(timeUntil, 30*time.Minute):
		notifType = fmt.Sprintf("%d_30m", apptID)
	case isInTimeRange(timeUntil, 5*time.Minute):
		notifType = fmt.Sprintf("%d_5m", apptID)
	default:
		return false
	}

	// Если уведомление уже отправлено, повторно его не шлем
	if ns.sentNotifications[notifType] {
		return false
	}

	ns.sentNotifications[notifType] = true
	return true
}

// Вспомогательная функция: проверяет, попадает ли timeUntil в интервал targetDuration ± 5 минут.
func isInTimeRange(timeUntil, targetDuration time.Duration) bool {
	const margin = 0 * time.Minute
	return timeUntil >= targetDuration-margin && timeUntil <= targetDuration+margin
}

func (ns *NotificationService) sendAppointmentNotification(
	appt domain.NotificationData,
	timeUntil time.Duration,
) {
	finalLangCode := appt.Language
	if ns.telegramBot != nil {
		if langCode, found := ns.telegramBot.GetUserLanguage(appt.PatientChatID); found {
			finalLangCode = langCode
		}
	}

	message := ns.formatNotificationMessage(appt, timeUntil, finalLangCode)
	var notifTypeStr string
	switch {
	case isInTimeRange(timeUntil, 24*time.Hour):
		notifTypeStr = "24-hour"
	case isInTimeRange(timeUntil, 6*time.Hour):
		notifTypeStr = "6-hour"
	case isInTimeRange(timeUntil, 1*time.Hour):
		notifTypeStr = "1-hour"
	case isInTimeRange(timeUntil, 30*time.Minute):
		notifTypeStr = "30-minute"
	case isInTimeRange(timeUntil, 5*time.Minute):
		notifTypeStr = "5-minute"
	default:
		notifTypeStr = fmt.Sprintf("unknown (timeUntil: %v)", timeUntil)
	}

	log.Printf("Sending %s text notification for AppointmentID: %d to PatientChatID: %d (Language: %s)",
		notifTypeStr, appt.AppointmentID, appt.PatientChatID, finalLangCode)
	if err := ns.telegramBot.SendNotification(appt.PatientChatID, message); err != nil {
		log.Printf("Error sending %s text notification for AppointmentID: %d to PatientChatID: %d (Language: %s): %v",
			notifTypeStr, appt.AppointmentID, appt.PatientChatID, finalLangCode, err)
	} else {
		log.Printf("Successfully sent %s text notification for AppointmentID: %d to PatientChatID: %d (Language: %s)",
			notifTypeStr, appt.AppointmentID, appt.PatientChatID, finalLangCode)
	}
}

func (ns *NotificationService) formatNotificationMessage(
	appt domain.NotificationData,
	timeUntil time.Duration,
	languageCode string,
) string {
	// Форматируем время приема, как оно хранится в БД, например: "03.06.2025 12:48"
	meetingTimeStr := appt.StartTime.Format("02.01.2006 15:04")
	var message string

	if languageCode == "kz" {
		if isInTimeRange(timeUntil, 24*time.Hour) {
			message = fmt.Sprintf(
				"🔔 Дәрігерге жазылу туралы еске салу\n\n"+
					"Дәрігер: %s\n"+
					"Қызмет: %s\n"+
					"Уақыт: %s\n\n"+
					"Қабылдауға дейін: 24 сағат",
				appt.DoctorName, appt.ServiceName, meetingTimeStr,
			)
		} else if isInTimeRange(timeUntil, 6*time.Hour) {
			message = fmt.Sprintf(
				"🔔 Дәрігерге жазылу туралы еске салу\n\n"+
					"Дәрігер: %s\n"+
					"Қызмет: %s\n"+
					"Уақыт: %s\n\n"+
					"Қабылдауға дейін: 6 сағат",
				appt.DoctorName, appt.ServiceName, meetingTimeStr,
			)
		} else if isInTimeRange(timeUntil, 1*time.Hour) {
			message = fmt.Sprintf(
				"🔔 Дәрігерге жазылу туралы еске салу\n\n"+
					"Дәрігер: %s\n"+
					"Қызмет: %s\n"+
					"Уақыт: %s\n\n"+
					"Қабылдауға дейін: 1 сағат",
				appt.DoctorName, appt.ServiceName, meetingTimeStr,
			)
		} else if isInTimeRange(timeUntil, 30*time.Minute) {
			message = fmt.Sprintf(
				"🔔 Дәрігерге жазылу туралы еске салу\n\n"+
					"Дәрігер: %s\n"+
					"Қызмет: %s\n"+
					"Уақыт: %s\n\n"+
					"Қабылдауға дейін: 30 минут",
				appt.DoctorName, appt.ServiceName, meetingTimeStr,
			)
		} else if isInTimeRange(timeUntil, 5*time.Minute) {
			message = fmt.Sprintf(
				"🔔 Дәрігерге жазылу туралы еске салу\n\n"+
					"Дәрігер: %s\n"+
					"Қызмет: %s\n"+
					"Уақыт: %s\n\n"+
					"5 минуттан кейін қабылдау басталады. Өтінеміз, дайындықты тексеріп, төмендегі сілтеме арқылы қосылыңыз:",
				appt.DoctorName, appt.ServiceName, meetingTimeStr,
			)
		} else {
			message = "Белгісіз уақыт аралығы үшін ескерту хабарламасы."
		}
	} else { // по умолчанию русский язык
		if isInTimeRange(timeUntil, 24*time.Hour) {
			message = fmt.Sprintf(
				"🔔 Напоминание о записи к врачу\n\n"+
					"Врач: %s\n"+
					"Услуга: %s\n"+
					"Время: %s\n\n"+
					"До приема осталось: 24 часа",
				appt.DoctorName, appt.ServiceName, meetingTimeStr,
			)
		} else if isInTimeRange(timeUntil, 6*time.Hour) {
			message = fmt.Sprintf(
				"🔔 Напоминание о записи к врачу\n\n"+
					"Врач: %s\n"+
					"Услуга: %s\n"+
					"Время: %s\n\n"+
					"До приема осталось: 6 часов",
				appt.DoctorName, appt.ServiceName, meetingTimeStr,
			)
		} else if isInTimeRange(timeUntil, 1*time.Hour) {
			message = fmt.Sprintf(
				"🔔 Напоминание о записи к врачу\n\n"+
					"Врач: %s\n"+
					"Услуга: %s\n"+
					"Время: %s\n\n"+
					"До приема осталось: 1 час",
				appt.DoctorName, appt.ServiceName, meetingTimeStr,
			)
		} else if isInTimeRange(timeUntil, 30*time.Minute) {
			message = fmt.Sprintf(
				"🔔 Напоминание о записи к врачу\n\n"+
					"Врач: %s\n"+
					"Услуга: %s\n"+
					"Время: %s\n\n"+
					"До приема осталось: 30 минут",
				appt.DoctorName, appt.ServiceName, meetingTimeStr,
			)
		} else if isInTimeRange(timeUntil, 5*time.Minute) {
			message = fmt.Sprintf(
				"🔔 Напоминание о записи к врачу\n\n"+
					"Врач: %s\n"+
					"Услуга: %s\n"+
					"Время: %s\n\n"+
					"Через 5 минут начнется прием. Пожалуйста, проверьте готовность оборудования и подключитесь к видеозвонку",
				appt.DoctorName, appt.ServiceName, meetingTimeStr,
			)
		} else {
			message = "Неизвестное окно времени для уведомления."
		}
	}

	return message
}
