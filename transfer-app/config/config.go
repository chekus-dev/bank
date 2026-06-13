package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	Environment           string
	DatabaseURL           string
	JWTSecret             string
	JWTExpiryHours        int
	RefreshTokenExpiry    int
	PaystackSecretKey     string
	PaystackPublicKey     string
	PaystackWebhookSecret string
	PaystackBaseURL       string
	TermiiAPIKey          string
	TermiiBaseURL         string
	TermiiSenderID        string
	OTPExpiryMins         int
	AppName               string
	SupportEmail          string
	MinTransferAmt        float64
	MaxTransferAmt        float64
	TransferFee           float64
}

var App *Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}
	App = &Config{
		Port:                  getEnv("PORT", "8080"),
		Environment:           getEnv("ENVIRONMENT", "development"),
		DatabaseURL:           mustGetEnv("DATABASE_URL"),
		JWTSecret:             mustGetEnv("JWT_SECRET"),
		JWTExpiryHours:        getEnvInt("JWT_EXPIRY_HOURS", 24),
		RefreshTokenExpiry:    getEnvInt("REFRESH_TOKEN_EXPIRY_DAYS", 30),
		PaystackSecretKey:     mustGetEnv("PAYSTACK_SECRET_KEY"),
		PaystackPublicKey:     mustGetEnv("PAYSTACK_PUBLIC_KEY"),
		PaystackWebhookSecret: mustGetEnv("PAYSTACK_WEBHOOK_SECRET"),
		PaystackBaseURL:       getEnv("PAYSTACK_BASE_URL", "https://api.paystack.co"),
		TermiiAPIKey:          mustGetEnv("TERMII_API_KEY"),
		TermiiBaseURL:         getEnv("TERMII_BASE_URL", "https://api.ng.termii.com/api"),
		TermiiSenderID:        getEnv("TERMII_SENDER_ID", "N-Alert"),
		OTPExpiryMins:         getEnvInt("OTP_EXPIRY_MINS", 5),
		AppName:               getEnv("APP_NAME", "NairaTransfer"),
		SupportEmail:          getEnv("SUPPORT_EMAIL", "support@nairatransfer.com"),
		MinTransferAmt:        getEnvFloat("MIN_TRANSFER_AMOUNT", 100.0),
		MaxTransferAmt:        getEnvFloat("MAX_TRANSFER_AMOUNT", 1000000.0),
		TransferFee:           getEnvFloat("TRANSFER_FEE", 10.0),
	}
	log.Printf("✅ Config loaded — env: %s, port: %s", App.Environment, App.Port)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("❌ Required env var %q is not set", key)
	}
	return val
}

func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvFloat(key string, fallback float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return fallback
	}
	return f
}
