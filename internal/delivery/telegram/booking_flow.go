package telegram

import (
	"sort"
	"strconv"
	"strings"
	"telemed/internal/usecase"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TimeSlot struct {
	ID         int
	ScheduleID int
	StartTime  time.Time
	EndTime    time.Time
}

func (h *BotHandler) handleCallback(cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	data := cb.Data
	lang := h.getUserLanguage(chatID)

	switch {
	case data == "book_appointment":
		h.handleBookingStart(chatID)

	case strings.HasPrefix(data, "spec_"):
		h.handleSpecSelected(chatID, data)

	case strings.HasPrefix(data, "doc_"):
		h.handleDoctorSelected(chatID, data)

	case strings.HasPrefix(data, "serv_"):
		h.handleServiceSelected(chatID, data)

	case strings.HasPrefix(data, "date_"):
		h.handleDateSelected(chatID, data)

	case strings.HasPrefix(data, "timeslot_"):
		h.handleTimeslotSelected(chatID, data)

	case data == "confirm_yes":
		h.handleBookingConfirm(chatID, true)

	case data == "confirm_no":
		h.handleBookingConfirm(chatID, false)

	case data == "ai_consultation":
		h.state[chatID] = "ai_consultation_waiting"
		msg := tgbotapi.NewMessage(chatID, h.loc.Get(lang, "ai_consultation_prompt"))
		h.bot.Send(msg)
	}

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
}

func (h *BotHandler) handleBookingStart(chatID int64) {
	lang := h.getUserLanguage(chatID)
	specs, err := h.doctorService.GetAllSpecializations()
	if err != nil || len(specs) == 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "no_specializations")))
		return
	}
	if h.temp[chatID] == nil {
		h.temp[chatID] = make(map[string]string)
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, s := range specs {
		rows = append(rows,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					s.Name, "spec_"+strconv.Itoa(s.ID),
				),
			),
		)
	}

	msg := tgbotapi.NewMessage(chatID, h.loc.Get(lang, "choose_specialization"))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.state[chatID] = "booking_specialization"
	h.bot.Send(msg)
}

func (h *BotHandler) handleSpecSelected(chatID int64, data string) {
	lang := h.getUserLanguage(chatID)
	parts := strings.Split(data, "_")
	specID, err := strconv.Atoi(parts[1])
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "invalid_specialization")))
		return
	}

	h.temp[chatID]["spec_id"] = parts[1]

	for _, s := range must(h.doctorService.GetAllSpecializations()) {
		if s.ID == specID {
			h.temp[chatID]["spec_name"] = s.Name
			break
		}
	}

	docs, err := h.doctorService.GetDoctorsBySpecialization(specID)
	if err != nil || len(docs) == 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "no_doctors")))
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, d := range docs {
		rows = append(rows,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					d.FullName, "doc_"+strconv.Itoa(d.ID),
				),
			),
		)
	}

	msg := tgbotapi.NewMessage(chatID, h.loc.Get(lang, "choose_doctor"))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.state[chatID] = "booking_doctor"
	h.bot.Send(msg)
}

func (h *BotHandler) handleDoctorSelected(chatID int64, data string) {
	lang := h.getUserLanguage(chatID)
	parts := strings.Split(data, "_")
	docID, err := strconv.Atoi(parts[1])
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "invalid_doctor")))
		return
	}

	h.temp[chatID]["doctor_id"] = parts[1]

	for _, d := range must(
		h.doctorService.GetDoctorsBySpecialization(
			mustAtoi(h.temp[chatID]["spec_id"]),
		),
	) {
		if d.ID == docID {
			h.temp[chatID]["doctor_name"] = d.FullName
			break
		}
	}

	services, err := h.doctorService.GetServicesByDoctor(docID)
	if err != nil || len(services) == 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "no_services")))
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, s := range services {
		rows = append(rows,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					s.Name, "serv_"+strconv.Itoa(s.ID),
				),
			),
		)
	}

	msg := tgbotapi.NewMessage(chatID, h.loc.Get(lang, "choose_service"))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.state[chatID] = "booking_service"
	h.bot.Send(msg)
}

func (h *BotHandler) handleServiceSelected(chatID int64, data string) {
	lang := h.getUserLanguage(chatID)
	parts := strings.Split(data, "_")
	servID, err := strconv.Atoi(parts[1])
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "invalid_service")))
		return
	}

	h.temp[chatID]["service_id"] = parts[1]

	for _, s := range must(
		h.doctorService.GetServicesByDoctor(
			mustAtoi(h.temp[chatID]["doctor_id"]),
		),
	) {
		if s.ID == servID {
			h.temp[chatID]["service_name"] = s.Name
			break
		}
	}

	h.showAvailableDates(chatID)
}

func (h *BotHandler) showAvailableDates(chatID int64) {
	lang := h.getUserLanguage(chatID)
	docID := mustAtoi(h.temp[chatID]["doctor_id"])

	slots, err := h.doctorService.GetAvailableTimeSlots(docID)
	if err != nil || len(slots) == 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "no_timeslots")))
		return
	}

	// Группируем слоты по датам
	dateMap := make(map[string][]interface{})
	for _, slot := range slots {
		dateStr := slot.StartTime.Format("2006-01-02")
		dateMap[dateStr] = append(dateMap[dateStr], slot)
	}

	if len(dateMap) == 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "no_available_dates")))
		return
	}

	var dates []string
	for dateStr := range dateMap {
		dates = append(dates, dateStr)
	}
	sort.Strings(dates)

	// Создаем кнопки для дат
	var rows [][]tgbotapi.InlineKeyboardButton
	for dateStr := range dateMap {
		date, _ := time.Parse("2006-01-02", dateStr)
		displayDate := date.Format("02.01.2006")

		rows = append(rows,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					displayDate, "date_"+dateStr,
				),
			),
		)
	}

	msg := tgbotapi.NewMessage(chatID, h.loc.Get(lang, "choose_date"))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.state[chatID] = "booking_date"
	h.bot.Send(msg)
}

// Новая функция для обработки выбора даты
func (h *BotHandler) handleDateSelected(chatID int64, data string) {
	lang := h.getUserLanguage(chatID)
	parts := strings.Split(data, "_")
	if len(parts) != 2 {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "invalid_command")))
		return
	}

	selectedDate := parts[1]
	h.temp[chatID]["selected_date"] = selectedDate

	// Получаем слоты для выбранной даты
	docID := mustAtoi(h.temp[chatID]["doctor_id"])
	slots, err := h.doctorService.GetAvailableTimeSlots(docID)
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "no_timeslots")))
		return
	}

	// Фильтруем слоты по выбранной дате и правильно типизируем
	var daySlots []TimeSlot
	for _, slot := range slots {
		slotDate := slot.StartTime.Format("2006-01-02")
		if slotDate == selectedDate {
			daySlots = append(daySlots, TimeSlot{
				ID:         slot.ID,
				ScheduleID: slot.ScheduleID,
				StartTime:  slot.StartTime,
				EndTime:    slot.EndTime,
			})
		}
	}

	if len(daySlots) == 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "no_slots_for_date")))
		return
	}

	// Сортируем слоты по времени
	sort.Slice(daySlots, func(i, j int) bool {
		return daySlots[i].StartTime.Before(daySlots[j].StartTime)
	})

	// Создаем кнопки для времени
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, slot := range daySlots {
		timeStr := slot.StartTime.Format("15:04")
		rows = append(rows,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					timeStr, "timeslot_"+strconv.Itoa(slot.ID),
				),
			),
		)
	}

	// Форматируем дату для отображения
	date, _ := time.Parse("2006-01-02", selectedDate)
	displayDate := date.Format("02.01.2006")

	msg := tgbotapi.NewMessage(chatID, h.loc.Get(lang, "choose_time_for_date")+" "+displayDate+":")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.state[chatID] = "booking_timeslot"
	h.bot.Send(msg)
}

func (h *BotHandler) handleTimeslotSelected(chatID int64, data string) {
	lang := h.getUserLanguage(chatID)
	parts := strings.Split(data, "_")
	tsID := mustAtoi(parts[1])

	h.temp[chatID]["timeslot_id"] = parts[1]

	for _, t := range must(
		h.doctorService.GetAvailableTimeSlots(
			mustAtoi(h.temp[chatID]["doctor_id"]),
		),
	) {
		if t.ID == tsID {
			h.temp[chatID]["timeslot_time"] = t.StartTime.Format("02.01.2006 15:04")
			break
		}
	}

	text := h.loc.Get(lang, "confirm_appointment") + "\n" +
		h.loc.Get(lang, "specialization") + " " + h.temp[chatID]["spec_name"] + "\n" +
		h.loc.Get(lang, "doctor") + " " + h.temp[chatID]["doctor_name"] + "\n" +
		h.loc.Get(lang, "service") + " " + h.temp[chatID]["service_name"] + "\n" +
		h.loc.Get(lang, "time") + " " + h.temp[chatID]["timeslot_time"]

	yes := tgbotapi.NewInlineKeyboardButtonData(h.loc.Get(lang, "yes"), "confirm_yes")
	no := tgbotapi.NewInlineKeyboardButtonData(h.loc.Get(lang, "no"), "confirm_no")
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(yes, no),
	)

	h.state[chatID] = "booking_confirm"
	h.bot.Send(msg)
}

func (h *BotHandler) handleBookingConfirm(chatID int64, ok bool) {
	lang := h.getUserLanguage(chatID)

	if !ok {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "appointment_cancelled")))
		delete(h.state, chatID)
		delete(h.temp, chatID)
		h.sendMainMenu(chatID)
		return
	}

	patientID, err := h.patientService.GetIDByChatID(chatID)
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "patient_data_error")))
		return
	}

	docID := mustAtoi(h.temp[chatID]["doctor_id"])
	tsID := mustAtoi(h.temp[chatID]["timeslot_id"])

	// First get the timeslot to access its schedule ID and times
	slots, err := h.doctorService.GetAvailableTimeSlots(docID)
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "slot_data_error")))
		return
	}

	// Find the selected timeslot
	var scheduleID int
	var startTime, endTime time.Time
	for _, slot := range slots {
		if slot.ID == tsID {
			scheduleID = slot.ScheduleID
			startTime = slot.StartTime
			endTime = slot.EndTime
			break
		}
	}

	apptID, err := h.appointmentService.BookAppointment(
		scheduleID, patientID, startTime, endTime,
	)

	if err != nil {
		if err == usecase.ErrSlotBooked {
			h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "slot_taken")))
		} else {
			h.bot.Send(tgbotapi.NewMessage(chatID, h.loc.Get(lang, "booking_error")+" "+err.Error()))
		}
		return
	}

	// Подтверждаем пользователю
	successText := h.loc.Get(lang, "appointment_success") + "\n" +
		h.loc.Get(lang, "doctor") + " " + h.temp[chatID]["doctor_name"] + "\n" +
		h.loc.Get(lang, "service") + " " + h.temp[chatID]["service_name"] + "\n" +
		h.loc.Get(lang, "time") + " " + h.temp[chatID]["timeslot_time"]

	h.bot.Send(tgbotapi.NewMessage(chatID, successText))

	// Генерим ссылку на видео-сессию и отправляем
	h.sendVideoLink(chatID, apptID)

	delete(h.state, chatID)
	delete(h.temp, chatID)
	h.sendMainMenu(chatID)
}

// Helper functions
func must[T any](v []T, err error) []T {
	if err != nil {
		return nil
	}
	return v
}

func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}
