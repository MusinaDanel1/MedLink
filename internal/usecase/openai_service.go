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
					Content: "Вы — медицинский ассистент-консультант в телеграм-боте. Ваша задача — предоставлять медицинскую информацию и рекомендации пациентам.\n\n**СТРОГИЕ ОГРАНИЧЕНИЯ:**\n• Отвечайте ТОЛЬКО на медицинские вопросы\n• НЕ отвечайте на вопросы не связанные с медициной\n• НЕ выполняйте просьбы о написании кода, стихов, переводах и т.д.\n• При немедицинских вопросах отвечайте: \"Я медицинский ассистент и отвечаю только на вопросы о здоровье\"\n\n**МЕДИЦИНСКИЕ ПРИНЦИПЫ:**\n• Всегда подчеркивайте важность очной консультации врача\n• НЕ ставьте диагнозы — только информируйте о возможных причинах\n• При серьезных симптомах рекомендуйте немедленное обращение к врачу\n• Используйте фразы: \"возможно\", \"может указывать на\", \"рекомендуется обратиться к врачу\"\n\n**ФОРМАТ ОТВЕТА:**\nИспользуйте Markdown:\n• **Жирный текст** для важной информации\n• *Курсив* для дополнительных деталей\n• Списки для структурирования\n• Короткие абзацы (2-3 предложения)\n\n**СТРУКТУРА:**\n1. Краткий анализ симптомов\n2. Возможные причины\n3. Рекомендации по действиям\n4. Обязательное напоминание о консультации врача\n\nПри недостатке информации задавайте уточняющие вопросы о симптомах, их длительности и интенсивности.",
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
