package usecase

import (
	"fmt"
	"log"
	"telemed/internal/domain"
	"time"
)

type NotificationService struct {
	appointmentRepo domain.AppointmentRepository
	patientRepo     domain.PatientRepository
	telegramBot     TelegramNotifier
}

type TelegramNotifier interface {
	SendNotification(chatID int64, message string) error
	SendVideoLink(chatID int64, appointmentID int) error
}

func NewNotificationService(
	appointmentRepo domain.AppointmentRepository,
	patientRepo domain.PatientRepository,
	telegramBot TelegramNotifier,
) *NotificationService {
	return &NotificationService{
		appointmentRepo: appointmentRepo,
		patientRepo:     patientRepo,
		telegramBot:     telegramBot,
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
	now := time.Now()

	appointments, err := ns.appointmentRepo.GetUpcomingAppointments(now, now.Add(25*time.Hour))
	if err != nil {
		log.Printf("Error getting upcoming appointments: %v", err)
		return
	}

	for _, appt := range appointments {
		timeUntil := appt.StartTime.Sub(now)

		if ns.shouldSendNotification(timeUntil) {
			ns.sendAppointmentNotification(appt, timeUntil)
		}
	}
}

func (ns *NotificationService) shouldSendNotification(timeUntil time.Duration) bool {
	// 24 часа (±2 минуты)
	if timeUntil >= 23*time.Hour+58*time.Minute && timeUntil <= 24*time.Hour+2*time.Minute {
		return true
	}

	// 6 часов (±2 минуты)
	if timeUntil >= 5*time.Hour+58*time.Minute && timeUntil <= 6*time.Hour+2*time.Minute {
		return true
	}

	// 1 час (±2 минуты)
	if timeUntil >= 58*time.Minute && timeUntil <= 1*time.Hour+2*time.Minute {
		return true
	}

	// 30 минут (±2 минуты)
	if timeUntil >= 28*time.Minute && timeUntil <= 32*time.Minute {
		return true
	}

	return false
}

func (ns *NotificationService) sendAppointmentNotification(appt domain.NotificationData, timeUntil time.Duration) {
	if timeUntil >= 28*time.Minute && timeUntil <= 32*time.Minute {
		// Отправляем ссылку на видеозвонок
		err := ns.telegramBot.SendVideoLink(appt.PatientChatID, appt.AppointmentID)
		if err != nil {
			log.Printf("Error sending video link: %v", err)
		}
		return
	}

	// Отправляем текстовое уведомление
	message := ns.formatNotificationMessage(appt, timeUntil)
	err := ns.telegramBot.SendNotification(appt.PatientChatID, message)
	if err != nil {
		log.Printf("Error sending notification: %v", err)
	}
}

func (ns *NotificationService) formatNotificationMessage(appt domain.NotificationData, timeUntil time.Duration) string {
	timeStr := appt.StartTime.Format("02.01.2006 15:04")

	var timeLeft string
	var message string

	if appt.Language == "kz" {
		if timeUntil >= 23*time.Hour {
			timeLeft = "24 сағат"
		} else if timeUntil >= 5*time.Hour {
			timeLeft = "6 сағат"
		} else {
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
	} else {
		if timeUntil >= 23*time.Hour {
			timeLeft = "24 часа"
		} else if timeUntil >= 5*time.Hour {
			timeLeft = "6 часов"
		} else {
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
