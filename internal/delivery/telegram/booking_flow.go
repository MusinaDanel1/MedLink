package telegram

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *BotHandler) handleCallback(cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	data := cb.Data

	switch {
	case data == "book_appointment":
		h.handleBookingStart(chatID)

	case strings.HasPrefix(data, "spec_"):
		h.handleSpecSelected(chatID, data)

	case strings.HasPrefix(data, "doc_"):
		h.handleDoctorSelected(chatID, data)

	case strings.HasPrefix(data, "serv_"):
		h.handleServiceSelected(chatID, data)

	case strings.HasPrefix(data, "timeslot_"):
		h.handleTimeslotSelected(chatID, data)

	case data == "confirm_yes":
		h.handleBookingConfirm(chatID, true)

	case data == "confirm_no":
		h.handleBookingConfirm(chatID, false)

	case data == "ai_consultation":
		h.bot.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, опишите вашу жалобу."))
	}

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
}

func (h *BotHandler) handleBookingStart(chatID int64) {
	specs, err := h.doctor.GetAllSpecializations()
	if err != nil || len(specs) == 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID, "Нет доступных специализаций."))
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

	msg := tgbotapi.NewMessage(chatID, "Выберите специализацию:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.state[chatID] = "booking_specialization"
	h.bot.Send(msg)
}

func (h *BotHandler) handleSpecSelected(chatID int64, data string) {
	parts := strings.Split(data, "_")
	specID, err := strconv.Atoi(parts[1])
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, "Не удалось распознать специализацию."))
		return
	}

	h.temp[chatID]["spec_id"] = parts[1]

	for _, s := range must(h.doctor.GetAllSpecializations()) {
		if s.ID == specID {
			h.temp[chatID]["spec_name"] = s.Name
			break
		}
	}

	docs, err := h.doctor.GetDoctorsBySpecialization(specID)
	if err != nil || len(docs) == 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID,
			"Нет врачей для выбранной специализации."))
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

	msg := tgbotapi.NewMessage(chatID, "Выберите врача:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.state[chatID] = "booking_doctor"
	h.bot.Send(msg)
}

func (h *BotHandler) handleDoctorSelected(chatID int64, data string) {
	parts := strings.Split(data, "_")
	docID, err := strconv.Atoi(parts[1])
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, "Не удалось распознать врача."))
		return
	}

	h.temp[chatID]["doctor_id"] = parts[1]

	for _, d := range must(
		h.doctor.GetDoctorsBySpecialization(
			mustAtoi(h.temp[chatID]["spec_id"]),
		),
	) {
		if d.ID == docID {
			h.temp[chatID]["doctor_name"] = d.FullName
			break
		}
	}

	services, err := h.doctor.GetServicesByDoctor(docID)
	if err != nil || len(services) == 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID,
			"У выбранного врача нет доступных услуг."))
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

	msg := tgbotapi.NewMessage(chatID, "Выберите услугу:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.state[chatID] = "booking_service"
	h.bot.Send(msg)
}

func (h *BotHandler) handleServiceSelected(chatID int64, data string) {
	parts := strings.Split(data, "_")
	servID, err := strconv.Atoi(parts[1])
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, "Не удалось распознать услугу."))
		return
	}

	h.temp[chatID]["service_id"] = parts[1]

	for _, s := range must(
		h.doctor.GetServicesByDoctor(
			mustAtoi(h.temp[chatID]["doctor_id"]),
		),
	) {
		if s.ID == servID {
			h.temp[chatID]["service_name"] = s.Name
			break
		}
	}

	docID := mustAtoi(h.temp[chatID]["doctor_id"])
	slots, err := h.doctor.GetAvailableTimeSlots(docID)
	if err != nil || len(slots) == 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID,
			"Нет доступных тайм-слотов."))
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, t := range slots {
		rows = append(rows,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					t.AppointmentTime.Format("02.01.2006 15:04"),
					"timeslot_"+strconv.Itoa(t.ID),
				),
			),
		)
	}

	msg := tgbotapi.NewMessage(chatID, "Выберите время:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.state[chatID] = "booking_timeslot"
	h.bot.Send(msg)
}

func (h *BotHandler) handleTimeslotSelected(chatID int64, data string) {
	parts := strings.Split(data, "_")
	tsID := mustAtoi(parts[1])

	h.temp[chatID]["timeslot_id"] = parts[1]

	for _, t := range must(
		h.doctor.GetAvailableTimeSlots(
			mustAtoi(h.temp[chatID]["doctor_id"]),
		),
	) {
		if t.ID == tsID {
			h.temp[chatID]["timeslot_time"] = t.AppointmentTime.Format("02.01.2006 15:04")
			break
		}
	}

	text := "Подтвердите прием:\n" +
		"Специализация: " + h.temp[chatID]["spec_name"] + "\n" +
		"Врач: " + h.temp[chatID]["doctor_name"] + "\n" +
		"Услуга: " + h.temp[chatID]["service_name"] + "\n" +
		"Время: " + h.temp[chatID]["timeslot_time"]

	yes := tgbotapi.NewInlineKeyboardButtonData("✅ Да", "confirm_yes")
	no := tgbotapi.NewInlineKeyboardButtonData("❌ Нет", "confirm_no")
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(yes, no),
	)

	h.state[chatID] = "booking_confirm"
	h.bot.Send(msg)
}

func (h *BotHandler) handleBookingConfirm(chatID int64, ok bool) {
	if !ok {
		h.bot.Send(tgbotapi.NewMessage(chatID, "Запись отменена."))
		delete(h.state, chatID)
		delete(h.temp, chatID)
		h.sendMainMenu(chatID)
		return
	}
	patientID, err := h.patient.GetIDByChatID(chatID)
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID,
			"Не удалось получить данные пациента."))
		return
	}
	docID := mustAtoi(h.temp[chatID]["doctor_id"])
	tsID := mustAtoi(h.temp[chatID]["timeslot_id"])
	serviceID := mustAtoi(h.temp[chatID]["service_id"])

	if err := h.appointment.BookAppointment(patientID, docID, serviceID, tsID); err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID,
			"Ошибка при записи. Попробуйте позже."))
	} else {
		h.bot.Send(tgbotapi.NewMessage(chatID,
			"Вы успешно записаны!\n"+
				"Врач: "+h.temp[chatID]["doctor_name"]+"\n"+
				"Услуга: "+h.temp[chatID]["service_name"]+"\n"+
				"Время: "+h.temp[chatID]["timeslot_time"],
		))
	}

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
