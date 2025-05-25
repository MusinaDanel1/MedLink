package telegram

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// ValidateFullName проверяет корректность ФИО
func ValidateFullName(name string, lang Language) (bool, string) {
	name = strings.TrimSpace(name)

	// Проверка на пустоту
	if name == "" {
		if lang == LangKazakh {
			return false, "Толық атыңызды енгізіңіз"
		}
		return false, "Введите ваше ФИО"
	}

	// Проверка минимальной длины
	if len(name) < 3 {
		if lang == LangKazakh {
			return false, "Толық атыңыз тым қысқа"
		}
		return false, "ФИО слишком короткое"
	}

	// Проверка максимальной длины
	if len(name) > 100 {
		if lang == LangKazakh {
			return false, "Толық атыңыз тым ұзын"
		}
		return false, "ФИО слишком длинное"
	}

	// Проверка на наличие только букв, пробелов и дефисов
	validNameRegex := regexp.MustCompile(`^[а-яёА-ЯЁәіңғүұқөһӘІҢҒҮҰҚӨҺa-zA-Z\s\-]+$`)
	if !validNameRegex.MatchString(name) {
		if lang == LangKazakh {
			return false, "Толық атында тек әріптер, бос орындар және сызықшалар болуы керек"
		}
		return false, "ФИО должно содержать только буквы, пробелы и дефисы"
	}

	// Проверка на минимальное количество слов (должно быть хотя бы 2 слова)
	words := strings.Fields(name)
	if len(words) < 2 {
		if lang == LangKazakh {
			return false, "Кемінде екі сөз енгізіңіз (Тегі мен Аты)"
		}
		return false, "Введите минимум два слова (Фамилия и Имя)"
	}

	// Проверка, что каждое слово начинается с заглавной буквы
	for _, word := range words {
		if len(word) == 0 {
			continue
		}
		firstRune := []rune(word)[0]
		if !unicode.IsUpper(firstRune) {
			if lang == LangKazakh {
				return false, "Әр сөз бас әріппен басталуы керек"
			}
			return false, "Каждое слово должно начинаться с заглавной буквы"
		}
	}

	return true, ""
}

// ValidateIIN проверяет корректность ИИН
func ValidateIIN(iin string, lang Language) (bool, string) {
	iin = strings.TrimSpace(iin)

	// Проверка на пустоту
	if iin == "" {
		if lang == LangKazakh {
			return false, "ЖСН енгізіңіз"
		}
		return false, "Введите ИИН"
	}

	// Проверка длины (ИИН должен содержать 12 цифр)
	if len(iin) != 12 {
		if lang == LangKazakh {
			return false, "ЖСН 12 саннан тұруы керек"
		}
		return false, "ИИН должен содержать 12 цифр"
	}

	// Проверка, что все символы - цифры
	if _, err := strconv.Atoi(iin); err != nil {
		if lang == LangKazakh {
			return false, "ЖСН тек сандардан тұруы керек"
		}
		return false, "ИИН должен содержать только цифры"
	}

	// Проверка контрольной суммы ИИН
	if !isValidIINChecksum(iin) {
		if lang == LangKazakh {
			return false, "ЖСН дұрыс емес"
		}
		return false, "Некорректный ИИН"
	}

	return true, ""
}

// isValidIINChecksum проверяет контрольную сумму ИИН по алгоритму РК
func isValidIINChecksum(iin string) bool {
	if len(iin) != 12 {
		return false
	}

	// Весовые коэффициенты для первых 11 цифр
	weights := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}

	// Вычисляем контрольную сумму
	sum := 0
	for i := 0; i < 11; i++ {
		digit, err := strconv.Atoi(string(iin[i]))
		if err != nil {
			return false
		}
		sum += digit * weights[i]
	}

	// Получаем остаток от деления на 11
	remainder := sum % 11

	// Получаем последнюю цифру ИИН
	lastDigit, err := strconv.Atoi(string(iin[11]))
	if err != nil {
		return false
	}

	// Если остаток меньше 10, то он должен равняться последней цифре
	if remainder < 10 {
		return remainder == lastDigit
	}

	// Если остаток равен 10, то используем дополнительные весовые коэффициенты
	if remainder == 10 {
		weights2 := []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 1, 2}
		sum2 := 0
		for i := 0; i < 11; i++ {
			digit, err := strconv.Atoi(string(iin[i]))
			if err != nil {
				return false
			}
			sum2 += digit * weights2[i]
		}
		remainder2 := sum2 % 11
		if remainder2 < 10 {
			return remainder2 == lastDigit
		}
	}

	return false
}

// FormatFullName приводит ФИО к правильному формату
func FormatFullName(name string) string {
	name = strings.TrimSpace(name)
	words := strings.Fields(name)

	var formattedWords []string
	for _, word := range words {
		if len(word) > 0 {
			// Приводим к формату: первая буква заглавная, остальные строчные
			runes := []rune(strings.ToLower(word))
			runes[0] = unicode.ToUpper(runes[0])
			formattedWords = append(formattedWords, string(runes))
		}
	}

	return strings.Join(formattedWords, " ")
}
