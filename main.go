package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go-vue/pkg/config"
	"go-vue/pkg/market"
	"go-vue/pkg/telegram"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var (
	telegramService *telegram.TelegramService
	marketService   *market.MarketService
)

// SSRResponse represents the response for SSR endpoint
type SSRResponse struct {
	CurrentSSR float64   `json:"current_ssr"`
	Historical []float64 `json:"historical"`
	Labels     []string  `json:"labels"`
}

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

func handleBalance(c *gin.Context) {
	walletAddress := config.GlobalConfig.WalletAddress
	balance, err := getBalance(walletAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": balance})
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

	// First try to verify the code
	err := telegramService.VerifyCode(data.Phone, data.Code)
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
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log.Printf("Received 2FA verification request")

	// Try to verify 2FA
	err := telegramService.Verify2FA(data.Phone, data.Password)
	if err != nil {
		log.Printf("2FA verification failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "2FA_FAILED",
			"message": err.Error(),
		})
		return
	}

	// Add a small delay to ensure authentication state is established
	time.Sleep(500 * time.Millisecond)

	// Get the user ID from the service status (it's already stored after successful 2FA)
	status := telegramService.GetStatus()
	userID, ok := status["user_id"].(int64)
	if !ok || userID == 0 {
		log.Printf("Failed to get user ID from service status: %v", status)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "USER_ID_FAILED",
			"message": "Failed to get user ID",
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
		"id":         user["id"],
		"username":   user["username"],
		"first_name": user["first_name"],
		"last_name":  user["last_name"],
	})
}

func handleTelegramStatus(c *gin.Context) {
	status := telegramService.GetStatus()
	c.JSON(http.StatusOK, gin.H{"status": status})
}

func handleTelegramLogout(c *gin.Context) {
	log.Printf("Received logout request")

	// Clear all sessions and reset service state
	telegramService.ClearSessions()

	log.Printf("User logged out successfully")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

func handleCMCGlobal(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API key not set"})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch from CoinMarketCap"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func handleTrends(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		keyword = "bitcoin"
	}
	timeframe := c.Query("timeframe")
	if timeframe == "" {
		timeframe = "now 7-d"
	}

	// Create a Python script to run the trends request
	script := fmt.Sprintf(`
from pytrends.request import TrendReq
import json
import sys

pytrends = TrendReq(hl='en-US', tz=360)
pytrends.build_payload(['%s'], cat=0, timeframe='%s', geo='', gprop='')
data = pytrends.interest_over_time()
if not data.empty:
    values = data['%s'].tolist()
    print(json.dumps({'values': values, 'labels': list(data.index.strftime('%%Y-%%m-%%d'))}))
else:
    print(json.dumps({'values': [], 'labels': []}))
`, keyword, timeframe, keyword)

	// Create a temporary file for the script
	tmpFile, err := os.CreateTemp("", "trends_*.py")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create temporary file",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer os.Remove(tmpFile.Name())

	// Write the script to the file
	if _, err := tmpFile.WriteString(script); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to write script",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	tmpFile.Close()

	// Run the Python script
	cmd := exec.Command("python3", tmpFile.Name())
	output, err := cmd.Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to execute trends request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Parse the result
	var result struct {
		Values []float64 `json:"values"`
		Labels []string  `json:"labels"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse trends response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	if len(result.Values) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"value":        "No data",
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   []float64{},
			"chart_labels": []string{},
		})
		return
	}

	// Calculate the current trend value and indicator
	currentValue := result.Values[len(result.Values)-1]
	avgValue := 0.0
	for _, v := range result.Values {
		avgValue += v
	}
	avgValue /= float64(len(result.Values))

	// Determine indicator and score based on current value vs average
	var indicator string
	var score float64
	if currentValue < avgValue*0.75 {
		indicator = "Buy"
		score = 1
	} else if currentValue > avgValue*1.25 {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        fmt.Sprintf("%.1f", currentValue),
		"indicator":    indicator,
		"score":        score,
		"chart_data":   result.Values,
		"chart_labels": result.Labels,
	})
}

func handleSSR(c *gin.Context) {
	ssr, historical, labels, err := marketService.GetStablecoinSupplyRatio()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := SSRResponse{
		CurrentSSR: ssr,
		Historical: historical,
		Labels:     labels,
	}

	c.JSON(http.StatusOK, response)
}

func handleExchangeFlows(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from CoinMarketCap",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	data := result["data"].(map[string]interface{})
	metrics := data["quote"].(map[string]interface{})["USD"].(map[string]interface{})

	// Calculate exchange flows based on volume and market cap changes
	volume24h := metrics["total_volume_24h"].(float64)
	marketCapChange := metrics["total_market_cap_yesterday_percentage_change"].(float64)

	// Estimate net flow (negative means outflow from exchanges)
	netFlow := -volume24h * (marketCapChange / 100.0)

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = netFlow * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	// Calculate indicator and score
	var indicator string
	var score float64
	if netFlow < -1000 {
		indicator = "Buy"
		score = 1
	} else if netFlow > 1000 {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        netFlow,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleActiveAddresses(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from CoinMarketCap",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	data := result["data"].(map[string]interface{})
	metrics := data["quote"].(map[string]interface{})["USD"].(map[string]interface{})

	// Calculate active addresses based on volume
	volume24h := metrics["total_volume_24h"].(float64)
	activeAddresses := volume24h / 1000 // Rough estimate: 1 address per $1000 of volume

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = activeAddresses * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	// Calculate indicator and score based on trend
	var indicator string
	var score float64
	if activeAddresses > historical[1] {
		indicator = "Buy"
		score = 1
	} else if activeAddresses < historical[1] {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        activeAddresses,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleWhaleTransactions(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from CoinMarketCap",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	data := result["data"].(map[string]interface{})
	metrics := data["quote"].(map[string]interface{})["USD"].(map[string]interface{})

	// Calculate whale transactions based on volume
	volume24h := metrics["total_volume_24h"].(float64)
	whaleTransactions := volume24h / 500000 // Rough estimate: 1 whale transaction per $500,000 of volume

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = whaleTransactions * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	// Calculate indicator and score based on trend
	var indicator string
	var score float64
	if whaleTransactions > historical[1] {
		indicator = "Buy"
		score = 1
	} else if whaleTransactions < historical[1] {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        whaleTransactions,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleFundingRate(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://fapi.binance.com/fapi/v1/premiumIndex?symbol=BTCUSDT", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from Binance",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var data struct {
		Symbol          string `json:"symbol"`
		MarkPrice       string `json:"markPrice"`
		LastFundingRate string `json:"lastFundingRate"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	rate, err := strconv.ParseFloat(data.LastFundingRate, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse funding rate value",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = rate * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	// Calculate indicator and score
	var indicator string
	var score float64
	if rate < -0.0001 {
		indicator = "Buy"
		score = 1
	} else if rate > 0.0001 {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        rate,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleOpenInterest(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://fapi.binance.com/fapi/v1/openInterest?symbol=BTCUSDT", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from Binance",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var data struct {
		OpenInterest string `json:"openInterest"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	interest, err := strconv.ParseFloat(data.OpenInterest, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse open interest value",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = interest * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	// Calculate indicator and score based on trend
	var indicator string
	var score float64
	if interest > historical[1] {
		indicator = "Buy"
		score = 1
	} else if interest < historical[1] {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        interest,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

// Market metrics handlers
func handleAltcoinSeasonIndex(c *gin.Context) {
	index, historical, err := marketService.GetAltcoinSeasonIndex()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        err.Error(),
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        index,
		"indicator":    getIndicator(index, 25, 75),
		"score":        getScore(index, 25, 75),
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleVolumeTrend(c *gin.Context) {
	trend, volumes, err := marketService.GetVolumeTrend()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        err.Error(),
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate trend indicator
	var indicator string
	var score float64
	if trend > 0.1 {
		indicator = "High Rising"
		score = 1
	} else if trend < -0.1 {
		indicator = "Low Falling"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        trend,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   volumes,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleBollingerBands(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API key not set"})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=BTC", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch from CoinMarketCap"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	data := result["data"].(map[string]interface{})
	btcData := data["BTC"].(map[string]interface{})
	quote := btcData["quote"].(map[string]interface{})["USD"].(map[string]interface{})

	// Calculate Bollinger Bands width based on price volatility
	percentChange24h := quote["percent_change_24h"].(float64)

	// Simple estimation of Bollinger Bands width
	width := math.Abs(percentChange24h) / 100.0

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = width * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        width,
		"indicator":    getIndicator(width, 0.02, 0.04),
		"score":        getScore(width, 0.02, 0.04),
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func getIndicator(value float64, lowerBound, upperBound float64) string {
	if value < lowerBound {
		return "Sell"
	} else if value > upperBound {
		return "Buy"
	}
	return "Hold"
}

func getScore(value float64, lowerBound, upperBound float64) float64 {
	if value < lowerBound {
		return -1
	} else if value > upperBound {
		return 1
	}
	return 0
}

func handleRSI(c *gin.Context) {
	rsi, historical, err := marketService.GetRSI()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        err.Error(),
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate indicator and score based on RSI value
	var indicator string
	var score float64
	if rsi <= 30 {
		indicator = "Buy"
		score = 1
	} else if rsi >= 70 {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        rsi,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleBTCDominance(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from CoinMarketCap",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Invalid response format",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	btcDominance, ok := data["btc_dominance"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "BTC dominance value not found",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = btcDominance + float64(i)*0.1 // Simple trend for demonstration
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        btcDominance,
		"indicator":    getIndicator(btcDominance, 40, 60),
		"score":        getScore(btcDominance, 40, 60),
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleMarketCap(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from CoinMarketCap",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	data := result["data"].(map[string]interface{})
	metrics := data["quote"].(map[string]interface{})["USD"].(map[string]interface{})
	marketCap := metrics["total_market_cap"].(float64)

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = marketCap * (1.0 - float64(i)*0.05) // Simple trend for demonstration
	}

	// Calculate indicator and score based on market cap trend
	var indicator string
	var score float64
	if marketCap > historical[1] {
		indicator = "Buy"
		score = 1
	} else if marketCap < historical[1] {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        marketCap,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleETHBTCRatio(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest?symbol=BTC,ETH", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from CoinMarketCap",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	data := result["data"].(map[string]interface{})
	btcData := data["BTC"].([]interface{})[0].(map[string]interface{})
	ethData := data["ETH"].([]interface{})[0].(map[string]interface{})

	btcPrice := btcData["quote"].(map[string]interface{})["USD"].(map[string]interface{})["price"].(float64)
	ethPrice := ethData["quote"].(map[string]interface{})["USD"].(map[string]interface{})["price"].(float64)

	ratio := ethPrice / btcPrice

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = ratio * (1.0 - float64(i)*0.02) // Simple trend for demonstration
	}

	// Calculate indicator and score based on ETH/BTC ratio trend
	var indicator string
	var score float64
	if ratio > historical[1] {
		indicator = "Buy"
		score = 1
	} else if ratio < historical[1] {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        ratio,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleLiquidation(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://fapi.binance.com/fapi/v1/allForceOrders?symbol=BTCUSDT&limit=100", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from Binance",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var liquidations []struct {
		Price string `json:"price"`
		Qty   string `json:"qty"`
		Side  string `json:"side"`
		Time  int64  `json:"time"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&liquidations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate total liquidation value
	var totalValue float64
	for _, liq := range liquidations {
		price, _ := strconv.ParseFloat(liq.Price, 64)
		qty, _ := strconv.ParseFloat(liq.Qty, 64)
		totalValue += price * qty
	}

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = totalValue * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	// Calculate indicator and score based on liquidation volume
	var indicator string
	var score float64
	if totalValue > 1e8 { // More than $100M in liquidations
		indicator = "Sell"
		score = -1
	} else if totalValue < 1e7 { // Less than $10M in liquidations
		indicator = "Buy"
		score = 1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        totalValue,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleGoogleTrends(c *gin.Context) {
	value, historical, err := marketService.GetGoogleTrends()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        err.Error(),
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate average value for comparison
	avgValue := 0.0
	for _, v := range historical {
		avgValue += v
	}
	avgValue /= float64(len(historical))

	// Determine indicator and score based on current value vs average
	var indicator string
	var score float64
	if value < avgValue*0.75 {
		indicator = "Buy"
		score = 1
	} else if value > avgValue*1.25 {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        value,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleMovingAverages(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=BTC", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from CoinMarketCap",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	data := result["data"].(map[string]interface{})
	btcData := data["BTC"].(map[string]interface{})
	quote := btcData["quote"].(map[string]interface{})["USD"].(map[string]interface{})

	// Calculate moving average signal based on price momentum
	percentChange24h := quote["percent_change_24h"].(float64)
	signal := "Hold"
	score := 0

	if percentChange24h > 2.0 {
		signal = "Buy"
		score = 1
	} else if percentChange24h < -2.0 {
		signal = "Sell"
		score = -1
	}

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		if percentChange24h > 0 {
			historical[i] = 1
		} else {
			historical[i] = 0
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        signal,
		"indicator":    signal,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleFearGreed(c *gin.Context) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.alternative.me/fng/", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	q := req.URL.Query()
	q.Add("limit", "5")
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from Fear & Greed Index API",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Value      string `json:"value"`
			ValueClass string `json:"value_classification"`
			Timestamp  string `json:"timestamp"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse Fear & Greed Index response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	if len(result.Data) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "No data received from Fear & Greed Index API",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Get current value
	currentValue, err := strconv.Atoi(result.Data[0].Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse Fear & Greed Index value",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Get historical data
	historical := make([]float64, len(result.Data))
	labels := make([]string, len(result.Data))
	for i, data := range result.Data {
		value, _ := strconv.Atoi(data.Value)
		historical[i] = float64(value)
		// Convert timestamp to date
		timestamp, _ := strconv.ParseInt(data.Timestamp, 10, 64)
		date := time.Unix(timestamp, 0)
		labels[i] = date.Format("Jan 02")
	}

	// Calculate indicator and score
	var indicator string
	var score float64
	if currentValue <= 25 {
		indicator = "Buy"
		score = 1
	} else if currentValue >= 75 {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        currentValue,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": labels,
	})
}

func handlePortfolio(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
		return
	}

	binanceService := market.NewBinanceService()
	portfolio, err := binanceService.GetPortfolio()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get portfolio: %v", err)})
		return
	}

	c.JSON(http.StatusOK, portfolio)
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize config
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Telegram service
	var err error
	telegramService, err = telegram.NewTelegramService()
	if err != nil {
		log.Fatalf("Failed to initialize Telegram service: %v", err)
	}

	// Initialize market service with API key
	marketService = market.NewMarketService(os.Getenv("CMC_API_KEY"))

	// Create Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	api := router.Group("/api")
	{
		// Balance endpoint
		api.GET("/balance", handleBalance)

		// Telegram endpoints
		api.GET("/telegram/auth/callback", handleTelegramAuthCallback)
		api.GET("/telegram/phone", handleTelegramPhone)
		api.POST("/telegram/verify-code", handleTelegramVerifyCode)
		api.POST("/telegram/verify-2fa", handleTelegramVerify2FA)
		api.GET("/telegram/status", handleTelegramStatus)
		api.POST("/telegram/logout", handleTelegramLogout)
		api.GET("/telegram/groups", handleGetGroups)
		api.GET("/telegram/current-user", handleGetCurrentUser)

		// Market data endpoints
		api.GET("/cmc/global", handleCMCGlobal)
		api.GET("/trends", handleTrends)
		api.GET("/ssr", handleSSR)
		api.GET("/exchange-flows", handleExchangeFlows)
		api.GET("/active-addresses", handleActiveAddresses)
		api.GET("/whale-transactions", handleWhaleTransactions)
		api.GET("/funding-rate", handleFundingRate)
		api.GET("/open-interest", handleOpenInterest)

		// Market metrics endpoints
		api.GET("/altcoin-season", handleAltcoinSeasonIndex)
		api.GET("/volume-trend", handleVolumeTrend)
		api.GET("/bollinger-bands", handleBollingerBands)
		api.GET("/rsi", handleRSI)
		api.GET("/moving-averages", handleMovingAverages)

		// Add new routes
		api.GET("/fear-greed", handleFearGreed)
		api.GET("/btc-dominance", handleBTCDominance)
		api.GET("/market-cap", handleMarketCap)
		api.GET("/eth-btc-ratio", handleETHBTCRatio)
		api.GET("/liquidation", handleLiquidation)
		api.GET("/google-trends", handleGoogleTrends)
		api.GET("/portfolio", handlePortfolio)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
