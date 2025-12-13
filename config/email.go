package config

import (
	"os"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

func NewEmailConfig() *EmailConfig {
	return &EmailConfig{
		SMTPHost:     "smtp.gmail.com", // Hardcoded for Gmail
		SMTPPort:     "587",            // Hardcoded for Gmail (TLS)
		SMTPUsername: os.Getenv("SMTP_ACC"),
		SMTPPassword: os.Getenv("SMTP_ACC_PASSWORD"),
		FromEmail:    getEnvWithDefault("SMTP_ACC", ""),
		FromName:     getEnvWithDefault("FROM_NAME", "Franchiso"),
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
