package config

import "os"

type Config struct {
	ListenAddr          string
	DatabaseURL         string
	AgentURL            string
	AgentLLMURL         string
	AgentToken          string
	SummarizationPrompt string
	LLMModel            string
}

func Load() Config {
	return Config{
		ListenAddr:          getEnv("WORKER_LISTEN_ADDR", ":8081"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/dbname?sslmode=disable"),
		AgentURL:            getEnv("AGENT_URL", ""),
		AgentLLMURL:         getEnv("AGENT_LLM_URL", ""),
		AgentToken:          getEnv("AGENT_TOKEN", ""),
		LLMModel:            getEnv("LLM_MODEL", "gpt-4o"),
		SummarizationPrompt: getSummarizationPrompt(),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getSummarizationPrompt() string {
	result := "Ты помощник, который анализирует расшифровки совещаний с диаризацией и создает структурированный отчет " +
		"на языке оригинала встречи (русский или английский). Твоя задача — извлечь максимум пользы из текста и " +
		"сформировать понятный, пригодный для PDF-документа результат. Включи:\n\n" +
		"1. Протокол совещания: участники, темы, принятые решения, задачи с ответственными.\n" +
		"2. Резюме по участникам: что кто говорил и какие действия взял на себя.\n" +
		"3. Резюме по темам: о чём говорили, краткие выводы.\n" +
		"4. Идеи и инициативы: если встреча была творческой — выдели идеи отдельно.\n\n" +
		"Структурируй ответ по разделам, используй списки и заголовки. Не используй текстовые таблицы. Пиши на том языке, на котором проходила встреча."

	return result
}
