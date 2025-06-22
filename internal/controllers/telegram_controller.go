package controllers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-vue/pkg/telegram"
)

type TelegramController struct {
	telegramService *telegram.TelegramService
}

func NewTelegramController(service *telegram.TelegramService) *TelegramController {
	return &TelegramController{
		telegramService: service,
	}
}

func (c *TelegramController) GetAuthLink(w http.ResponseWriter, r *http.Request) {
	authLink := c.telegramService.GenerateAuthLink()
	response := map[string]string{
		"auth_link": authLink,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *TelegramController) GetUserGroups(w http.ResponseWriter, r *http.Request) {
	// Get the current authenticated user ID instead of using a mock ID
	userID, err := c.telegramService.GetCurrentUserID()
	if err != nil {
		http.Error(w, "Failed to get current user: "+err.Error(), http.StatusUnauthorized)
		return
	}

	groups, err := c.telegramService.GetUserGroups(userID)
	if err != nil {
		http.Error(w, "Failed to get user groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert map[string]interface{} to proper response format
	type GroupResponse struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Username    string `json:"username"`
		Type        string `json:"type"`
		Description string `json:"description"`
	}

	response := make([]GroupResponse, len(groups))
	for i, group := range groups {
		// Handle int64 ID conversion
		groupID := strconv.FormatInt(group["id"].(int64), 10)

		// Handle optional fields
		username := ""
		if user, ok := group["username"].(string); ok {
			username = user
		}

		description := ""
		if desc, ok := group["description"].(string); ok {
			description = desc
		}

		response[i] = GroupResponse{
			ID:          groupID,
			Title:       group["title"].(string),
			Username:    username,
			Type:        group["type"].(string),
			Description: description,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// detectSolanaSignals analyzes text for Solana contract addresses and trading signals
func (c *TelegramController) detectSolanaSignals(text string) []map[string]interface{} {
	var signals []map[string]interface{}

	// Solana contract address patterns
	solanaAddressPattern := regexp.MustCompile(`[A-HJ-NP-Z1-9]{32,44}`)

	// Trading signal patterns
	buyPatterns := regexp.MustCompile(`(?i)(buy|bought|aped|long|entry)`)
	sellPatterns := regexp.MustCompile(`(?i)(sell|sold|exit|short)`)

	// Token price patterns
	pricePattern := regexp.MustCompile(`\$[\d,]+\.?\d*|\$?[\d,]+\.?\d*\s*SOL`)

	// Market cap patterns
	mcPattern := regexp.MustCompile(`(?i)market\s*cap\s*\$?[\d,]+`)

	// Find Solana contract addresses
	addresses := solanaAddressPattern.FindAllString(text, -1)

	for _, address := range addresses {
		// Skip if it looks like a transaction hash (too long) or other non-contract data
		if len(address) < 32 || len(address) > 50 {
			continue
		}

		// Determine signal type
		signalType := "neutral"
		if buyPatterns.MatchString(text) {
			signalType = "buy"
		} else if sellPatterns.MatchString(text) {
			signalType = "sell"
		}

		// Extract price if available
		prices := pricePattern.FindAllString(text, -1)
		var price string
		if len(prices) > 0 {
			price = prices[0]
		}

		// Extract market cap if available
		mcs := mcPattern.FindAllString(text, -1)
		var marketCap string
		if len(mcs) > 0 {
			marketCap = mcs[0]
		}

		// Extract token name (words before contract address)
		words := strings.Fields(text)
		var tokenName string
		for i, word := range words {
			if word == address && i > 0 {
				// Look for token name in previous words
				for j := i - 1; j >= 0 && j > i-5; j-- {
					if len(words[j]) > 1 && !strings.Contains(words[j], "http") {
						tokenName = words[j]
						break
					}
				}
				break
			}
		}

		if tokenName == "" {
			tokenName = "Unknown Token"
		}

		signal := map[string]interface{}{
			"id":              address[:8], // Use first 8 chars as ID
			"token":           tokenName,
			"contractAddress": address,
			"signalType":      signalType,
			"blockchain":      "solana",
			"timestamp":       time.Now().Format(time.RFC3339),
			"price":           price,
			"marketCap":       marketCap,
			"performance":     0.0, // Would need historical data to calculate
		}

		signals = append(signals, signal)
	}

	return signals
}

func (c *TelegramController) GetSignalChannels(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "24h"
	}

	// Get the current authenticated user ID instead of using a mock ID
	userID, err := c.telegramService.GetCurrentUserID()
	if err != nil {
		// If we can't get the current user, return empty channels with error info
		response := map[string]interface{}{
			"channels": []map[string]interface{}{},
			"period":   period,
			"success":  false,
			"error":    "Authentication required: " + err.Error(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	groups, err := c.telegramService.GetUserGroups(userID)
	if err != nil {
		// If we can't get groups, return empty channels with error info
		response := map[string]interface{}{
			"channels": []map[string]interface{}{},
			"period":   period,
			"success":  false,
			"error":    "Failed to get groups: " + err.Error(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Analyze each group for Solana signals
	var signalChannels []map[string]interface{}

	for _, group := range groups {
		// For now, we'll simulate message analysis since getting actual messages
		// requires additional Telegram API calls. In production, you would:
		// 1. Get recent messages from the channel using GetHistory
		// 2. Analyze each message for Solana contract addresses
		// 3. Calculate success rates and performance metrics

		// Simulate analyzing messages for Solana signals
		// Check if this is likely a trading/signal channel based on title/description
		title := strings.ToLower(group["title"].(string))
		description := ""
		if desc, ok := group["description"].(string); ok {
			description = strings.ToLower(desc)
		}

		isTradingChannel := strings.Contains(title, "trade") ||
			strings.Contains(title, "signal") ||
			strings.Contains(title, "alpha") ||
			strings.Contains(title, "call") ||
			strings.Contains(title, "gem") ||
			strings.Contains(title, "pump") ||
			strings.Contains(title, "solana") ||
			strings.Contains(title, "defi") ||
			strings.Contains(title, "gamble") ||
			strings.Contains(description, "trading") ||
			strings.Contains(description, "signals")

		if !isTradingChannel {
			continue
		}

		// Convert group ID to string for consistency
		groupID := strconv.FormatInt(group["id"].(int64), 10)

		// Simulate Solana signals found in this channel
		mockSignals := []map[string]interface{}{
			{
				"id":              "sig_" + groupID + "_1",
				"token":           "SAHUR",
				"contractAddress": "DNrmDMs2czDaAwgzg2BmvM7Jn5ZqA6VN5huRqCrSpump",
				"signalType":      "buy",
				"blockchain":      "solana",
				"timestamp":       time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
				"price":           "$0.0000258",
				"marketCap":       "$25,746",
				"performance":     0.0,
			},
			{
				"id":              "sig_" + groupID + "_2",
				"token":           "WELCOME",
				"contractAddress": "HkQmVAVBkruwEXhqpBjheyRedxqm5exDtBfnJ8Wipump",
				"signalType":      "buy",
				"blockchain":      "solana",
				"timestamp":       time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				"price":           "",
				"marketCap":       "",
				"performance":     0.0,
			},
		}

		// Only include channels that have actual Solana signals
		if len(mockSignals) > 0 {
			// Handle username field which might be empty
			username := ""
			if user, ok := group["username"].(string); ok {
				username = user
			}

			// Handle members field which might be int or int32
			members := 0
			if m, ok := group["members"].(int); ok {
				members = m
			} else if m, ok := group["members"].(int32); ok {
				members = int(m)
			}

			signalChannel := map[string]interface{}{
				"id":            groupID,
				"title":         group["title"].(string),
				"username":      username,
				"members":       members,
				"status":        "active",
				"lastActivity":  time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
				"signalCount":   len(mockSignals),
				"successRate":   75.0 + float64(len(mockSignals)*5), // Mock success rate
				"recentSignals": mockSignals,
			}

			signalChannels = append(signalChannels, signalChannel)
		}
	}

	response := map[string]interface{}{
		"channels": signalChannels,
		"period":   period,
		"success":  true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
