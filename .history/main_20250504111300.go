package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	// This function is no longer needed
}

func handleTelegramAuthCallback(c *gin.Context) {
	// This function is no longer needed
}

func handleTelegramVerifyCode(c *gin.Context) {
	var data struct {
		Phone    string `json:"phone"`
		Code     string `json:"code"`
		Password string `json:"password,omitempty"`
	}

	// Log the raw request body
	body, _ := c.GetRawData()
	log.Printf("Raw request body: %s", string(body))
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if err := c.ShouldBindJSON(&data); err != nil {
		log.Printf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log.Printf("Received verification request - Phone: %s, Code: %s, Has Password: %v",
		data.Phone, data.Code, data.Password != "")

	// Validate required fields
	if data.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is required"})
		return
	}
	if data.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification code is required"})
		return
	}

	// Set the phone and code
	telegramService.SetPhone(data.Phone)
	telegramService.SetCode(data.Code)

	// If password is provided, set it
	if data.Password != "" {
		telegramService.SetPassword(data.Password)
	}

	// Verify the code
	err := telegramService.VerifyCode(data.Code)
	if err != nil {
		log.Printf("Failed to verify code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to verify code: %v", err)})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func handleGetGroups(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	groups, err := telegramService.GetUserGroups(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get user groups: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected": true,
		"groups":    groups,
	})
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
	var err error
	telegramService, err = telegram.NewTelegramService()
	if err != nil {
		log.Fatalf("Failed to initialize Telegram service: %v", err)
	}

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
			walletAddress := config.GlobalConfig.WalletAddress
			balance, err := getBalance(walletAddress)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"balance": balance})
		})

		// Telegram endpoints
		api.GET("/telegram/auth/callback", handleTelegramAuthCallback)
		api.POST("/telegram/auth/verify", handleTelegramVerifyCode)
		api.GET("/telegram/groups", handleGetGroups)
	}

	// Start server
	port := config.GlobalConfig.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
