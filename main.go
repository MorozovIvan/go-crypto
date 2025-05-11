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
	"os"
	"strconv"
	"strings"

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

func handleTelegramAuthCallback(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		log.Printf("Phone number is missing from request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone number is required"})
		return
	}

	log.Printf("Received authentication request for phone: %s", phone)

	// Start authentication process
	err := telegramService.AuthenticateUser(phone)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the phone code hash
	hash := telegramService.GetPhoneCodeHash()
	if hash == "" {
		log.Printf("Failed to get phone code hash")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get phone code hash"})
		return
	}

	log.Printf("Successfully sent verification code to %s", phone)

	// Return the phone code hash along with success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"hash":    hash,
	})
}

func handleTelegramPhone(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"phone": telegramService.GetPhone(),
	})
}

func handleTelegramVerifyCode(c *gin.Context) {
	var data struct {
		Phone    string `json:"phone"`
		Code     string `json:"code"`
		Password string `json:"password,omitempty"`
		Hash     string `json:"hash"`
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

	log.Printf("Received verification request - Phone: %s, Code: %s, Hash: %s, Has Password: %v",
		data.Phone, data.Code, data.Hash, data.Password != "")

	// Validate required fields
	if data.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is required"})
		return
	}
	if data.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification code is required"})
		return
	}
	if data.Hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone code hash is required"})
		return
	}

	// Set the phone, code, and hash
	telegramService.SetPhone(data.Phone)
	telegramService.SetCode(data.Code)
	telegramService.SetHash(data.Hash)

	// First try to verify the code
	err := telegramService.VerifyCode(data.Code)
	if err != nil {
		log.Printf("Code verification result: %v", err)

		// Check if this is a 2FA required error
		if strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED") {
			// 2FA is required
			if data.Password == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "2FA_REQUIRED",
					"message": "2FA password is required",
				})
				return
			}

			// Try to verify 2FA
			err = telegramService.Verify2FA(data.Password)
			if err != nil {
				log.Printf("2FA verification failed: %v", err)
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "2FA_FAILED",
					"message": err.Error(),
				})
				return
			}
		} else {
			log.Printf("Code verification failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "CODE_VERIFICATION_FAILED",
				"message": err.Error(),
			})
			return
		}
	}

	// Get the user ID after successful authentication
	userID, err := telegramService.GetCurrentUserID()
	if err != nil {
		log.Printf("Failed to get user ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "USER_ID_FAILED",
			"message": "Failed to get user ID",
		})
		return
	}

	// Verify that we got a valid user ID
	if userID == 0 {
		log.Printf("Invalid user ID received: %d", userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "Failed to get valid user ID",
		})
		return
	}

	// Return success response with user ID
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user_id": userID,
	})
}

func handleTelegramVerify2FA(c *gin.Context) {
	var data struct {
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		log.Printf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log.Printf("Received 2FA verification request")

	// Try to verify 2FA
	err := telegramService.Verify2FA(data.Password)
	if err != nil {
		log.Printf("2FA verification failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "2FA_FAILED",
			"message": err.Error(),
		})
		return
	}

	// Get the user ID after successful authentication
	userID, err := telegramService.GetCurrentUserID()
	if err != nil {
		log.Printf("Failed to get user ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "USER_ID_FAILED",
			"message": "Failed to get user ID",
		})
		return
	}

	// Verify that we got a valid user ID
	if userID == 0 {
		log.Printf("Invalid user ID received: %d", userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "Failed to get valid user ID",
		})
		return
	}

	// Return success response with user ID
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user_id": userID,
	})
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

func handleGetCurrentUser(c *gin.Context) {
	user, err := telegramService.GetCurrentUser()
	if err != nil {
		if strings.Contains(err.Error(), "session expired") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get current user: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
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
	// Set up logging to file
	logFile, logErr := os.OpenFile("telegram.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if logErr != nil {
		log.Fatalf("Failed to open log file: %v", logErr)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

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
		AllowOrigins:     []string{"http://localhost:5173"},
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
		api.POST("/telegram/auth/verify2fa", handleTelegramVerify2FA)
		api.GET("/telegram/groups", handleGetGroups)
		api.GET("/telegram/phone", handleTelegramPhone)
		api.GET("/telegram/user", handleGetCurrentUser)
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
