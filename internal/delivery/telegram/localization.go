package telegram

type Language string

const (
	LangRussian Language = "ru"
	LangKazakh  Language = "kz"
)

type Localization struct {
	texts map[Language]map[string]string
}

func NewLocalization() *Localization {
	return &Localization{
		texts: map[Language]map[string]string{
			LangRussian: {
				"choose_language":         "Выберите язык:",
				"russian":                 "🇷🇺 Русский",
				"kazakh":                  "🇰🇿 Қазақша",
				"language_selected":       "Язык выбран: Русский",
				"registration_required":   "Для начала работы необходимо зарегистрироваться.\n\nПожалуйста, введите ваше ФИО:",
				"enter_name":              "Пожалуйста, введите ваше ФИО:",
				"enter_iin":               "Введите ваш ИИН:",
				"registration_success":    "Вы успешно зарегистрированы!",
				"registration_error":      "Ошибка регистрации. Попробуйте позже.",
				"main_menu_question":      "Что вы хотите сделать?",
				"book_appointment":        "📅 Записаться к врачу",
				"ai_consultation":         "💬 Консультация с ИИ",
				"choose_specialization":   "Выберите специализацию:",
				"choose_doctor":           "Выберите врача:",
				"choose_service":          "Выберите услугу:",
				"choose_time":             "Выберите время:",
				"confirm_appointment":     "Подтвердите прием:",
				"specialization":          "Специализация:",
				"doctor":                  "Врач:",
				"service":                 "Услуга:",
				"time":                    "Время:",
				"yes":                     "✅ Да",
				"no":                      "❌ Нет",
				"appointment_cancelled":   "Запись отменена.",
				"appointment_success":     "Вы успешно записаны!",
				"video_link_message":      "В назначенное время, пожалуйста, подключитесь к видеоконсультации, перейдя по нижеуказанной ссылке. Перед звонком убедитесь, что ваше устройство, камера, микрофон и интернет-соединение работают корректно. Рекомендуем провести тест оборудования заранее для обеспечения бесперебойного приёма.\n\nСсылка для подключения к звонку:",
				"ai_consultation_prompt":  "Пожалуйста, опишите вашу жалобу, и я проконсультирую вас.",
				"ai_error":                "Ошибка при обращении к ИИ:",
				"no_specializations":      "Нет доступных специализаций.",
				"no_doctors":              "Нет врачей для выбранной специализации.",
				"no_services":             "У выбранного врача нет доступных услуг.",
				"no_timeslots":            "Нет доступных тайм-слотов.",
				"slot_taken":              "Извините, этот слот уже занят.",
				"booking_error":           "Ошибка при записи:",
				"invalid_command":         "Пожалуйста, выберите действие из меню или введите /start для начала.",
				"invalid_specialization":  "Не удалось распознать специализацию.",
				"invalid_doctor":          "Не удалось распознать врача.",
				"invalid_service":         "Не удалось распознать услугу.",
				"patient_data_error":      "Не удалось получить данные пациента.",
				"slot_data_error":         "Ошибка при получении данных слота.",
				"name_validation_empty":   "Введите ваше ФИО",
				"name_validation_short":   "ФИО слишком короткое",
				"name_validation_long":    "ФИО слишком длинное",
				"name_validation_invalid": "ФИО должно содержать только буквы, пробелы и дефисы",
				"name_validation_words":   "Введите минимум два слова (Фамилия и Имя)",
				"name_validation_capital": "Каждое слово должно начинаться с заглавной буквы",
				"iin_validation_empty":    "Введите ИИН",
				"iin_validation_length":   "ИИН должен содержать 12 цифр",
				"iin_validation_digits":   "ИИН должен содержать только цифры",
				"iin_validation_invalid":  "Некорректный ИИН",
				"choose_date":             "Выберите дату:",
				"choose_time_for_date":    "Выберите время на",
				"no_available_dates":      "Нет доступных дат.",
				"no_slots_for_date":       "На выбранную дату нет доступных слотов.",
				"end_chat":                "❌ Завершить чат",
			},
			LangKazakh: {
				"choose_language":         "Тілді таңдаңыз:",
				"russian":                 "🇷🇺 Русский",
				"kazakh":                  "🇰🇿 Қазақша",
				"language_selected":       "Тіл таңдалды: Қазақша",
				"registration_required":   "Жұмысты бастау үшін тіркелу қажет.\n\nТолық атыңызды енгізіңіз:",
				"enter_name":              "Толық атыңызды енгізіңіз:",
				"enter_iin":               "ЖСН енгізіңіз:",
				"registration_success":    "Сіз сәтті тіркелдіңіз!",
				"registration_error":      "Тіркелу қатесі. Кейінірек қайталаңыз.",
				"main_menu_question":      "Не істегіңіз келеді?",
				"book_appointment":        "📅 Дәрігерге жазылу",
				"ai_consultation":         "💬 ИИ кеңесі",
				"choose_specialization":   "Мамандықты таңдаңыз:",
				"choose_doctor":           "Дәрігерді таңдаңыз:",
				"choose_service":          "Қызметті таңдаңыз:",
				"choose_time":             "Уақытты таңдаңыз:",
				"confirm_appointment":     "Қабылдауды растаңыз:",
				"specialization":          "Мамандық:",
				"doctor":                  "Дәрігер:",
				"service":                 "Қызмет:",
				"time":                    "Уақыт:",
				"yes":                     "✅ Иә",
				"no":                      "❌ Жоқ",
				"appointment_cancelled":   "Жазылу болдырылмады.",
				"appointment_success":     "Сіз сәтті жазылдыңыз!",
				"video_link_message":      "Белгіленген уақытта, өтінеміз, төменде көрсетілген сілтемеге өтіп, бейне-консультацияға қосылыңыз. Қоңырау басталмас бұрын, құрылғыңыздың, камераңыздың, микрофоныңыздың және интернет байланысыңыздың дұрыс жұмыс істеп тұрғанына көз жеткізіңіз. Қабылдаудың үздіксіз өтуі үшін жабдықты алдын ала тексеруді ұсынамыз.\n\nҚоңырауға қосылу сілтемесі:",
				"ai_consultation_prompt":  "Шағымыңызды сипаттаңыз, мен сізге кеңес беремін.",
				"ai_error":                "ИИ-ға жүгінуде қате:",
				"no_specializations":      "Қолжетімді мамандықтар жоқ.",
				"no_doctors":              "Таңдалған мамандық үшін дәрігерлер жоқ.",
				"no_services":             "Таңдалған дәрігердің қолжетімді қызметтері жоқ.",
				"no_timeslots":            "Қолжетімді уақыт аралықтары жоқ.",
				"slot_taken":              "Кешіріңіз, бұл уақыт алынған.",
				"booking_error":           "Жазылу қатесі:",
				"invalid_command":         "Мәзірден әрекетті таңдаңыз немесе бастау үшін /start енгізіңіз.",
				"invalid_specialization":  "Мамандықты тану мүмкін болмады.",
				"invalid_doctor":          "Дәрігерді тану мүмкін болмады.",
				"invalid_service":         "Қызметті тану мүмкін болмады.",
				"patient_data_error":      "Науқас деректерін алу мүмкін болмады.",
				"slot_data_error":         "Слот деректерін алуда қате.",
				"name_validation_empty":   "Толық атыңызды енгізіңіз",
				"name_validation_short":   "Толық атыңыз тым қысқа",
				"name_validation_long":    "Толық атыңыз тым ұзын",
				"name_validation_invalid": "Толық атында тек әріптер, бос орындар және сызықшалар болуы керек",
				"name_validation_words":   "Кемінде екі сөз енгізіңіз (Тегі мен Аты)",
				"name_validation_capital": "Әр сөз бас әріппен басталуы керек",
				"iin_validation_empty":    "ЖСН енгізіңіз",
				"iin_validation_length":   "ЖСН 12 саннан тұруы керек",
				"iin_validation_digits":   "ЖСН тек сандардан тұруы керек",
				"iin_validation_invalid":  "ЖСН дұрыс емес",
				"choose_date":             "Күнді таңдаңыз:",
				"choose_time_for_date":    "Уақытты таңдаңыз",
				"no_available_dates":      "Қолжетімді күндер жоқ.",
				"no_slots_for_date":       "Таңдалған күнге қолжетімді слоттар жоқ.",
				"end_chat":                "❌ Сөйлесуді аяқтау",
			},
		},
	}
}

func (l *Localization) Get(lang Language, key string) string {
	if texts, ok := l.texts[lang]; ok {
		if text, ok := texts[key]; ok {
			return text
		}
	}
	// Fallback to Russian if key not found
	if texts, ok := l.texts[LangRussian]; ok {
		if text, ok := texts[key]; ok {
			return text
		}
	}
	return key // Return key if nothing found
}
