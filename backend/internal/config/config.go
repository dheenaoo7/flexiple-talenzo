package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GeminiAPIKey string
	GeminiModel  string
	Port         string
	DBPath       string
	ProfilesPath string
}

func Load() (*Config, error) {
	// Overload (not Load): the project's own .env should take priority over
	// any same-named variable that happens to already be set in the shell.
	_ = godotenv.Overload()

	cfg := &Config{
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		GeminiModel:  getEnvDefault("GEMINI_MODEL", "gemini-2.5-flash"),
		Port:         getEnvDefault("PORT", "8080"),
		DBPath:       getEnvDefault("DB_PATH", "./talent_search.db"),
		ProfilesPath: getEnvDefault("PROFILES_PATH", "./data/profiles.json"),
	}

	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set (copy .env.example to .env and fill it in)")
	}

	return cfg, nil
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
