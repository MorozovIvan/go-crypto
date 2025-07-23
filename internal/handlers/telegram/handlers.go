package telegram

import (
	"bytes"
	"go-vue/internal/services"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// TODO: Move actual implementations from main.go

func handleTelegramAuthCallback(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		log.Printf("Phone number is missing from request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone number is required"})
		return
	}

	log.Printf("Received authentication request for phone: %s", phone)

	// Get telegram service from global registry
	telegramService := services.GlobalRegistry.Telegram
	if telegramService == nil {
		log.Printf("Telegram service not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Telegram service not available"})
		return
	}

	// Start authentication process
	err := telegramService.AuthenticateUser(phone)
	if err != nil {
		log.Printf("Authentication failed: %v", err)

		// Check if this is a connection error that requires client reinitialization
		if strings.Contains(err.Error(), "engine was closed") ||
			strings.Contains(err.Error(), "AUTH_KEY_UNREGISTERED") ||
			strings.Contains(err.Error(), "connection dead") ||
			strings.Contains(err.Error(), "client is not initialized") ||
			strings.Contains(err.Error(), "context canceled") {
			log.Printf("Telegram client connection issue detected, attempting to reinitialize...")

			// Clear existing session files to force fresh authentication
			os.Remove("session.json")
			os.Remove("auth_state.json")

			// Reinitialize the Telegram service
			newTelegramService := services.NewTelegramService()
			if newTelegramService == nil {
				log.Printf("Failed to reinitialize Telegram service")
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to reinitialize Telegram connection",
					"details": "Service initialization failed",
				})
				return
			}

			// Update the service in the global registry
			services.GlobalRegistry.Telegram = newTelegramService
			telegramService = newTelegramService

			log.Printf("Telegram service reinitialized, retrying authentication...")

			// Retry authentication with the new client
			err = telegramService.AuthenticateUser(phone)
			if err != nil {
				log.Printf("Authentication failed after reinitializing: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Get the phone code hash
	hash := telegramService.GetPhoneCodeHash(phone)
	if hash == "" {
		log.Printf("Failed to get phone code hash")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get phone code hash"})
		return
	}

	log.Printf("Successfully sent verification code to %s", phone)
	log.Printf("Generated hash for phone %s: %s", phone, hash)

	// Return the phone code hash along with success response
	response := gin.H{
		"success": true,
		"hash":    hash,
	}
	log.Printf("Sending response: %+v", response)
	c.JSON(http.StatusOK, response)
}

func handleTelegramPhone(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		log.Printf("Phone number is missing from request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone number is required"})
		return
	}

	log.Printf("Received phone request for: %s", phone)

	// Get telegram service from global registry
	telegramService := services.GlobalRegistry.Telegram
	if telegramService == nil {
		log.Printf("Telegram service not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Telegram service not available"})
		return
	}

	// Start authentication process
	err := telegramService.AuthenticateUser(phone)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the phone code hash
	hash := telegramService.GetPhoneCodeHash(phone)
	if hash == "" {
		log.Printf("Failed to get phone code hash")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get phone code hash"})
		return
	}

	log.Printf("Successfully sent verification code to %s", phone)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"hash":    hash,
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

	// Get telegram service from global registry
	telegramService := services.GlobalRegistry.Telegram
	if telegramService == nil {
		log.Printf("Telegram service not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Telegram service not available"})
		return
	}

	// First try to verify the code
	err := telegramService.VerifyCode(data.Phone, data.Code, data.Hash)
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
			err = telegramService.Verify2FA(data.Phone, data.Password)
			if err != nil {
				log.Printf("2FA verification failed: %v", err)
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "2FA_FAILED",
					"message": err.Error(),
				})
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "CODE_FAILED",
				"message": err.Error(),
			})
			return
		}
	}

	// Add a small delay to ensure authentication state is established
	time.Sleep(500 * time.Millisecond)

	// Explicitly save the authentication state
	telegramService.SaveAuthState()

	// Get the user ID directly from the service (more reliable than GetStatus)
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

	log.Printf("Authentication successful, auth state saved for user ID: %d", userID)

	// Return success response with user ID
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user_id": userID,
	})
}

func handleTelegramVerify2FA(c *gin.Context) {
	var request struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&request); err != nil {
		log.Printf("Failed to parse 2FA verification request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	log.Printf("Received 2FA verification request")

	// Get telegram service from global registry
	telegramService := services.GlobalRegistry.Telegram
	if telegramService == nil {
		log.Printf("Telegram service not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Telegram service not available"})
		return
	}

	// Verify 2FA password
	err := telegramService.Verify2FA(request.Phone, request.Password)
	if err != nil {
		log.Printf("2FA verification failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Successfully verified 2FA")

	// Add a small delay to ensure authentication state is established
	time.Sleep(500 * time.Millisecond)

	// Explicitly save the authentication state
	telegramService.SaveAuthState()

	// Get the user ID directly from the service (more reliable than GetStatus)
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

	log.Printf("2FA authentication successful, auth state saved for user ID: %d", userID)

	// Return success response with user ID
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user_id": userID,
	})
}

func handleTelegramStatus(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"debug": r,
			})
		}
	}()

	if services.GlobalRegistry == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Service registry not initialized",
		})
		return
	}

	if services.GlobalRegistry.Telegram == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Telegram service not initialized",
		})
		return
	}

	status := services.GlobalRegistry.Telegram.GetStatus()
	c.JSON(http.StatusOK, gin.H{"status": status})
}

func handleTelegramLogout(c *gin.Context) {
	log.Printf("Received logout request")

	// Get telegram service from global registry
	telegramService := services.GlobalRegistry.Telegram
	if telegramService == nil {
		log.Printf("Telegram service not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Telegram service not available"})
		return
	}

	// Logout
	err := telegramService.Logout()
	if err != nil {
		log.Printf("Logout failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Successfully logged out")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully logged out",
	})
}

func handleGetGroups(c *gin.Context) {
	log.Printf("Received get groups request")

	// Get telegram service from global registry
	telegramService := services.GlobalRegistry.Telegram
	if telegramService == nil {
		log.Printf("Telegram service not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Telegram service not available"})
		return
	}

	// Get groups
	groups, err := telegramService.GetGroups()
	if err != nil {
		log.Printf("Failed to get groups: %v", err)

		// Check if this is a connection error that requires client reinitialization
		if strings.Contains(err.Error(), "engine was closed") ||
			strings.Contains(err.Error(), "AUTH_KEY_UNREGISTERED") ||
			strings.Contains(err.Error(), "connection dead") ||
			strings.Contains(err.Error(), "client is not initialized") ||
			strings.Contains(err.Error(), "context canceled") {
			log.Printf("Telegram client connection issue detected, attempting to reinitialize...")

			// Clear existing session files to force fresh authentication
			os.Remove("session.json")
			os.Remove("auth_state.json")

			// Reinitialize the Telegram service
			newTelegramService := services.NewTelegramService()
			if newTelegramService == nil {
				log.Printf("Failed to reinitialize Telegram service")
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to reinitialize Telegram connection",
					"details": "Service initialization failed",
				})
				return
			}

			// Update the service in the global registry
			services.GlobalRegistry.Telegram = newTelegramService
			telegramService = newTelegramService

			log.Printf("Telegram service reinitialized, retrying to get groups...")

			// Retry getting groups with the new client
			groups, err = telegramService.GetGroups()
			if err != nil {
				log.Printf("Failed to get groups after reinitializing: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	log.Printf("Successfully retrieved %d groups", len(groups))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"groups":  groups,
	})
}

func handleGetCurrentUser(c *gin.Context) {
	log.Printf("Received get current user request")

	// Get telegram service from global registry
	telegramService := services.GlobalRegistry.Telegram
	if telegramService == nil {
		log.Printf("Telegram service not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Telegram service not available"})
		return
	}

	// Get current user
	user, err := telegramService.GetCurrentUser()
	if err != nil {
		log.Printf("Failed to get current user: %v", err)

		// Check if this is a connection error that requires client reinitialization
		if strings.Contains(err.Error(), "engine was closed") ||
			strings.Contains(err.Error(), "AUTH_KEY_UNREGISTERED") ||
			strings.Contains(err.Error(), "connection dead") ||
			strings.Contains(err.Error(), "client is not initialized") ||
			strings.Contains(err.Error(), "context canceled") {
			log.Printf("Telegram client connection issue detected, attempting to reinitialize...")

			// Clear existing session files to force fresh authentication
			os.Remove("session.json")
			os.Remove("auth_state.json")

			// Reinitialize the Telegram service
			newTelegramService := services.NewTelegramService()
			if newTelegramService == nil {
				log.Printf("Failed to reinitialize Telegram service")
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to reinitialize Telegram connection",
					"details": "Service initialization failed",
				})
				return
			}

			// Update the service in the global registry
			services.GlobalRegistry.Telegram = newTelegramService
			telegramService = newTelegramService

			log.Printf("Telegram service reinitialized, retrying to get current user...")

			// Retry getting current user with the new client
			user, err = telegramService.GetCurrentUser()
			if err != nil {
				log.Printf("Failed to get current user after reinitializing: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	log.Printf("Successfully retrieved current user")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    user,
	})
}

func handleTelegramSignalChannels(c *gin.Context) {
	log.Printf("Received get signal channels request")

	// Get telegram service from global registry
	telegramService := services.GlobalRegistry.Telegram
	if telegramService == nil {
		log.Printf("Telegram service not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Telegram service not available"})
		return
	}

	// Get query parameters
	period := c.DefaultQuery("period", "24h")
	userIDStr := c.DefaultQuery("user_id", "")

	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id parameter is required"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	// Get signal channels with real processing
	channels, err := telegramService.GetSignalChannelsWithProcessing(userID, period)
	if err != nil {
		log.Printf("Failed to get signal channels: %v", err)

		// Check if this is a connection error that requires client reinitialization
		if strings.Contains(err.Error(), "engine was closed") ||
			strings.Contains(err.Error(), "AUTH_KEY_UNREGISTERED") ||
			strings.Contains(err.Error(), "connection dead") ||
			strings.Contains(err.Error(), "client is not initialized") ||
			strings.Contains(err.Error(), "context canceled") {
			log.Printf("Telegram client connection issue detected, attempting to reinitialize...")

			// Clear existing session files to force fresh authentication
			os.Remove("session.json")
			os.Remove("auth_state.json")

			// Reinitialize the Telegram service
			newTelegramService := services.NewTelegramService()
			if newTelegramService == nil {
				log.Printf("Failed to reinitialize Telegram service")
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to reinitialize Telegram connection",
					"details": "Service initialization failed",
				})
				return
			}

			// Update the service in the global registry
			services.GlobalRegistry.Telegram = newTelegramService
			telegramService = newTelegramService

			log.Printf("Telegram service reinitialized, retrying to get signal channels...")

			// Retry getting signal channels with the new client
			channels, err = telegramService.GetSignalChannelsWithProcessing(userID, period)
			if err != nil {
				log.Printf("Failed to get signal channels after reinitializing: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	log.Printf("Successfully retrieved %d signal channels", len(channels))

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"channels": channels,
	})
}

func handleTokenAnalysis(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing token address"})
		return
	}

	log.Printf("Received token analysis request for address: %s", address)

	// Get telegram service from global registry
	telegramService := services.GlobalRegistry.Telegram
	if telegramService == nil {
		log.Printf("Telegram service not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Telegram service not available"})
		return
	}

	// Analyze token
	analysis, err := telegramService.AnalyzeToken(address)
	if err != nil {
		log.Printf("Failed to analyze token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Successfully analyzed token: %s", address)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"analysis": analysis,
	})
}

func handleBulkTokenAnalysis(c *gin.Context) {
	var request struct {
		GroupID   string   `json:"group_id"`
		Addresses []string `json:"addresses"`
	}

	if err := c.BindJSON(&request); err != nil {
		log.Printf("Failed to parse bulk token analysis request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	log.Printf("Received bulk token analysis request for group: %s with %d addresses",
		request.GroupID, len(request.Addresses))

	// Get telegram service from global registry
	telegramService := services.GlobalRegistry.Telegram
	if telegramService == nil {
		log.Printf("Telegram service not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Telegram service not available"})
		return
	}

	// Analyze tokens in bulk
	analyses, err := telegramService.AnalyzeTokensBulk(request.GroupID, request.Addresses)
	if err != nil {
		log.Printf("Failed to analyze tokens in bulk: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Successfully analyzed %d tokens in bulk", len(analyses))

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"analyses": analyses,
	})
}
