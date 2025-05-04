package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                 string
	RpcEndpoint          string
	RpcWebsocketEndpoint string
	WalletAddress        string
	TelegramAPIID        string
	TelegramAPIHash      string
	PostgresHost         string
	PostgresPort         string
	PostgresUser         string
	PostgresPassword     string
	PostgresDB           string
}

var GlobalConfig Config

func LoadConfig() error {
	GlobalConfig = Config{
		Port:                 getEnv("PORT", "8080"),
		RpcEndpoint:          getEnv("RPC_ENDPOINT", "https://api.mainnet-beta.solana.com"),
		RpcWebsocketEndpoint: getEnv("RPC_WEBSOCKET_ENDPOINT", ""),
		WalletAddress:        getEnv("WALLET_ADDRESS", ""),
		TelegramAPIID:        getEnv("TELEGRAM_API_ID", ""),
		TelegramAPIHash:      getEnv("TELEGRAM_API_HASH", ""),
		PostgresHost:         getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:         getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:         getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword:     getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:           getEnv("POSTGRES_DB", "go_vue"),
	}

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
