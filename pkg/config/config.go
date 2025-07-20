package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Environment          string
	Port                 int
	RpcEndpoint          string
	RpcWebsocketEndpoint string
	WalletAddress        string
	TelegramAPIID        string
	TelegramAPIHash      string
	DefaultPhoneNumber   string
	PostgresHost         string
	PostgresPort         string
	PostgresUser         string
	PostgresPassword     string
	PostgresDB           string
}

var GlobalConfig Config

func Load() *Config {
	return &Config{
		Environment:          getEnv("ENVIRONMENT", "development"),
		Port:                 getEnvAsInt("PORT", 8080),
		RpcEndpoint:          getEnv("RPC_ENDPOINT", "https://api.mainnet-beta.solana.com"),
		RpcWebsocketEndpoint: getEnv("RPC_WEBSOCKET_ENDPOINT", ""),
		WalletAddress:        getEnv("WALLET_ADDRESS", ""),
		TelegramAPIID:        getEnv("TELEGRAM_API_ID", ""),
		TelegramAPIHash:      getEnv("TELEGRAM_API_HASH", ""),
		DefaultPhoneNumber:   getEnv("DEFAULT_PHONE_NUMBER", "+380500232802"),
		PostgresHost:         getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:         getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:         getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword:     getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:           getEnv("POSTGRES_DB", "go_vue"),
	}
}

func LoadConfig() error {
	GlobalConfig = *Load()

	if GlobalConfig.TelegramAPIID == "" {
		return fmt.Errorf("TELEGRAM_API_ID is required")
	}
	if GlobalConfig.TelegramAPIHash == "" {
		return fmt.Errorf("TELEGRAM_API_HASH is required")
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
