package services

import (
	"fmt"
	"go-vue/pkg/market"
	"go-vue/pkg/telegram"
	"os"
	"strings"
)

// Registry holds all services used by handlers
type Registry struct {
	Market   *MarketService
	Telegram *TelegramService
}

// NewRegistry creates a new service registry with all services initialized
func NewRegistry() *Registry {
	return &Registry{
		Market:   NewMarketService(),
		Telegram: NewTelegramService(),
	}
}

// MarketService wraps the existing market service
type MarketService struct {
	service *market.MarketService
}

// NewMarketService creates a new market service instance
func NewMarketService() *MarketService {
	cmcAPIKey := os.Getenv("CMC_API_KEY")
	return &MarketService{
		service: market.NewMarketService(cmcAPIKey),
	}
}

// TelegramService wraps the existing telegram service
type TelegramService struct {
	service *telegram.TelegramService
}

// NewTelegramService creates a new telegram service instance
func NewTelegramService() *TelegramService {
	service, err := telegram.NewTelegramService()
	if err != nil {
		// Log the error for debugging
		fmt.Printf("Failed to initialize Telegram service: %v\n", err)
		// Handle error gracefully - return service with nil underlying service
		return &TelegramService{service: nil}
	}
	return &TelegramService{
		service: service,
	}
}

// Market service methods
func (ms *MarketService) GetRSI() (float64, []float64, error) {
	return ms.service.GetRSI()
}

func (ms *MarketService) GetGoogleTrends() (float64, []float64, error) {
	return ms.service.GetGoogleTrends()
}

func (ms *MarketService) GetBollingerBands() (float64, []float64, error) {
	return ms.service.GetBollingerBands()
}

func (ms *MarketService) GetMovingAverages() (float64, []float64, error) {
	return ms.service.GetMovingAverages()
}

func (ms *MarketService) GetVolumeTrend() (float64, []float64, error) {
	return ms.service.GetVolumeTrend()
}

func (ms *MarketService) GetAltcoinSeasonIndex() (float64, []float64, error) {
	return ms.service.GetAltcoinSeasonIndex()
}

func (ms *MarketService) GetStablecoinSupplyRatio() (float64, []float64, []string, error) {
	return ms.service.GetStablecoinSupplyRatio()
}

// Telegram service methods
func (ts *TelegramService) GetStatus() map[string]interface{} {
	if ts.service == nil {
		return map[string]interface{}{
			"authenticated": false,
			"phone":         "",
			"error":         "Telegram service not initialized",
		}
	}
	return ts.service.GetStatus()
}

func (ts *TelegramService) GetPhone() string {
	if ts.service == nil {
		return ""
	}
	return ts.service.GetPhone()
}

func (ts *TelegramService) AuthenticateUser(phoneNumber string) error {
	if ts.service == nil {
		return fmt.Errorf("telegram service not initialized")
	}
	return ts.service.AuthenticateUser(phoneNumber)
}

func (ts *TelegramService) GetPhoneCodeHash(phoneNumber string) string {
	if ts.service == nil {
		return ""
	}
	return ts.service.GetPhoneCodeHash(phoneNumber)
}

func (ts *TelegramService) VerifyCode(phone, code, hash string) error {
	if ts.service == nil {
		return fmt.Errorf("telegram service not initialized")
	}
	// The underlying service only takes phone and code, hash is not used
	return ts.service.VerifyCode(phone, code)
}

func (ts *TelegramService) Verify2FA(phone, password string) error {
	if ts.service == nil {
		return fmt.Errorf("telegram service not initialized")
	}
	return ts.service.Verify2FA(phone, password)
}

func (ts *TelegramService) Logout() error {
	if ts.service == nil {
		return fmt.Errorf("telegram service not initialized")
	}
	// Use ClearSessions as logout functionality
	ts.service.ClearSessions()
	return nil
}

func (ts *TelegramService) GetGroups() ([]interface{}, error) {
	if ts.service == nil {
		return nil, fmt.Errorf("telegram service not initialized")
	}
	// Get current user ID first
	userID, err := ts.service.GetCurrentUserID()
	if err != nil {
		return nil, err
	}
	// Get user groups
	groups, err := ts.service.GetUserGroups(userID)
	if err != nil {
		return nil, err
	}
	// Convert to []interface{}
	result := make([]interface{}, len(groups))
	for i, group := range groups {
		result[i] = group
	}
	return result, nil
}

func (ts *TelegramService) GetCurrentUser() (interface{}, error) {
	if ts.service == nil {
		return nil, fmt.Errorf("telegram service not initialized")
	}
	return ts.service.GetCurrentUser()
}

func (ts *TelegramService) GetCurrentUserID() (int64, error) {
	if ts.service == nil {
		return 0, fmt.Errorf("telegram service not initialized")
	}
	return ts.service.GetCurrentUserID()
}

func (ts *TelegramService) SaveAuthState() {
	if ts.service == nil {
		return
	}
	ts.service.SaveAuthState()
}

func (ts *TelegramService) GetSignalChannels() ([]interface{}, error) {
	if ts.service == nil {
		return nil, fmt.Errorf("telegram service not initialized")
	}

	// Get basic groups first
	groups, err := ts.GetGroups()
	if err != nil {
		return nil, err
	}

	// Process groups to create signal channel data structure
	var signalChannels []interface{}

	for _, group := range groups {
		if groupMap, ok := group.(map[string]interface{}); ok {
			title, _ := groupMap["title"].(string)
			titleLower := strings.ToLower(title)

			// Check if this looks like a trading/signal channel
			isTradingChannel := strings.Contains(titleLower, "signal") ||
				strings.Contains(titleLower, "trade") ||
				strings.Contains(titleLower, "alpha") ||
				strings.Contains(titleLower, "gamble") ||
				strings.Contains(titleLower, "pump") ||
				strings.Contains(titleLower, "gem") ||
				strings.Contains(titleLower, "call") ||
				strings.Contains(titleLower, "crypto") ||
				strings.Contains(titleLower, "coin") ||
				strings.Contains(titleLower, "token")

			if isTradingChannel {
				// Create signal channel data structure expected by frontend
				signalChannel := map[string]interface{}{
					"id":           groupMap["id"],
					"title":        title,
					"username":     groupMap["username"],
					"members":      getIntValue(groupMap, "members", 0),
					"signalCount":  10,                     // Placeholder - would need message analysis
					"winRate":      75.5,                   // Placeholder - would need performance analysis
					"totalPnl":     1250.0,                 // Placeholder - would need performance analysis
					"lastActivity": "2024-01-15T10:30:00Z", // Placeholder - would need message analysis
					"contractAddresses": []map[string]interface{}{
						// Placeholder contract data - would need message parsing
						{
							"address":    "So11111111111111111111111111111111111111112",
							"token":      "SOL",
							"blockchain": "solana",
							"is_pumpfun": false,
							"token_type": "standard",
						},
					},
					"status": "active",
				}
				signalChannels = append(signalChannels, signalChannel)
			}
		}
	}

	return signalChannels, nil
}

func (ts *TelegramService) GetSignalChannelsWithProcessing(userID int64, period string) ([]interface{}, error) {
	if ts.service == nil {
		return nil, fmt.Errorf("telegram service not initialized")
	}

	// Get user groups using the underlying service
	groups, err := ts.service.GetUserGroups(userID)
	if err != nil {
		return nil, err
	}

	// Process groups to create signal channel data structure
	var signalChannels []interface{}

	for _, groupMap := range groups {
		title, _ := groupMap["title"].(string)
		titleLower := strings.ToLower(title)

		// Check if this looks like a trading/signal channel
		isTradingChannel := strings.Contains(titleLower, "signal") ||
			strings.Contains(titleLower, "trade") ||
			strings.Contains(titleLower, "alpha") ||
			strings.Contains(titleLower, "gamble") ||
			strings.Contains(titleLower, "pump") ||
			strings.Contains(titleLower, "gem") ||
			strings.Contains(titleLower, "call") ||
			strings.Contains(titleLower, "crypto") ||
			strings.Contains(titleLower, "coin") ||
			strings.Contains(titleLower, "token") ||
			strings.Contains(titleLower, "degen") ||
			strings.Contains(titleLower, "moon") ||
			strings.Contains(titleLower, "ape")

		if isTradingChannel {
			// Get real messages from this channel
			groupID := getInt64Value(groupMap, "id", 0)
			accessHash := getInt64Value(groupMap, "access_hash", 0)

			messages, err := ts.service.GetGroupMessages(groupID, accessHash, 50, period)
			if err != nil {
				fmt.Printf("Failed to get messages for group %d (%s): %v\n", groupID, title, err)
				// Continue with basic channel info even if messages fail
				messages = []map[string]interface{}{}
			}

			// Extract and process contract addresses from messages
			var contractAddresses []map[string]interface{}
			contractMap := make(map[string]map[string]interface{}) // Deduplicate contracts

			for _, msg := range messages {
				if contracts, ok := msg["contracts"].([]map[string]interface{}); ok {
					for _, contract := range contracts {
						if addr, ok := contract["address"].(string); ok {
							if _, exists := contractMap[addr]; !exists {
								// Store contract data
								contractData := map[string]interface{}{
									"address":     addr,
									"token":       contract["token"],
									"blockchain":  contract["blockchain"],
									"is_pumpfun":  isPumpfunToken(addr),
									"token_type":  getTokenType(addr),
									"signal_time": msg["date"],
									"message":     msg["message"],
								}
								contractMap[addr] = contractData
								contractAddresses = append(contractAddresses, contractData)
							}
						}
					}
				}
			}

			// Calculate basic statistics
			signalCount := len(contractAddresses)
			if signalCount == 0 {
				signalCount = len(messages) // Use message count if no contracts found
			}

			// Create signal channel data structure expected by frontend
			signalChannel := map[string]interface{}{
				"id":                fmt.Sprintf("%v", groupMap["id"]),
				"title":             title,
				"username":          groupMap["username"],
				"members":           getIntValue(groupMap, "members", 0),
				"signalCount":       signalCount,
				"winRate":           calculateWinRate(contractAddresses),
				"totalPnl":          calculateTotalPnl(contractAddresses),
				"lastActivity":      getLastActivity(messages),
				"contractAddresses": contractAddresses,
				"status":            "active",
			}
			signalChannels = append(signalChannels, signalChannel)
		}
	}

	return signalChannels, nil
}

// Helper function to safely get int values from map
func getIntValue(m map[string]interface{}, key string, defaultValue int) int {
	if val, ok := m[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
		if floatVal, ok := val.(float64); ok {
			return int(floatVal)
		}
	}
	return defaultValue
}

// Helper function to safely get int64 values from map
func getInt64Value(m map[string]interface{}, key string, defaultValue int64) int64 {
	if val, ok := m[key]; ok {
		if intVal, ok := val.(int64); ok {
			return intVal
		}
		if intVal, ok := val.(int); ok {
			return int64(intVal)
		}
		if floatVal, ok := val.(float64); ok {
			return int64(floatVal)
		}
	}
	return defaultValue
}

// Helper function to check if token is pumpfun
func isPumpfunToken(address string) bool {
	// Simple heuristic - pumpfun tokens are typically 44 characters (base58)
	// and start with certain patterns. This is a simplified check.
	return len(address) == 44 && strings.HasPrefix(address, "pump")
}

// Helper function to get token type
func getTokenType(address string) string {
	if isPumpfunToken(address) {
		return "pumpfun"
	}
	return "standard"
}

// Helper function to calculate win rate (placeholder)
func calculateWinRate(contracts []map[string]interface{}) float64 {
	if len(contracts) == 0 {
		return 0.0
	}
	// Placeholder calculation - in reality would need price analysis
	return 65.0 + float64(len(contracts))*2.5
}

// Helper function to calculate total PnL (placeholder)
func calculateTotalPnl(contracts []map[string]interface{}) float64 {
	if len(contracts) == 0 {
		return 0.0
	}
	// Placeholder calculation - in reality would need price analysis
	return float64(len(contracts)) * 150.0
}

// Helper function to get last activity
func getLastActivity(messages []map[string]interface{}) string {
	if len(messages) == 0 {
		return "2024-01-01T00:00:00Z"
	}
	// Get the most recent message date
	if lastMsg := messages[0]; lastMsg != nil {
		if date, ok := lastMsg["date"].(string); ok {
			return date
		}
	}
	return "2024-01-01T00:00:00Z"
}

func (ts *TelegramService) AnalyzeToken(address string) (interface{}, error) {
	if ts.service == nil {
		return nil, fmt.Errorf("telegram service not initialized")
	}
	// This would need to be implemented in the telegram service
	// For now, return a placeholder
	return map[string]interface{}{
		"address": address,
		"status":  "analysis not implemented yet",
	}, nil
}

func (ts *TelegramService) AnalyzeTokensBulk(groupID string, addresses []string) ([]interface{}, error) {
	if ts.service == nil {
		return nil, fmt.Errorf("telegram service not initialized")
	}
	// This would need to be implemented in the telegram service
	// For now, return a placeholder
	results := make([]interface{}, len(addresses))
	for i, addr := range addresses {
		results[i] = map[string]interface{}{
			"address":  addr,
			"group_id": groupID,
			"status":   "bulk analysis not implemented yet",
		}
	}
	return results, nil
}

// Global service registry instance
var GlobalRegistry *Registry

// InitializeServices initializes the global service registry
func InitializeServices() {
	GlobalRegistry = NewRegistry()
}
