package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"go-vue/pkg/config"
	"go-vue/pkg/telegram"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type BalanceResponse struct {
	Balance float64 `json:"balance"`
	Error   string  `json:"error,omitempty"`
}

var telegramService *telegram.TelegramService

func getBalance(walletAddress string) (float64, error) {
	pubKey, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		return 0, err
	}

	rpcClient := rpc.New(config.GlobalConfig.RpcEndpoint)

	balance, err := rpcClient.GetBalance(
		context.Background(),
		pubKey,
		rpc.CommitmentConfirmed,
	)
	if err != nil {
		return 0, err
	}

	lamports := balance.Value
	solBalance := new(big.Float).SetUint64(uint64(lamports))
	solBalanceInSOL := new(big.Float).Quo(solBalance, new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))

	solBalanceFloat, _ := solBalanceInSOL.Float64()
	return solBalanceFloat, nil
}

func handleBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	walletAddress := config.GlobalConfig.WalletAddress
	balance, err := getBalance(walletAddress)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"balance": balance,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTelegramAuthLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authLink := telegramService.GenerateAuthLink()
	response := map[string]string{
		"authLink": authLink,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTelegramAuthCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var authData telegram.TelegramAuthData
	if err := json.NewDecoder(r.Body).Decode(&authData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !telegramService.VerifyAuthData(&authData) {
		http.Error(w, "Invalid auth data", http.StatusUnauthorized)
		return
	}

	// Enable user authentication mode
	telegramService.SetUserAuth(true)

	// Get user's groups
	groups, err := telegramService.GetUserGroups(authData.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user groups: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user":   authData,
		"groups": groups,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTelegramStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session or database
	// For now, we'll use a mock user ID
	userID := int64(123456789)

	groups, err := telegramService.GetUserGroups(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user groups: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"connected": true,
		"groups":    groups,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Telegram service
	telegramService, err := telegram.NewTelegramService()
	if err != nil {
		log.Fatalf("Failed to initialize Telegram service: %v", err)
	}

	// Test bot configuration
	if err := telegramService.TestBot(); err != nil {
		log.Fatalf("Failed to test bot: %v", err)
	}
	log.Printf("Bot configured successfully: @%s", telegramService.GetBotUsername())

	// Initialize Gin router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}))

	// API routes
	api := router.Group("/api")
	{
		// Add balance endpoint
		api.GET("/balance", func(c *gin.Context) {
			balance, err := getBalance(config.GlobalConfig.WalletAddress)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"balance": balance})
		})

		api.GET("/telegram/auth", func(c *gin.Context) {
			// Enable user authentication mode
			telegramService.SetUserAuth(true)
			authLink := telegramService.GenerateAuthLink()
			c.JSON(http.StatusOK, gin.H{"auth_link": authLink})
		})

		api.POST("/telegram/auth/callback", func(c *gin.Context) {
			var authData telegram.TelegramAuthData
			if err := c.ShouldBindJSON(&authData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
				return
			}

			if !telegramService.VerifyAuthData(&authData) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication data"})
				return
			}

			// Enable user authentication mode
			telegramService.SetUserAuth(true)

			// Get user's groups
			groups, err := telegramService.GetUserGroups(authData.ID)
			if err != nil {
				log.Printf("Error getting user groups: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to get user groups: %v", err),
				})
				return
			}

			// Store user session (in a production app, you'd use a proper session store)
			// For now, we'll just return the user data and groups
			c.JSON(http.StatusOK, gin.H{
				"user": gin.H{
					"id":        authData.ID,
					"username":  authData.Username,
					"firstName": authData.FirstName,
					"lastName":  authData.LastName,
				},
				"groups": groups,
			})
		})

		api.GET("/telegram/groups", func(c *gin.Context) {
			// Get user ID from query parameter, but don't require it
			userIDStr := c.Query("user_id")
			var userID int64
			var err error

			if userIDStr != "" {
				userID, err = strconv.ParseInt(userIDStr, 10, 64)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
					return
				}
			} else {
				// Default to bot mode if no user_id provided
				telegramService.SetUserAuth(false)
			}

			groups, err := telegramService.GetUserGroups(userID)
			if err != nil {
				log.Printf("Error getting user groups: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to get user groups: %v", err),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{"groups": groups})
		})

		api.GET("/telegram/test", func(c *gin.Context) {
			if err := telegramService.TestBot(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "Bot is configured correctly"})
		})
	}

	// Start server
	port := config.GlobalConfig.Port
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Invalid port number: %v", err)
	}

	for i := 0; i < 10; i++ {
		log.Printf("Trying to start server on port %d...", portInt+i)
		if err := router.Run(fmt.Sprintf(":%d", portInt+i)); err != nil {
			if i == 9 {
				log.Fatalf("Failed to start server after 10 attempts: %v", err)
			}
			continue
		}
		break
	}
}
