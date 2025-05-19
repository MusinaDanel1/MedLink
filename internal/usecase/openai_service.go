package usecase

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type Service interface {
	AskChatGPT(prompt string) (string, error)
}

type service struct {
	client *openai.Client
}

func New(apiKey string) Service {
	return &service{
		client: openai.NewClient(apiKey),
	}
}

func (s *service) AskChatGPT(prompt string) (string, error) {
	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "Вы — медицинский ассистент-эксперт. Отвечайте кратко и по существу.\n\nИспользуйте Markdown форматирование:\n1. Для жирного текста: **текст**\n2. Для курсива: *текст*\n3. Для подчеркивания: __текст__\n4. Для зачеркивания: ~~текст~~\n5. Для моноширинного текста: `текст`\n6. Для блоков кода: ```текст```\n\nПример ответа:\n**Основной симптом**: описание\n*Дополнительная информация*\n\n• Пункт 1\n• Пункт 2\n\nЕсли в вопросе недостаточно данных, задайте уточняющий вопрос в конце ответа.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("ошибка OpenAI API: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("пустой ответ от OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}
