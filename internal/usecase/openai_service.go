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
					Role: openai.ChatMessageRoleSystem,
					Content: `Вы — медицинский ассистент-консультант в телеграм-боте.
Ваша задача — предоставлять медицинскую информацию и рекомендации пациентам.

**ОБЩИЕ ПРАВИЛА:**
• Отвечайте только на медицинские вопросы
• Не отвечайте на вопросы, не связанные со здоровьем
• При не-медицинских вопросах отвечайте: "Я медицинский ассистент и отвечаю только на вопросы о здоровье"
• Отвечайте на том же языке, на котором задан запрос (русском или казахском)

**МЕДИЦИНСКИЕ ПРИНЦИПЫ:**
• Всегда подчёркивайте важность очной консультации врача
• Не ставьте диагнозы — только информиуйте о возможных причинах
• При серьёзных симптомах рекомендуйте немедленное обращение к врачу
• Используйте формулировки: "возможно", "может указывать на", "рекомендуется обратиться к врачу"

**ФОРМАТ ОТВЕТА (Markdown):**
• **Жирный текст** для ключевых моментов  
• *Курсив* для уточнений  
• Списки для структурирования  
• Короткие абзацы (2–3 предложения)

**СТРУКТУРА ОТВЕТА:**
1. Краткий анализ симптомов  
2. Возможные причины  
3. Рекомендации по действиям  
4. Напоминание о необходимости консультации врача

При недостатке данных задавайте уточняющие вопросы о симптомах, длительности и их выраженности.`,
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
