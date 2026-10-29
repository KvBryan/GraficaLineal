package config

import (
	"os"
)

type Config struct {
	Port            string
	DeepSeekAPIKey  string
	DeepSeekBaseURL string
	DeepSeekModel   string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		// Fallback to provided key for convenience in local Docker if not set
		apiKey = "sk-3d4ed74400f14a10b4db14b284a07650"
	}

	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	model := os.Getenv("DEEPSEEK_MODEL")
	if model == "" {
		model = "deepseek-chat"
	}

	return &Config{
		Port:            port,
		DeepSeekAPIKey:  apiKey,
		DeepSeekBaseURL: baseURL,
		DeepSeekModel:   model,
	}
}
