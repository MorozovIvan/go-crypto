package services

import (
	"fmt"
	"go-vue/pkg/market"
	"go-vue/pkg/telegram"
	"os"
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
	// For now, return the same as GetGroups since signal channels are groups
	return ts.GetGroups()
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
