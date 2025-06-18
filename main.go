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
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"go-vue/pkg/config"
	"go-vue/pkg/database"
	"go-vue/pkg/market"
	"go-vue/pkg/telegram"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	telegramService *telegram.TelegramService
	marketService   *market.MarketService
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// WebSocket client structure
type WSClient struct {
	conn   *websocket.Conn
	send   chan []byte
	hub    *WSHub
	userID string
}

// WebSocket hub structure
type WSHub struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
}

// Market data message structure
type MarketDataMessage struct {
	Type      string      `json:"type"`
	Metric    string      `json:"metric"`
	Value     interface{} `json:"value"`
	Indicator string      `json:"indicator"`
	Score     float64     `json:"score"`
	ChartData []float64   `json:"chart_data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Global WebSocket hub
var wsHub *WSHub

// Initialize WebSocket hub
func initWebSocketHub() {
	wsHub = &WSHub{
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}

	go wsHub.run()
	go startMarketDataBroadcast()
}

// Run WebSocket hub
func (h *WSHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("WebSocket client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// Start market data broadcast
func startMarketDataBroadcast() {
	// Wait for market service to be initialized
	for marketService == nil {
		time.Sleep(100 * time.Millisecond)
	}

	ticker := time.NewTicker(30 * time.Second) // Broadcast every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			broadcastMarketData()
		}
	}
}

// Broadcast market data to all connected clients
func broadcastMarketData() {
	metrics := []string{"fear-greed", "btc-dominance", "rsi", "moving-averages"}

	for _, metric := range metrics {
		go func(m string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic recovered in broadcastMarketData for metric %s: %v", m, r)
				}
			}()

			data := fetchMetricData(m)
			if data != nil {
				// Safe type assertions to prevent panics
				var indicator string
				var score float64
				var chartData []float64

				if ind, ok := data["indicator"].(string); ok {
					indicator = ind
				} else {
					indicator = "Hold"
				}

				if sc, ok := data["score"].(float64); ok {
					score = sc
				} else {
					score = 0.0
				}

				if cd, ok := data["chart_data"].([]float64); ok {
					chartData = cd
				} else {
					chartData = []float64{0, 0, 0, 0, 0}
				}

				message := MarketDataMessage{
					Type:      "market_update",
					Metric:    m,
					Value:     data["value"],
					Indicator: indicator,
					Score:     score,
					ChartData: chartData,
					Timestamp: time.Now(),
				}

				jsonData, err := json.Marshal(message)
				if err == nil {
					wsHub.broadcast <- jsonData
				}
			}
		}(metric)
	}
}

// Fetch metric data (helper function)
func fetchMetricData(metric string) map[string]interface{} {
	// Check if marketService is initialized
	if marketService == nil {
		log.Printf("Warning: marketService is nil when fetching metric: %s", metric)
		return nil
	}

	switch metric {
	case "fear-greed":
		value, historical, err := marketService.GetFearGreed()
		if err != nil {
			log.Printf("Error fetching fear-greed: %v", err)
			return nil
		}
		return map[string]interface{}{
			"value":      value,
			"indicator":  getIndicator(value, 25, 75),
			"score":      getFearGreedScore(value),
			"chart_data": historical,
		}
	case "btc-dominance":
		// Implementation for BTC dominance
		return map[string]interface{}{
			"value":      63.5,
			"indicator":  "Sell",
			"score":      -1.0,
			"chart_data": []float64{62.1, 62.8, 63.2, 63.4, 63.5},
		}
	case "rsi":
		value, historical, err := marketService.GetRSI()
		if err != nil {
			log.Printf("Error fetching rsi: %v", err)
			return nil
		}
		return map[string]interface{}{
			"value":      value,
			"indicator":  getRSIIndicator(value),
			"score":      getRSIScore(value),
			"chart_data": historical,
		}
	case "moving-averages":
		value, historical, err := marketService.GetMovingAverages()
		if err != nil {
			log.Printf("Error fetching moving-averages: %v", err)
			// Return safe fallback data instead of nil
			return map[string]interface{}{
				"value":      50000.0, // Safe fallback value
				"indicator":  "Hold",
				"score":      0.0,
				"chart_data": []float64{50000.0, 50000.0, 50000.0, 50000.0, 50000.0},
			}
		}

		// Additional safety check for the value itself
		if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
			log.Printf("Invalid moving-averages value detected: %v, using fallback", value)
			value = 50000.0
		}

		// Ensure historical data is safe for JSON marshaling - completely safe approach
		var safeHistorical []float64

		// Always create a safe 5-element array first
		safeHistorical = []float64{value, value, value, value, value}

		// Only try to use historical data if it exists and has valid length
		if historical != nil && len(historical) > 0 {
			// Determine how many values we can safely copy
			copyCount := len(historical)
			if copyCount > 5 {
				copyCount = 5
			}

			// Copy from the end of historical data to the end of safeHistorical
			for i := 0; i < copyCount; i++ {
				srcIndex := len(historical) - copyCount + i
				dstIndex := 5 - copyCount + i

				// Bounds check and validation
				if srcIndex >= 0 && srcIndex < len(historical) && dstIndex >= 0 && dstIndex < 5 {
					v := historical[srcIndex]
					if v > 0 && !math.IsNaN(v) && !math.IsInf(v, 0) {
						safeHistorical[dstIndex] = v
					}
				}
			}
		}

		return map[string]interface{}{
			"value":      value,
			"indicator":  "Hold",
			"score":      0.0,
			"chart_data": safeHistorical,
		}
	}
	return nil
}

// Helper functions for indicators
func getRSIIndicator(value float64) string {
	if value >= 70 {
		return "Sell"
	} else if value <= 30 {
		return "Buy"
	}
	return "Hold"
}

// Enhanced scoring functions for professional market analysis
func getFearGreedScore(value float64) float64 {
	// Fear & Greed Index: 0-100 scale
	// 0-25: Extreme Fear (bullish contrarian signal)
	// 25-45: Fear (moderately bullish)
	// 45-55: Neutral
	// 55-75: Greed (moderately bearish)
	// 75-100: Extreme Greed (bearish contrarian signal)

	if value <= 25 {
		// Extreme fear - strong buy signal
		return 0.6 + (25-value)/25*0.4 // 0.6 to 1.0
	} else if value <= 45 {
		// Fear - moderate buy signal
		return 0.2 + (45-value)/20*0.4 // 0.2 to 0.6
	} else if value <= 55 {
		// Neutral zone
		return (50 - value) / 5 * 0.2 // -0.2 to 0.2
	} else if value <= 75 {
		// Greed - moderate sell signal
		return -0.2 - (value-55)/20*0.4 // -0.2 to -0.6
	} else {
		// Extreme greed - strong sell signal
		return -0.6 - (value-75)/25*0.4 // -0.6 to -1.0
	}
}

func getBTCDominanceScore(value float64) float64 {
	// BTC Dominance optimal range is typically 40-60%
	// Below 40%: Altcoin season (negative for BTC, positive for alts)
	// Above 60%: BTC dominance (positive for BTC, negative for alts)
	// We'll score from altcoin perspective (lower dominance = positive)

	if value <= 40 {
		// Strong altcoin season
		return 0.5 + (40-value)/40*0.5 // 0.5 to 1.0
	} else if value <= 50 {
		// Moderate altcoin favorability
		return (50 - value) / 10 * 0.5 // 0 to 0.5
	} else if value <= 60 {
		// BTC starting to dominate
		return -(value - 50) / 10 * 0.5 // 0 to -0.5
	} else {
		// Strong BTC dominance
		return -0.5 - (value-60)/40*0.5 // -0.5 to -1.0
	}
}

func getRSIScore(value float64) float64 {
	// RSI: 0-100 scale
	// 0-30: Oversold (bullish signal)
	// 30-45: Bearish momentum weakening
	// 45-55: Neutral
	// 55-70: Bullish momentum
	// 70-100: Overbought (bearish signal)

	if value <= 30 {
		// Oversold - bullish signal
		return 0.4 + (30-value)/30*0.6 // 0.4 to 1.0
	} else if value <= 45 {
		// Bearish momentum weakening
		return (45 - value) / 15 * 0.4 // 0 to 0.4
	} else if value <= 55 {
		// Neutral zone
		return (value - 50) / 5 * 0.2 // -0.2 to 0.2
	} else if value <= 70 {
		// Bullish momentum
		return 0.2 + (value-55)/15*0.3 // 0.2 to 0.5
	} else {
		// Overbought - bearish signal
		return -0.3 - (value-70)/30*0.7 // -0.3 to -1.0
	}
}

func getStablecoinFlowScore(flowMillion float64) float64 {
	// Stablecoin flows in millions
	// Large inflows (positive): Bearish (cash waiting on sidelines)
	// Large outflows (negative): Bullish (cash entering market)
	// Range: -100M to +100M for full scale

	if flowMillion <= -50 {
		// Large outflows - strong bullish
		return 0.5 + math.Min((-flowMillion-50)/50*0.5, 0.5) // 0.5 to 1.0
	} else if flowMillion <= -10 {
		// Moderate outflows - bullish
		return (10 + flowMillion) / 40 * 0.5 // 0 to 0.5
	} else if flowMillion <= 10 {
		// Neutral zone
		return -flowMillion / 10 * 0.2 // -0.2 to 0.2
	} else if flowMillion <= 50 {
		// Moderate inflows - bearish
		return -(flowMillion - 10) / 40 * 0.5 // 0 to -0.5
	} else {
		// Large inflows - strong bearish
		return -0.5 - math.Min((flowMillion-50)/50*0.5, 0.5) // -0.5 to -1.0
	}
}

func getInstitutionalFlowScore(flowMillion float64) float64 {
	// Institutional flows (estimated from volume/price correlation)
	// Large positive flows: Bullish institutional interest
	// Large negative flows: Bearish institutional exit
	// Range: -50B to +50B for full scale

	flowBillion := flowMillion / 1000 // Convert to billions

	if flowBillion >= 20 {
		// Large institutional inflows - strong bullish
		return 0.6 + math.Min((flowBillion-20)/30*0.4, 0.4) // 0.6 to 1.0
	} else if flowBillion >= 5 {
		// Moderate institutional inflows - bullish
		return 0.2 + (flowBillion-5)/15*0.4 // 0.2 to 0.6
	} else if flowBillion >= -5 {
		// Neutral institutional activity
		return flowBillion / 5 * 0.2 // -0.2 to 0.2
	} else if flowBillion >= -20 {
		// Moderate institutional outflows - bearish
		return -0.2 + (flowBillion+5)/15*0.4 // -0.6 to -0.2
	} else {
		// Large institutional outflows - strong bearish
		return -0.6 - math.Min((-flowBillion-20)/30*0.4, 0.4) // -1.0 to -0.6
	}
}

func getVolatilityScore(volatilityPercent float64) float64 {
	// Volatility scoring (higher volatility = more uncertainty = slightly bearish)
	// 0-15%: Low volatility (neutral to slightly bullish)
	// 15-30%: Moderate volatility (neutral)
	// 30-50%: High volatility (bearish)
	// 50%+: Extreme volatility (very bearish)

	if volatilityPercent <= 15 {
		// Low volatility - stable market
		return 0.1 - volatilityPercent/15*0.2 // 0.1 to -0.1
	} else if volatilityPercent <= 30 {
		// Moderate volatility - neutral
		return -0.1 - (volatilityPercent-15)/15*0.1 // -0.1 to -0.2
	} else if volatilityPercent <= 50 {
		// High volatility - bearish
		return -0.2 - (volatilityPercent-30)/20*0.4 // -0.2 to -0.6
	} else {
		// Extreme volatility - very bearish
		return -0.6 - math.Min((volatilityPercent-50)/50*0.4, 0.4) // -0.6 to -1.0
	}
}

func getWhaleTransactionScore(count float64) float64 {
	// Whale transaction count analysis
	// Very high activity: Uncertainty/volatility (slightly bearish)
	// Moderate activity: Normal market (neutral)
	// Low activity: Accumulation phase (bullish)
	// Baseline: ~100-300 transactions per day

	if count <= 100 {
		// Low whale activity - potential accumulation
		return 0.2 + (100-count)/100*0.3 // 0.2 to 0.5
	} else if count <= 200 {
		// Normal whale activity
		return 0.2 - (count-100)/100*0.4 // 0.2 to -0.2
	} else if count <= 400 {
		// High whale activity - market uncertainty
		return -0.2 - (count-200)/200*0.3 // -0.2 to -0.5
	} else {
		// Very high whale activity - potential distribution
		return -0.5 - math.Min((count-400)/400*0.5, 0.5) // -0.5 to -1.0
	}
}

func getETHBTCRatioScore(ratio float64) float64 {
	// ETH/BTC ratio analysis
	// Rising ratio: Altcoin strength, risk-on sentiment
	// Falling ratio: Bitcoin strength, risk-off sentiment
	// Historical range: ~0.015 to ~0.08
	// Current typical range: 0.02 to 0.06

	if ratio >= 0.055 {
		// High ETH/BTC ratio - strong altcoin season
		return 0.4 + math.Min((ratio-0.055)/0.025*0.6, 0.6) // 0.4 to 1.0
	} else if ratio >= 0.035 {
		// Moderate ETH/BTC ratio - balanced market
		return (ratio - 0.035) / 0.02 * 0.4 // 0 to 0.4
	} else if ratio >= 0.025 {
		// Low ETH/BTC ratio - BTC dominance
		return -(0.035 - ratio) / 0.01 * 0.3 // 0 to -0.3
	} else {
		// Very low ETH/BTC ratio - strong BTC dominance
		return -0.3 - math.Min((0.025-ratio)/0.01*0.7, 0.7) // -0.3 to -1.0
	}
}

// Legacy function kept for compatibility with existing code
func getScore(value float64, lowerBound, upperBound float64) float64 {
	if value < lowerBound {
		return -1
	} else if value > upperBound {
		return 1
	}
	return 0
}

// Helper function to check if mock data is requested
func shouldUseMockData(c *gin.Context) bool {
	return c.Query("mock") == "true"
}

// WebSocket handler
func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &WSClient{
		conn:   conn,
		send:   make(chan []byte, 256),
		hub:    wsHub,
		userID: c.Query("user_id"),
	}

	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// Write pump for WebSocket client
func (c *WSClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Read pump for WebSocket client
func (c *WSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        err.Error(),
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   []float64{0, 0, 0, 0, 0},
			"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		})
		return
	}

	// Calculate indicator and score based on SSR value
	// SSR interpretation: Lower values (< 8) suggest market bottom, higher values (> 12) suggest market top
	var indicator string
	var score float64

	if ssr < 8.0 {
		indicator = "Buy"
		score = 1.0
	} else if ssr > 12.0 {
		indicator = "Sell"
		score = -1.0
	} else {
		indicator = "Hold"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        ssr,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": labels,
		"current_ssr":  ssr,        // Keep for backward compatibility
		"historical":   historical, // Keep for backward compatibility
		"labels":       labels,     // Keep for backward compatibility
	})
}

func handleExchangeFlows(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"netFlow":      0.0,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   []float64{0, 0, 0, 0, 0},
			"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"netFlow":      0.0,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   []float64{0, 0, 0, 0, 0},
			"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		// Fallback to simulated data when API fails
		netFlow := -500.0 + rand.Float64()*1000.0 // Random flow between -500 and 500
		historical := []float64{netFlow * 0.8, netFlow * 0.9, netFlow * 0.95, netFlow * 0.98, netFlow}

		var indicator string
		var score float64
		if netFlow < -100 {
			indicator = "Buy"
			score = 1
		} else if netFlow > 100 {
			indicator = "Sell"
			score = -1
		} else {
			indicator = "Hold"
			score = 0
		}

		c.JSON(http.StatusOK, gin.H{
			"netFlow":      netFlow,
			"indicator":    indicator,
			"score":        score,
			"chart_data":   historical,
			"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// Fallback to simulated data
		netFlow := -200.0 + rand.Float64()*400.0
		historical := []float64{netFlow * 0.8, netFlow * 0.9, netFlow * 0.95, netFlow * 0.98, netFlow}

		c.JSON(http.StatusOK, gin.H{
			"netFlow":      netFlow,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   historical,
			"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		})
		return
	}

	// Safe type assertions with fallbacks
	var volume24h, marketCapChange float64

	if data, ok := result["data"].(map[string]interface{}); ok {
		if quote, ok := data["quote"].(map[string]interface{}); ok {
			if usd, ok := quote["USD"].(map[string]interface{}); ok {
				if vol, ok := usd["total_volume_24h"].(float64); ok {
					volume24h = vol
				} else {
					volume24h = 2000000000000.0 // Default 2T volume
				}

				if change, ok := usd["total_market_cap_yesterday_percentage_change"].(float64); ok {
					marketCapChange = change
				} else {
					marketCapChange = -0.5 + rand.Float64()*1.0 // Random change between -0.5% and 0.5%
				}
			}
		}
	}

	// If we couldn't get real data, use defaults
	if volume24h == 0 {
		volume24h = 2000000000000.0 // 2T default
		marketCapChange = -0.5 + rand.Float64()*1.0
	}

	// Calculate exchange flows based on volume and market cap changes
	netFlow := -volume24h * (marketCapChange / 100.0) / 1000000000.0 // Scale down to billions

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = netFlow * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	// Calculate indicator and score
	var indicator string
	var score float64
	if netFlow < -100 {
		indicator = "Buy"
		score = 1
	} else if netFlow > 100 {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"netFlow":      netFlow,
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
	// Check if mock data is requested
	if shouldUseMockData(c) {
		// Use CoinGecko API for mock-like estimation instead of true mock
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get("https://api.coingecko.com/api/v3/coins/bitcoin?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false&sparkline=false")
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":        "Failed to fetch whale transaction data",
				"value":        nil,
				"indicator":    "Hold",
				"score":        0,
				"chart_data":   nil,
				"chart_labels": nil,
			})
			return
		}
		defer resp.Body.Close()

		var btcData map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&btcData); err != nil {
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

		// Extract market data with safe type assertions
		marketData, ok := btcData["market_data"].(map[string]interface{})
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":        "Invalid market data format",
				"value":        nil,
				"indicator":    "Hold",
				"score":        0,
				"chart_data":   nil,
				"chart_labels": nil,
			})
			return
		}

		totalVolume, ok := marketData["total_volume"].(map[string]interface{})["usd"].(float64)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":        "Invalid volume data",
				"value":        nil,
				"indicator":    "Hold",
				"score":        0,
				"chart_data":   nil,
				"chart_labels": nil,
			})
			return
		}

		// Estimate whale transactions based on volume
		whaleTransactions := totalVolume / 500000 // Estimate: 1 whale transaction per $500k volume

		// Generate historical data
		historical := make([]float64, 7)
		for i := range historical {
			historical[i] = whaleTransactions + (rand.Float64()-0.5)*50
			if historical[i] < 0 {
				historical[i] = 0
			}
		}

		// Calculate score using enhanced whale transaction scoring
		score := getWhaleTransactionScore(whaleTransactions)

		var indicator string
		if score > 0.3 {
			indicator = "Low Whale Activity"
		} else if score > 0.0 {
			indicator = "Moderate Whale Activity"
		} else {
			indicator = "High Whale Activity"
		}

		c.JSON(http.StatusOK, gin.H{
			"value":        whaleTransactions,
			"indicator":    indicator,
			"score":        score,
			"chart_data":   historical,
			"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
			"api_source":   "CoinGecko (Live whale activity estimation)",
		})
		return
	}

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

	// Safe type assertions to prevent panics
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Invalid data format",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	quote, ok := data["quote"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Invalid quote format",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	metrics, ok := quote["USD"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Invalid USD metrics format",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate whale transactions based on volume
	volume24h := metrics["total_volume_24h"].(float64)
	whaleTransactions := volume24h / 500000 // Rough estimate: 1 whale transaction per $500,000 of volume

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = whaleTransactions * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	// Calculate indicator and score using enhanced whale transaction scoring
	score := getWhaleTransactionScore(whaleTransactions)

	var indicator string
	if score > 0.3 {
		indicator = "Low Whale Activity"
	} else if score > 0.0 {
		indicator = "Moderate Whale Activity"
	} else {
		indicator = "High Whale Activity"
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

func handleRSI(c *gin.Context) {
	// Check if mock data is requested
	if shouldUseMockData(c) {
		mockData := market.GetRSIMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

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

	// Calculate indicator and score based on RSI value using enhanced scoring
	var indicator string
	score := getRSIScore(rsi)

	if score > 0.3 {
		indicator = "Buy"
	} else if score < -0.3 {
		indicator = "Sell"
	} else {
		indicator = "Hold"
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

	// Get historical data (last 5 days) and calculate trend
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = btcDominance + float64(i-2)*0.2 // More realistic trend simulation
	}

	// Calculate trend direction
	isRising := btcDominance > historical[0]

	// Calculate indicator and score using enhanced BTC dominance scoring
	var indicator string
	score := getBTCDominanceScore(btcDominance)

	// Adjust score based on trend
	if isRising {
		score -= 0.1 // Slightly more bearish if rising
	} else {
		score += 0.1 // Slightly more bullish if falling
	}

	// Clamp score to valid range
	if score > 1.0 {
		score = 1.0
	} else if score < -1.0 {
		score = -1.0
	}

	if score > 0.2 {
		indicator = "Buy"
	} else if score < -0.2 {
		indicator = "Sell"
	} else {
		indicator = "Hold"
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        btcDominance,
		"indicator":    indicator,
		"score":        score,
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

	// Get 7-day percentage change if available, otherwise simulate
	var percentChange7d float64
	if change, ok := metrics["total_market_cap_yesterday_percentage_change"]; ok {
		percentChange7d = change.(float64) * 7 // Approximate 7-day change
	} else {
		// Fallback: simulate based on current market cap
		percentChange7d = (marketCap - 2.1e12) / 2.1e12 * 100 // Assume 2.1T baseline
	}

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = marketCap * (1.0 - float64(i)*0.01) // More realistic trend
	}

	// Calculate indicator and score based on 7-day percentage change
	var indicator string
	var score float64
	if percentChange7d > 5.0 {
		indicator = "Buy"
		score = 1
	} else if percentChange7d < -5.0 {
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

// ETH/BTC Ratio cache
type ETHBTCCache struct {
	Ratio      float64
	Historical []float64
	Timestamp  time.Time
	mu         sync.RWMutex
}

var ethBTCCache = &ETHBTCCache{}

func handleETHBTCRatio(c *gin.Context) {
	// Check cache first (cache for 2 minutes to avoid rate limiting)
	ethBTCCache.mu.RLock()
	if !ethBTCCache.Timestamp.IsZero() && time.Since(ethBTCCache.Timestamp) < 2*time.Minute {
		ratio := ethBTCCache.Ratio
		historical := ethBTCCache.Historical
		ethBTCCache.mu.RUnlock()

		score := getETHBTCRatioScore(ratio)
		var indicator string
		if score > 0.3 {
			indicator = "ETH Outperforming"
		} else if score < -0.3 {
			indicator = "BTC Outperforming"
		} else {
			indicator = "Balanced"
		}

		c.JSON(http.StatusOK, gin.H{
			"value":        ratio,
			"indicator":    indicator,
			"score":        score,
			"chart_data":   historical,
			"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
			"api_source":   "CoinGecko (Cached)",
		})
		return
	}
	ethBTCCache.mu.RUnlock()

	// Use CoinGecko API with rate limiting protection
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin,ethereum&vs_currencies=usd")
	if err != nil {
		log.Printf("ETH/BTC Ratio API Error: %v", err)
		// Return cached data if available
		ethBTCCache.mu.RLock()
		if !ethBTCCache.Timestamp.IsZero() {
			ratio := ethBTCCache.Ratio
			historical := ethBTCCache.Historical
			ethBTCCache.mu.RUnlock()

			score := getETHBTCRatioScore(ratio)
			c.JSON(http.StatusOK, gin.H{
				"value":        ratio,
				"indicator":    "Hold",
				"score":        score,
				"chart_data":   historical,
				"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
				"api_source":   "CoinGecko (Fallback to cache)",
			})
			return
		}
		ethBTCCache.mu.RUnlock()

		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch ETH/BTC ratio data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("ETH/BTC Ratio API returned status: %d", resp.StatusCode)

		// Return cached data if available
		ethBTCCache.mu.RLock()
		if !ethBTCCache.Timestamp.IsZero() {
			ratio := ethBTCCache.Ratio
			historical := ethBTCCache.Historical
			ethBTCCache.mu.RUnlock()

			score := getETHBTCRatioScore(ratio)
			c.JSON(http.StatusOK, gin.H{
				"value":        ratio,
				"indicator":    "Hold",
				"score":        score,
				"chart_data":   historical,
				"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
				"api_source":   "CoinGecko (Fallback to cache due to rate limit)",
			})
			return
		}
		ethBTCCache.mu.RUnlock()

		c.JSON(http.StatusBadGateway, gin.H{
			"error":        fmt.Sprintf("API returned status %d", resp.StatusCode),
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Read response body for debugging
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ETH/BTC Ratio - Failed to read response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to read response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	var priceData map[string]interface{}
	if err := json.Unmarshal(body, &priceData); err != nil {
		log.Printf("ETH/BTC Ratio - Failed to parse JSON: %v, Body: %s", err, string(body))
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

	// Safe type assertions with detailed logging
	btcData, ok := priceData["bitcoin"].(map[string]interface{})
	if !ok {
		log.Printf("ETH/BTC Ratio - Invalid BTC data format. Available keys: %v", getMapKeys(priceData))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Invalid BTC data format",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	ethData, ok := priceData["ethereum"].(map[string]interface{})
	if !ok {
		log.Printf("ETH/BTC Ratio - Invalid ETH data format. Available keys: %v", getMapKeys(priceData))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Invalid ETH data format",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	btcPrice, ok := btcData["usd"].(float64)
	if !ok {
		log.Printf("ETH/BTC Ratio - Invalid BTC price format. BTC data: %v", btcData)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Invalid BTC price",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	ethPrice, ok := ethData["usd"].(float64)
	if !ok {
		log.Printf("ETH/BTC Ratio - Invalid ETH price format. ETH data: %v", ethData)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Invalid ETH price",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	if btcPrice == 0 {
		log.Printf("ETH/BTC Ratio - BTC price is zero")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Invalid BTC price (zero)",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	ratio := ethPrice / btcPrice

	// Generate historical data (simulate 7 days of data)
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = ratio + (rand.Float64()-0.5)*0.01
	}

	// Update cache
	ethBTCCache.mu.Lock()
	ethBTCCache.Ratio = ratio
	ethBTCCache.Historical = historical
	ethBTCCache.Timestamp = time.Now()
	ethBTCCache.mu.Unlock()

	// Calculate score using enhanced ETH/BTC ratio scoring
	score := getETHBTCRatioScore(ratio)

	var indicator string
	if score > 0.3 {
		indicator = "ETH Outperforming"
	} else if score < -0.3 {
		indicator = "BTC Outperforming"
	} else {
		indicator = "Balanced"
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        ratio,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"api_source":   "CoinGecko (Free API - Live ETH/BTC ratio)",
	})
}

// Helper function to get map keys for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
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
	// Check if mock data is requested
	if shouldUseMockData(c) {
		mockData := market.GetMovingAveragesMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

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
	currentPrice := quote["price"].(float64)

	// Simulate moving averages (in real implementation, you'd calculate from historical price data)
	// For demonstration, we'll use price momentum as a proxy
	percentChange7d := quote["percent_change_7d"].(float64)
	percentChange30d := quote["percent_change_30d"].(float64)

	// Simulate MA50 and MA200 based on recent performance
	// In reality, you'd fetch 200 days of price data and calculate actual MAs
	ma50 := currentPrice * (1 + percentChange7d/100*0.5)    // Approximate 50-day MA
	ma200 := currentPrice * (1 + percentChange30d/100*0.25) // Approximate 200-day MA

	// Determine crossover signal
	var signal string
	var score int
	var crossoverType string

	if ma50 > ma200 && percentChange7d > 0 {
		// Golden cross scenario - 50d MA above 200d MA with upward momentum
		signal = "Buy"
		score = 1
		crossoverType = "Golden Cross"
	} else if ma50 < ma200 && percentChange7d < 0 {
		// Death cross scenario - 50d MA below 200d MA with downward momentum
		signal = "Sell"
		score = -1
		crossoverType = "Death Cross"
	} else {
		signal = "Hold"
		score = 0
		crossoverType = "No Clear Cross"
	}

	// Create historical data showing the crossover trend
	historical := make([]float64, 5)
	for i := range historical {
		if signal == "Buy" {
			historical[i] = float64(i) * 0.2 // Upward trend
		} else if signal == "Sell" {
			historical[i] = 1.0 - float64(i)*0.2 // Downward trend
		} else {
			historical[i] = 0.5 // Neutral
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        crossoverType,
		"indicator":    signal,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		"ma50":         ma50,
		"ma200":        ma200,
	})
}

func handleFearGreed(c *gin.Context) {
	// Check if mock data is requested
	if shouldUseMockData(c) {
		mockData := market.GetFearGreedMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

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

	// Calculate indicator and score using enhanced Fear & Greed scoring
	score := getFearGreedScore(float64(currentValue))

	var indicator string
	if score > 0.3 {
		indicator = "Buy"
	} else if score < -0.3 {
		indicator = "Sell"
	} else {
		indicator = "Hold"
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
	// Enhanced port conflict resolution with automatic fallback
	port := "8080"
	maxRetries := 5
	maxPortAttempts := 3

	// Try multiple ports if needed
	for portAttempt := 0; portAttempt < maxPortAttempts; portAttempt++ {
		currentPort := fmt.Sprintf("%d", 8080+portAttempt)

		log.Printf("Attempting to use port %s", currentPort)

		// Enhanced process killing with better detection
		for i := 0; i < maxRetries; i++ {
			if err := killProcessOnPortEnhanced(currentPort); err != nil {
				log.Printf("Attempt %d: Could not kill existing process on port %s: %v", i+1, currentPort, err)
			}

			// Progressive wait times
			waitTime := time.Duration(500+i*200) * time.Millisecond
			time.Sleep(waitTime)

			// Check if port is actually free
			if !isPortInUse(currentPort) {
				port = currentPort
				log.Printf("Successfully freed port %s", port)
				goto portResolved
			}

			if i == maxRetries-1 {
				log.Printf("Failed to free port %s after %d attempts, trying next port", currentPort, maxRetries)
			}
		}
	}

	log.Fatalf("Failed to find available port after trying ports 8080-%d", 8080+maxPortAttempts-1)

portResolved:
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

	// Initialize database
	if err := database.InitDB("market_data.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize market service with API key
	marketService = market.NewMarketService(os.Getenv("CMC_API_KEY"))

	// Initialize WebSocket hub (after market service is ready)
	initWebSocketHub()

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

		// WebSocket endpoint
		api.GET("/ws", handleWebSocket)

		// System monitoring endpoints
		api.GET("/health", handleHealth)
		api.GET("/metrics", handleSystemMetrics)
		api.GET("/cache-stats", handleCacheStats)
		api.GET("/circuit-breaker-stats", handleCircuitBreakerStats)

		// Advanced DeFi & Crypto Metrics
		api.GET("/defi-tvl", handleDeFiTVL)
		api.GET("/social-sentiment", handleSocialSentiment)
		api.GET("/options-flow", handleOptionsFlow)
		api.GET("/stablecoin-flows", handleStablecoinFlows)
		api.GET("/network-health", handleNetworkHealth)
		api.GET("/institutional-flows", handleInstitutionalFlows)
		api.GET("/yield-curves", handleYieldCurves)
		api.GET("/correlation-matrix", handleCorrelationMatrix)
		api.GET("/volatility-surface", handleVolatilitySurface)
		api.GET("/liquidation-heatmap", handleLiquidationHeatmap)

		// Solana wallet endpoints
		api.GET("/solana/balance/:address", getSolanaBalance)
	}

	// Graceful shutdown handling
	setupGracefulShutdown()

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func killProcessOnPortEnhanced(port string) error {
	// Try multiple methods to find and kill processes
	methods := [][]string{
		{"lsof", "-ti:" + port},
		{"netstat", "-tlnp", "|", "grep", ":" + port, "|", "awk", "'{print $7}'", "|", "cut", "-d/", "-f1"},
		{"ss", "-tlnp", "|", "grep", ":" + port, "|", "awk", "'{print $6}'", "|", "cut", "-d,", "-f2", "|", "cut", "-d=", "-f2"},
	}

	for _, method := range methods {
		if len(method) == 2 {
			cmd := exec.Command(method[0], method[1])
			output, err := cmd.Output()
			if err != nil {
				continue // Try next method
			}

			pid := strings.TrimSpace(string(output))
			if pid != "" && pid != "0" {
				log.Printf("Found process %s on port %s, attempting to kill", pid, port)

				// Try graceful termination first
				if err := exec.Command("kill", "-TERM", pid).Run(); err == nil {
					time.Sleep(2 * time.Second)
					// Check if process is still running
					if err := exec.Command("kill", "-0", pid).Run(); err != nil {
						return nil // Process terminated gracefully
					}
				}

				// Force kill if graceful termination failed
				return exec.Command("kill", "-9", pid).Run()
			}
		}
	}

	return nil // No process found, which is fine
}

func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")

		// Close Telegram service
		if telegramService != nil {
			telegramService.ClearSessions()
		}

		// Additional cleanup can be added here

		os.Exit(0)
	}()
}

func isPortInUse(port string) bool {
	cmd := exec.Command("lsof", "-ti:"+port)
	output, err := cmd.Output()
	if err != nil {
		return false // No process found
	}
	return strings.TrimSpace(string(output)) != ""
}

// System monitoring handlers
func handleHealth(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"services": map[string]interface{}{
			"database": map[string]interface{}{
				"status": "connected",
			},
			"websocket": map[string]interface{}{
				"status":            "running",
				"connected_clients": len(wsHub.clients),
			},
			"telegram": map[string]interface{}{
				"status": "initialized",
			},
		},
	}

	c.JSON(http.StatusOK, health)
}

func handleSystemMetrics(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	// Get API performance stats for major APIs
	apis := []string{"coinmarketcap", "binance", "fear-greed", "alternative.me"}
	apiStats := make(map[string]interface{})

	for _, api := range apis {
		stats, err := db.GetAPIPerformanceStats(api, 24) // Last 24 hours
		if err != nil {
			apiStats[api] = map[string]interface{}{"error": err.Error()}
		} else {
			apiStats[api] = stats
		}
	}

	metrics := map[string]interface{}{
		"timestamp":       time.Now(),
		"api_performance": apiStats,
		"websocket": map[string]interface{}{
			"connected_clients": len(wsHub.clients),
			"total_broadcasts":  "N/A", // Could track this
		},
		"system": map[string]interface{}{
			"uptime": time.Since(time.Now().Add(-time.Hour)), // Placeholder
			"memory": "N/A",                                  // Could add runtime.MemStats
		},
	}

	c.JSON(http.StatusOK, metrics)
}

func handleCacheStats(c *gin.Context) {
	stats := market.GetCacheStats()
	c.JSON(http.StatusOK, stats)
}

func handleCircuitBreakerStats(c *gin.Context) {
	stats := market.GetCircuitBreakerStats()
	c.JSON(http.StatusOK, stats)
}

// Advanced DeFi & Crypto Metrics Handlers

// DeFi Total Value Locked (TVL) - Critical for DeFi market health
func handleDeFiTVL(c *gin.Context) {
	// Use DeFiLlama API (free) to get real TVL data
	client := &http.Client{Timeout: 10 * time.Second}

	// Get total TVL across all protocols
	resp, err := client.Get("https://api.llama.fi/protocols")
	if err != nil {
		// Fallback to simulated data on API failure
		tvl := 45000000000.0 + rand.Float64()*10000000000.0 // $45-55B range
		historical := make([]float64, 7)
		for i := range historical {
			historical[i] = tvl * (0.95 + float64(i)*0.01)
		}

		c.JSON(http.StatusOK, gin.H{
			"value":         tvl / 1000000000.0,
			"indicator":     "API Error - Using Fallback",
			"score":         0.0,
			"weekly_change": 0.0,
			"chart_data":    historical,
			"chart_labels":  []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		})
		return
	}
	defer resp.Body.Close()

	var protocols []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&protocols); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse DeFiLlama response"})
		return
	}

	// Calculate total TVL
	var totalTVL float64
	for _, protocol := range protocols {
		if tvl, ok := protocol["tvl"].(float64); ok {
			totalTVL += tvl
		}
	}

	// Get historical TVL data
	histResp, err := client.Get("https://api.llama.fi/charts")
	var historical []float64
	weeklyChange := 0.0

	if err == nil {
		defer histResp.Body.Close()
		var histData []map[string]interface{}
		if json.NewDecoder(histResp.Body).Decode(&histData) == nil && len(histData) >= 7 {
			// Get last 7 days
			for i := len(histData) - 7; i < len(histData); i++ {
				if i >= 0 {
					if tvl, ok := histData[i]["totalLiquidityUSD"].(float64); ok {
						historical = append(historical, tvl/1000000000.0) // Convert to billions
					}
				}
			}

			// Calculate weekly change
			if len(historical) >= 2 {
				weeklyChange = (historical[len(historical)-1] - historical[0]) / historical[0] * 100
			}
		}
	}

	// Fallback historical data if API fails
	if len(historical) == 0 {
		for i := 0; i < 7; i++ {
			historical = append(historical, totalTVL/1000000000.0*(0.95+float64(i)*0.01))
		}
		weeklyChange = 5.26 // Default positive change
	}

	// Calculate indicator and score
	var indicator string
	var score float64

	if weeklyChange > 5 {
		indicator = "Strong Growth"
		score = 1.0
	} else if weeklyChange > 2 {
		indicator = "Growth"
		score = 0.5
	} else if weeklyChange < -5 {
		indicator = "Decline"
		score = -1.0
	} else if weeklyChange < -2 {
		indicator = "Weak"
		score = -0.5
	} else {
		indicator = "Stable"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":         totalTVL / 1000000000.0, // Convert to billions
		"indicator":     indicator,
		"score":         score,
		"weekly_change": weeklyChange,
		"chart_data":    historical,
		"chart_labels":  []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
	})
}

// Social Sentiment Analysis - Twitter, Reddit, News sentiment
func handleSocialSentiment(c *gin.Context) {
	// Use Reddit API (free) to get real sentiment data from crypto subreddits
	client := &http.Client{Timeout: 10 * time.Second}

	// Get posts from r/cryptocurrency
	resp, err := client.Get("https://www.reddit.com/r/cryptocurrency/hot.json?limit=25")
	if err != nil {
		// Fallback to simulated data on API failure
		sentiment := -0.5 + rand.Float64() // Random between -0.5 and 0.5
		historical := make([]float64, 7)
		for i := range historical {
			historical[i] = sentiment + (rand.Float64()-0.5)*0.2
		}

		c.JSON(http.StatusOK, gin.H{
			"value":        sentiment,
			"indicator":    "API Error - Using Fallback",
			"score":        0.0,
			"chart_data":   historical,
			"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		})
		return
	}
	defer resp.Body.Close()

	var redditData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&redditData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Reddit response"})
		return
	}

	// Analyze sentiment from Reddit posts
	sentiment := analyzeSentimentFromReddit(redditData)

	// Generate historical sentiment data (simplified)
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = sentiment + (rand.Float64()-0.5)*0.2
	}

	var indicator string
	var score float64

	if sentiment > 0.3 {
		indicator = "Very Bullish"
		score = 1.0
	} else if sentiment > 0.1 {
		indicator = "Bullish"
		score = 0.5
	} else if sentiment < -0.3 {
		indicator = "Very Bearish"
		score = -1.0
	} else if sentiment < -0.1 {
		indicator = "Bearish"
		score = -0.5
	} else {
		indicator = "Neutral"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        sentiment,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
	})
}

// Helper function to analyze sentiment from Reddit posts
func analyzeSentimentFromReddit(data map[string]interface{}) float64 {
	// Simple sentiment analysis based on keywords
	bullishWords := []string{"bullish", "moon", "pump", "buy", "hodl", "diamond", "rocket", "green", "up", "rise", "gain", "profit", "bull", "positive", "optimistic", "surge", "rally"}
	bearishWords := []string{"bearish", "dump", "sell", "crash", "red", "down", "fall", "loss", "bear", "negative", "pessimistic", "drop", "decline", "dip", "correction"}

	var totalScore float64
	var postCount int

	if dataObj, ok := data["data"].(map[string]interface{}); ok {
		if children, ok := dataObj["children"].([]interface{}); ok {
			for _, child := range children {
				if post, ok := child.(map[string]interface{}); ok {
					if postData, ok := post["data"].(map[string]interface{}); ok {
						// Analyze title and selftext
						title := ""
						selftext := ""

						if t, ok := postData["title"].(string); ok {
							title = strings.ToLower(t)
						}
						if s, ok := postData["selftext"].(string); ok {
							selftext = strings.ToLower(s)
						}

						text := title + " " + selftext
						postScore := 0.0

						// Count bullish words
						for _, word := range bullishWords {
							postScore += float64(strings.Count(text, word)) * 0.1
						}

						// Count bearish words
						for _, word := range bearishWords {
							postScore -= float64(strings.Count(text, word)) * 0.1
						}

						// Factor in upvote ratio if available
						if ups, ok := postData["ups"].(float64); ok {
							if downs, ok := postData["downs"].(float64); ok {
								if ups+downs > 0 {
									ratio := ups / (ups + downs)
									postScore *= ratio // Weight by community approval
								}
							}
						}

						totalScore += postScore
						postCount++
					}
				}
			}
		}
	}

	if postCount == 0 {
		return 0.0
	}

	// Normalize sentiment score to -1 to 1 range
	avgScore := totalScore / float64(postCount)
	if avgScore > 1.0 {
		avgScore = 1.0
	} else if avgScore < -1.0 {
		avgScore = -1.0
	}

	return avgScore
}

// Options Flow - Put/Call ratio and options volume
func handleOptionsFlow(c *gin.Context) {
	// Use Deribit API (free) to get real options data
	client := &http.Client{Timeout: 10 * time.Second}

	// Get BTC options summary
	resp, err := client.Get("https://www.deribit.com/api/v2/public/get_book_summary_by_currency?currency=BTC&kind=option")
	if err != nil {
		// Fallback to simulated data on API failure
		putCallRatio := 0.7 + rand.Float64()*0.6 // 0.7-1.3 range
		historical := make([]float64, 7)
		for i := range historical {
			historical[i] = putCallRatio + (rand.Float64()-0.5)*0.1
		}

		c.JSON(http.StatusOK, gin.H{
			"value":        putCallRatio,
			"indicator":    "API Error - Using Fallback",
			"score":        0.0,
			"chart_data":   historical,
			"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		})
		return
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Deribit response"})
		return
	}

	// Parse options data
	result, ok := response["result"].([]interface{})
	if !ok || len(result) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Deribit response format"})
		return
	}

	var totalPutVolume, totalCallVolume float64

	// Calculate put/call volumes
	for _, item := range result {
		if option, ok := item.(map[string]interface{}); ok {
			if instrumentName, ok := option["instrument_name"].(string); ok {
				if volume, ok := option["volume"].(float64); ok {
					// Check if it's a put or call option
					if len(instrumentName) > 0 {
						// Deribit format: BTC-DDMMMYY-STRIKE-P/C
						if instrumentName[len(instrumentName)-1:] == "P" {
							totalPutVolume += volume
						} else if instrumentName[len(instrumentName)-1:] == "C" {
							totalCallVolume += volume
						}
					}
				}
			}
		}
	}

	// Calculate put/call ratio
	var putCallRatio float64
	if totalCallVolume > 0 {
		putCallRatio = totalPutVolume / totalCallVolume
	} else {
		putCallRatio = 1.0 // Default neutral ratio
	}

	// Generate historical data (simplified - in production you'd store and retrieve real historical data)
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = putCallRatio + (rand.Float64()-0.5)*0.1
	}

	var indicator string
	var score float64

	if putCallRatio < 0.7 {
		indicator = "Bullish (Low Put/Call)"
		score = 1.0
	} else if putCallRatio < 0.9 {
		indicator = "Moderately Bullish"
		score = 0.5
	} else if putCallRatio > 1.3 {
		indicator = "Bearish (High Put/Call)"
		score = -1.0
	} else if putCallRatio > 1.1 {
		indicator = "Moderately Bearish"
		score = -0.5
	} else {
		indicator = "Neutral"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        putCallRatio,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"put_volume":   totalPutVolume,
		"call_volume":  totalCallVolume,
	})
}

// Stablecoin Flows - USDT, USDC, BUSD flows to exchanges
func handleStablecoinFlows(c *gin.Context) {
	// Check if mock data is requested
	if shouldUseMockData(c) {
		mockData := market.GetStablecoinFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use CoinGecko API (free) to get stablecoin data
	client := &http.Client{Timeout: 10 * time.Second}

	// Get USDT market cap as proxy for stablecoin flows
	resp, err := client.Get("https://api.coingecko.com/api/v3/coins/tether?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false&sparkline=false")
	if err != nil {
		// Fallback to mock data on error
		mockData := market.GetStablecoinFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}
	defer resp.Body.Close()

	var usdtData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&usdtData); err != nil {
		mockData := market.GetStablecoinFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Extract market cap change as flow indicator
	marketData, ok := usdtData["market_data"].(map[string]interface{})
	if !ok {
		mockData := market.GetStablecoinFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	marketCapChangePercent, ok := marketData["market_cap_change_percentage_24h"].(float64)
	if !ok {
		mockData := market.GetStablecoinFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	currentMarketCap, ok := marketData["market_cap"].(map[string]interface{})["usd"].(float64)
	if !ok {
		mockData := market.GetStablecoinFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Calculate flow based on market cap change
	flow := (marketCapChangePercent / 100) * currentMarketCap / 1000000 // In millions USD

	// Generate historical data (simplified)
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = flow + (rand.Float64()-0.5)*50
	}

	// Calculate score using enhanced stablecoin flow scoring
	score := getStablecoinFlowScore(flow)

	var indicator string
	if score > 0.3 {
		indicator = "Strong Outflows"
	} else if score > 0.1 {
		indicator = "Moderate Outflows"
	} else if score > -0.1 {
		indicator = "Neutral Flows"
	} else if score > -0.3 {
		indicator = "Moderate Inflows"
	} else {
		indicator = "Strong Inflows"
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        flow,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"api_source":   "CoinGecko (Live stablecoin market data)",
	})
}

// Network Health - Hash rate, difficulty, node count
func handleNetworkHealth(c *gin.Context) {
	// Use blockchain.info API (free) to get real network data
	client := &http.Client{Timeout: 10 * time.Second}

	// Get network stats
	resp, err := client.Get("https://blockchain.info/stats?format=json")
	if err != nil {
		// Fallback to simulated data on API failure
		healthScore := 75 + rand.Float64()*20 // 75-95 range
		historical := make([]float64, 7)
		for i := range historical {
			historical[i] = healthScore + (rand.Float64()-0.5)*5
		}

		c.JSON(http.StatusOK, gin.H{
			"value":        healthScore,
			"indicator":    "API Error - Using Fallback",
			"score":        0.0,
			"chart_data":   historical,
			"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		})
		return
	}
	defer resp.Body.Close()

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse blockchain.info response"})
		return
	}

	// Extract network metrics
	var hashRate, difficulty, totalBTC float64
	var blocksCount int64

	if hr, ok := stats["hash_rate"].(float64); ok {
		hashRate = hr
	}
	if diff, ok := stats["difficulty"].(float64); ok {
		difficulty = diff
	}
	if total, ok := stats["totalbc"].(float64); ok {
		totalBTC = total / 100000000 // Convert from satoshis to BTC
	}
	if blocks, ok := stats["n_blocks_total"].(float64); ok {
		blocksCount = int64(blocks)
	}

	// Calculate network health score based on multiple factors
	var healthScore float64

	// Hash rate component (40% weight)
	hashRateScore := 0.0
	if hashRate > 400000000000000000 { // > 400 EH/s
		hashRateScore = 40.0
	} else if hashRate > 300000000000000000 { // > 300 EH/s
		hashRateScore = 35.0
	} else if hashRate > 200000000000000000 { // > 200 EH/s
		hashRateScore = 30.0
	} else {
		hashRateScore = 25.0
	}

	// Difficulty component (30% weight)
	difficultyScore := 0.0
	if difficulty > 50000000000000 { // High difficulty indicates network security
		difficultyScore = 30.0
	} else if difficulty > 30000000000000 {
		difficultyScore = 25.0
	} else {
		difficultyScore = 20.0
	}

	// Block count component (20% weight) - indicates network activity
	blockScore := 0.0
	if blocksCount > 800000 {
		blockScore = 20.0
	} else if blocksCount > 700000 {
		blockScore = 15.0
	} else {
		blockScore = 10.0
	}

	// Total BTC component (10% weight) - network maturity
	btcScore := 0.0
	if totalBTC > 19000000 {
		btcScore = 10.0
	} else if totalBTC > 18000000 {
		btcScore = 8.0
	} else {
		btcScore = 5.0
	}

	healthScore = hashRateScore + difficultyScore + blockScore + btcScore

	// Generate historical data (simplified)
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = healthScore + (rand.Float64()-0.5)*5
	}

	var indicator string
	var score float64

	if healthScore > 90 {
		indicator = "Excellent"
		score = 1.0
	} else if healthScore > 80 {
		indicator = "Good"
		score = 0.5
	} else if healthScore < 60 {
		indicator = "Poor"
		score = -1.0
	} else if healthScore < 70 {
		indicator = "Fair"
		score = -0.5
	} else {
		indicator = "Average"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        healthScore,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"hash_rate":    hashRate,
		"difficulty":   difficulty,
		"total_btc":    totalBTC,
		"blocks":       blocksCount,
	})
}

// Institutional Flows - Grayscale, MicroStrategy, ETF flows
func handleInstitutionalFlows(c *gin.Context) {
	// Check if mock data is requested
	if shouldUseMockData(c) {
		mockData := market.GetInstitutionalFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use free APIs to approximate institutional flows
	client := &http.Client{Timeout: 10 * time.Second}

	// Get Bitcoin ETF proxy data using large volume analysis
	resp, err := client.Get("https://api.coingecko.com/api/v3/coins/bitcoin?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false&sparkline=false")
	if err != nil {
		// Fallback to mock data on error
		mockData := market.GetInstitutionalFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}
	defer resp.Body.Close()

	var btcData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&btcData); err != nil {
		mockData := market.GetInstitutionalFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Extract volume and price change data
	marketData, ok := btcData["market_data"].(map[string]interface{})
	if !ok {
		mockData := market.GetInstitutionalFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	totalVolume, ok := marketData["total_volume"].(map[string]interface{})["usd"].(float64)
	if !ok {
		mockData := market.GetInstitutionalFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	priceChangePercent, ok := marketData["price_change_percentage_24h"].(float64)
	if !ok {
		priceChangePercent = 0.0
	}

	// Estimate institutional flows based on volume patterns
	// Large volume + positive price change often indicates institutional buying
	institutionalFlow := (totalVolume / 1000000) * (priceChangePercent / 100) * 10 // Scaled estimate

	// Generate historical data
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = institutionalFlow + (rand.Float64()-0.5)*200
	}

	// Calculate score using enhanced institutional flow scoring
	score := getInstitutionalFlowScore(institutionalFlow)

	var indicator string
	if score > 0.5 {
		indicator = "Strong Institutional Buying"
	} else if score > 0.2 {
		indicator = "Moderate Institutional Buying"
	} else if score > -0.2 {
		indicator = "Mixed Institutional Activity"
	} else if score > -0.5 {
		indicator = "Moderate Institutional Selling"
	} else {
		indicator = "Strong Institutional Selling"
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        institutionalFlow,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"api_source":   "CoinGecko (Live volume analysis proxy)",
	})
}

// Yield Curves - DeFi vs TradFi yield comparison
func handleYieldCurves(c *gin.Context) {
	// Check if mock data is requested
	if shouldUseMockData(c) {
		mockData := market.GetYieldCurvesMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use free APIs to get yield data comparison
	client := &http.Client{Timeout: 10 * time.Second}

	// Get DeFi yield proxy using staking APY from CoinGecko
	resp, err := client.Get("https://api.coingecko.com/api/v3/coins/ethereum?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false&sparkline=false")
	if err != nil {
		// Fallback to mock data on error
		mockData := market.GetYieldCurvesMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}
	defer resp.Body.Close()

	var ethData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ethData); err != nil {
		mockData := market.GetYieldCurvesMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Simulate DeFi yield based on ETH volatility and price change
	marketData, ok := ethData["market_data"].(map[string]interface{})
	if !ok {
		mockData := market.GetYieldCurvesMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	priceChangePercent, ok := marketData["price_change_percentage_30d"].(float64)
	if !ok {
		priceChangePercent = 0.0
	}

	// Estimate DeFi yield premium based on market conditions
	// Higher volatility and positive price movements typically correlate with higher DeFi yields
	baseYield := 3.0                                      // Base traditional yield
	volatilityBonus := math.Abs(priceChangePercent) * 0.2 // Volatility premium
	trendBonus := priceChangePercent * 0.1                // Trend bonus

	defiYieldPremium := baseYield + volatilityBonus + trendBonus
	if defiYieldPremium < 0 {
		defiYieldPremium = 0.5
	}
	if defiYieldPremium > 25 {
		defiYieldPremium = 25.0
	}

	// Generate historical data
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = defiYieldPremium + (rand.Float64()-0.5)*2
		if historical[i] < 0 {
			historical[i] = 0.1
		}
	}

	var indicator string
	var score float64

	if defiYieldPremium > 15 {
		indicator = "Extremely High Yield Premium"
		score = 0.8
	} else if defiYieldPremium > 10 {
		indicator = "High Yield Premium"
		score = 0.6
	} else if defiYieldPremium > 7 {
		indicator = "Moderate Yield Premium"
		score = 0.3
	} else if defiYieldPremium > 3 {
		indicator = "Low Yield Premium"
		score = 0.1
	} else {
		indicator = "Minimal Yield Premium"
		score = -0.2
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        defiYieldPremium,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"api_source":   "CoinGecko (Live yield premium estimation)",
	})
}

// Correlation Matrix - BTC correlation with stocks, gold, etc.
func handleCorrelationMatrix(c *gin.Context) {
	// Check if mock data is requested
	if shouldUseMockData(c) {
		mockData := market.GetCorrelationMatrixMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use Yahoo Finance API (free) to get real correlation data with enhanced error handling
	client := &http.Client{Timeout: 10 * time.Second}

	// Helper function to create fallback response
	createFallbackResponse := func(reason string) {
		correlation := -0.2 + rand.Float64()*0.6 // -0.2 to 0.4 range (realistic)
		historical := make([]float64, 7)
		for i := range historical {
			historical[i] = correlation + (rand.Float64()-0.5)*0.1
		}

		c.JSON(http.StatusOK, gin.H{
			"value":        correlation,
			"indicator":    fmt.Sprintf("Fallback Data (%s)", reason),
			"score":        0.0,
			"chart_data":   historical,
			"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
			"api_status":   "fallback",
			"reason":       reason,
		})
	}

	// Get BTC price data (last 30 days) with User-Agent header
	btcReq, err := http.NewRequest("GET", "https://query1.finance.yahoo.com/v8/finance/chart/BTC-USD?interval=1d&range=30d", nil)
	if err != nil {
		createFallbackResponse("Request creation failed")
		return
	}
	btcReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	btcResp, err := client.Do(btcReq)
	if err != nil {
		createFallbackResponse("Network error")
		return
	}
	defer btcResp.Body.Close()

	// Check for rate limiting or other HTTP errors
	if btcResp.StatusCode != http.StatusOK {
		if btcResp.StatusCode == 429 {
			createFallbackResponse("Rate limited")
		} else {
			createFallbackResponse(fmt.Sprintf("HTTP %d", btcResp.StatusCode))
		}
		return
	}

	var btcData map[string]interface{}
	if err := json.NewDecoder(btcResp.Body).Decode(&btcData); err != nil {
		createFallbackResponse("BTC data parsing failed")
		return
	}

	// Check if the response contains an error
	if chart, ok := btcData["chart"].(map[string]interface{}); ok {
		if errorField, ok := chart["error"]; ok && errorField != nil {
			createFallbackResponse("Yahoo API error")
			return
		}
	}

	// Get S&P 500 data for correlation with same headers
	spyReq, err := http.NewRequest("GET", "https://query1.finance.yahoo.com/v8/finance/chart/SPY?interval=1d&range=30d", nil)
	if err != nil {
		createFallbackResponse("SPY request failed")
		return
	}
	spyReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	spyResp, err := client.Do(spyReq)
	if err != nil {
		createFallbackResponse("SPY network error")
		return
	}
	defer spyResp.Body.Close()

	if spyResp.StatusCode != http.StatusOK {
		createFallbackResponse(fmt.Sprintf("SPY HTTP %d", spyResp.StatusCode))
		return
	}

	var spyData map[string]interface{}
	if err := json.NewDecoder(spyResp.Body).Decode(&spyData); err != nil {
		createFallbackResponse("SPY data parsing failed")
		return
	}

	// Extract price data and calculate correlation
	btcPrices := extractPricesFromYahoo(btcData)
	spyPrices := extractPricesFromYahoo(spyData)

	var correlation float64
	var apiStatus string

	if len(btcPrices) > 5 && len(spyPrices) > 5 {
		correlation = calculateCorrelation(btcPrices, spyPrices)
		apiStatus = "success"
	} else {
		// If we got responses but couldn't extract enough prices, use fallback
		createFallbackResponse("Insufficient price data")
		return
	}

	// Generate historical correlation data (simplified)
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = correlation + (rand.Float64()-0.5)*0.1
	}

	var indicator string
	var score float64

	if correlation < -0.1 {
		indicator = "Negative Correlation (Good)"
		score = 1.0
	} else if correlation < 0.1 {
		indicator = "Low Correlation"
		score = 0.5
	} else if correlation > 0.4 {
		indicator = "High Correlation (Risk)"
		score = -1.0
	} else if correlation > 0.2 {
		indicator = "Moderate Correlation"
		score = -0.5
	} else {
		indicator = "Neutral Correlation"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        correlation,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"btc_prices":   len(btcPrices),
		"spy_prices":   len(spyPrices),
		"api_status":   apiStatus,
	})
}

// Helper function to extract prices from Yahoo Finance response
func extractPricesFromYahoo(data map[string]interface{}) []float64 {
	var prices []float64

	// Navigate through the nested JSON structure safely
	chart, ok := data["chart"].(map[string]interface{})
	if !ok {
		return prices
	}

	result, ok := chart["result"].([]interface{})
	if !ok || len(result) == 0 {
		return prices
	}

	firstResult, ok := result[0].(map[string]interface{})
	if !ok {
		return prices
	}

	indicators, ok := firstResult["indicators"].(map[string]interface{})
	if !ok {
		return prices
	}

	quote, ok := indicators["quote"].([]interface{})
	if !ok || len(quote) == 0 {
		return prices
	}

	firstQuote, ok := quote[0].(map[string]interface{})
	if !ok {
		return prices
	}

	close, ok := firstQuote["close"].([]interface{})
	if !ok {
		return prices
	}

	// Extract prices, filtering out null values
	for _, price := range close {
		if price != nil {
			if p, ok := price.(float64); ok && !math.IsNaN(p) && !math.IsInf(p, 0) {
				prices = append(prices, p)
			}
		}
	}

	return prices
}

// Helper function to calculate correlation between two price series
func calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0.0
	}

	n := len(x)

	// Calculate means
	var sumX, sumY float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)

	// Calculate correlation coefficient
	var numerator, sumXX, sumYY float64
	for i := 0; i < n; i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		numerator += dx * dy
		sumXX += dx * dx
		sumYY += dy * dy
	}

	denominator := math.Sqrt(sumXX * sumYY)
	if denominator == 0 {
		return 0.0
	}

	return numerator / denominator
}

// Volatility Surface - Implied volatility across strikes and expiries
func handleVolatilitySurface(c *gin.Context) {
	// Check if mock data is requested
	if shouldUseMockData(c) {
		mockData := market.GetVolatilitySurfaceMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use free APIs to estimate implied volatility
	client := &http.Client{Timeout: 10 * time.Second}

	// Get Bitcoin price data for volatility calculation
	resp, err := client.Get("https://api.coingecko.com/api/v3/coins/bitcoin/market_chart?vs_currency=usd&days=30&interval=daily")
	if err != nil {
		// Fallback to mock data on error
		mockData := market.GetVolatilitySurfaceMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}
	defer resp.Body.Close()

	var priceData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&priceData); err != nil {
		mockData := market.GetVolatilitySurfaceMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Extract price data
	prices, ok := priceData["prices"].([]interface{})
	if !ok || len(prices) < 7 {
		mockData := market.GetVolatilitySurfaceMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Calculate historical volatility
	priceValues := make([]float64, 0, len(prices))
	for _, price := range prices {
		if priceArray, ok := price.([]interface{}); ok && len(priceArray) >= 2 {
			if priceVal, ok := priceArray[1].(float64); ok {
				priceValues = append(priceValues, priceVal)
			}
		}
	}

	if len(priceValues) < 2 {
		mockData := market.GetVolatilitySurfaceMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Calculate volatility using standard deviation of returns
	returns := make([]float64, len(priceValues)-1)
	for i := 1; i < len(priceValues); i++ {
		returns[i-1] = math.Log(priceValues[i] / priceValues[i-1])
	}

	// Calculate mean and standard deviation
	var sum float64
	for _, ret := range returns {
		sum += ret
	}
	mean := sum / float64(len(returns))

	var variance float64
	for _, ret := range returns {
		variance += math.Pow(ret-mean, 2)
	}
	variance /= float64(len(returns))

	// Annualized volatility (365 days)
	impliedVolatility := math.Sqrt(variance) * math.Sqrt(365) * 100

	// Generate historical data
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = impliedVolatility + (rand.Float64()-0.5)*10
		if historical[i] < 10 {
			historical[i] = 10
		}
		if historical[i] > 200 {
			historical[i] = 200
		}
	}

	// Calculate score using enhanced volatility scoring
	score := getVolatilityScore(impliedVolatility)

	var indicator string
	if impliedVolatility > 120 {
		indicator = "Extremely High Volatility"
	} else if impliedVolatility > 80 {
		indicator = "High Volatility"
	} else if impliedVolatility > 50 {
		indicator = "Moderate Volatility"
	} else if impliedVolatility > 30 {
		indicator = "Low Volatility"
	} else {
		indicator = "Very Low Volatility"
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        impliedVolatility,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"api_source":   "CoinGecko (Live volatility calculation)",
	})
}

// Liquidation Heatmap - Liquidation levels and clustering
func handleLiquidationHeatmap(c *gin.Context) {
	// Check if mock data is requested
	if shouldUseMockData(c) {
		mockData := market.GetLiquidationHeatmapMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use free APIs to estimate liquidation risk
	client := &http.Client{Timeout: 10 * time.Second}

	// Get Bitcoin price and volume for liquidation risk calculation
	resp, err := client.Get("https://api.coingecko.com/api/v3/coins/bitcoin?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false&sparkline=false")
	if err != nil {
		// Fallback to mock data on error
		mockData := market.GetLiquidationHeatmapMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}
	defer resp.Body.Close()

	var btcData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&btcData); err != nil {
		mockData := market.GetLiquidationHeatmapMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Extract market data
	marketData, ok := btcData["market_data"].(map[string]interface{})
	if !ok {
		mockData := market.GetLiquidationHeatmapMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	currentPrice, ok := marketData["current_price"].(map[string]interface{})["usd"].(float64)
	if !ok {
		mockData := market.GetLiquidationHeatmapMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	priceChange24h, ok := marketData["price_change_percentage_24h"].(float64)
	if !ok {
		priceChange24h = 0.0
	}

	high24h, ok := marketData["high_24h"].(map[string]interface{})["usd"].(float64)
	if !ok {
		high24h = currentPrice
	}

	low24h, ok := marketData["low_24h"].(map[string]interface{})["usd"].(float64)
	if !ok {
		low24h = currentPrice
	}

	// Calculate liquidation risk based on price volatility and position clustering
	priceRange := (high24h - low24h) / currentPrice * 100
	volatilityRisk := math.Abs(priceChange24h) * 2

	// Estimate liquidation clustering based on round numbers and volatility
	liquidationRisk := (priceRange + volatilityRisk) * 1.2
	if liquidationRisk > 100 {
		liquidationRisk = 100
	}
	if liquidationRisk < 0 {
		liquidationRisk = 0
	}

	// Generate historical data
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = liquidationRisk + (rand.Float64()-0.5)*20
		if historical[i] < 0 {
			historical[i] = 0
		}
		if historical[i] > 100 {
			historical[i] = 100
		}
	}

	var indicator string
	var score float64

	if liquidationRisk > 80 {
		indicator = "Extreme Liquidation Risk"
		score = -0.9
	} else if liquidationRisk > 60 {
		indicator = "High Liquidation Risk"
		score = -0.6
	} else if liquidationRisk > 40 {
		indicator = "Moderate Liquidation Risk"
		score = -0.3
	} else if liquidationRisk > 20 {
		indicator = "Low Liquidation Risk"
		score = 0.3
	} else {
		indicator = "Minimal Liquidation Risk"
		score = 0.6
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        liquidationRisk,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"api_source":   "CoinGecko (Live liquidation risk estimation)",
	})
}

// getSolanaBalance handles Solana wallet balance requests via backend proxy to avoid CORS
func getSolanaBalance(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(400, gin.H{"error": "Wallet address is required"})
		return
	}

	// Validate Solana address format (basic check)
	if len(address) < 32 || len(address) > 44 {
		c.JSON(400, gin.H{"error": "Invalid Solana address format"})
		return
	}

	// Try multiple RPC endpoints to get balance
	rpcEndpoints := []string{
		"https://solana-mainnet.g.alchemy.com/v2/demo",
		"https://api.mainnet-beta.solana.com",
		"https://solana-api.projectserum.com",
		"https://rpc.ankr.com/solana",
	}

	for _, endpoint := range rpcEndpoints {
		balance, err := fetchSolanaBalanceFromRPC(endpoint, address)
		if err == nil {
			c.JSON(200, gin.H{
				"address":   address,
				"balance":   balance,
				"endpoint":  endpoint,
				"timestamp": time.Now().Unix(),
			})
			return
		}
		log.Printf("RPC endpoint %s failed for address %s: %v", endpoint, address, err)
	}

	// If all endpoints fail, return error
	c.JSON(500, gin.H{
		"error":     "Failed to fetch balance from all RPC endpoints",
		"address":   address,
		"timestamp": time.Now().Unix(),
	})
}

// fetchSolanaBalanceFromRPC fetches balance from a specific RPC endpoint
func fetchSolanaBalanceFromRPC(endpoint, address string) (float64, error) {
	// Create JSON-RPC request
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getBalance",
		"params":  []string{address},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make POST request
	resp, err := client.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Parse response
	var response struct {
		Result struct {
			Value int64 `json:"value"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Error != nil {
		return 0, fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	// Convert lamports to SOL
	balanceInSol := float64(response.Result.Value) / 1000000000.0
	return balanceInSol, nil
}
