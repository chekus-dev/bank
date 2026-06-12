package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	JWTExpiry       string
	PaystackSecret  string
	PaystackBaseURL string
	WebhookSecret   string
	OTPApiKey       string
	OTPSenderID     string
	AppEnv          string
}

func Load() (*Config, error) {
	// only load .env in development
	if os.Getenv("APP_ENV") != "production" {
		_ = godotenv.Load()
	}

	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		JWTExpiry:       getEnv("JWT_EXPIRY", "24h"),
		PaystackSecret:  os.Getenv("PAYSTACK_SECRET"),
		PaystackBaseURL: getEnv("PAYSTACK_BASE_URL", "https://api.paystack.co"),
		WebhookSecret:   os.Getenv("PAYSTACK_WEBHOOK_SECRET"),
		OTPApiKey:       os.Getenv("OTP_API_KEY"),
		OTPSenderID:     getEnv("OTP_SENDER_ID", "Transfer"),
		AppEnv:          getEnv("APP_ENV", "development"),
	}

	return cfg, cfg.validate()
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required")
	}
	if c.PaystackSecret == "" {
		return errors.New("PAYSTACK_SECRET is required")
	}
	if c.WebhookSecret == "" {
		return errors.New("PAYSTACK_WEBHOOK_SECRET is required")
	}
	if c.OTPApiKey == "" {
		return errors.New("OTP_API_KEY is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}