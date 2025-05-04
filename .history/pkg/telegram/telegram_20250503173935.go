package telegram

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"go-vue/pkg/config"

	"github.com/gotd/td/telegram"
	"go.uber.org/zap"
)

type TelegramService struct {
	botToken string
	client   *telegram.Client
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	userAuth bool
	userID   int64
	userHash string
}

type TelegramAuthData struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

type ChatInfo struct {
	ID          int64  `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Username    string `json:"username"`
	Description string `json:"description"`
}

func NewTelegramService() (*TelegramService, error) {
	token := config.GlobalConfig.TelegramBotToken
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is not set")
	}

	// Set environment variables for the Telegram client
	os.Setenv("APP_ID", config.GlobalConfig.TelegramAPIID)
	os.Setenv("APP_HASH", config.GlobalConfig.TelegramAPIHash)

	logger, _ := zap.NewDevelopment()

	ctx, cancel := context.WithCancel(context.Background())

	return &TelegramService{
		botToken: token,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
		userAuth: false,
	}, nil
}

func (s *TelegramService) GenerateAuthLink() string {
	botID := s.botToken[:strings.Index(s.botToken, ":")]
	return fmt.Sprintf("https://oauth.telegram.org/auth?bot_id=%s&origin=%s&return_to=%s&request_access=write",
		botID,
		"http://localhost:8080",
		"http://localhost:8080/api/telegram/auth/callback")
}

func (s *TelegramService) VerifyAuthData(data *TelegramAuthData) bool {
	checkString := fmt.Sprintf("auth_date=%d\nfirst_name=%s\nid=%d\nlast_name=%s\nphoto_url=%s\nusername=%s",
		data.AuthDate,
		data.FirstName,
		data.ID,
		data.LastName,
		data.PhotoURL,
		data.Username,
	)

	secretKey := sha256.Sum256([]byte(s.botToken))
	hmacObj := hmac.New(sha256.New, secretKey[:])
	hmacObj.Write([]byte(checkString))
	hash := hex.EncodeToString(hmacObj.Sum(nil))

	return hash == data.Hash
}

func (s *TelegramService) GetUserGroups(userID int64) ([]map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a new context for this operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use Bot API to get updates
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", s.botToken)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get updates: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result struct {
		OK     bool `json:"ok"`
		Result []struct {
			Message struct {
				Chat struct {
					ID          int64  `json:"id"`
					Type        string `json:"type"`
					Title       string `json:"title"`
					Username    string `json:"username"`
					Description string `json:"description"`
				} `json:"chat"`
			} `json:"message"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("bot API error: %s", string(body))
	}

	// Create a map to store unique chats
	uniqueChats := make(map[int64]map[string]interface{})

	// Process updates to extract unique chats
	for _, update := range result.Result {
		chat := update.Message.Chat
		if chat.Type == "group" || chat.Type == "supergroup" || chat.Type == "channel" {
			if _, exists := uniqueChats[chat.ID]; !exists {
				uniqueChats[chat.ID] = map[string]interface{}{
					"id":          chat.ID,
					"type":        chat.Type,
					"title":       chat.Title,
					"username":    chat.Username,
					"description": chat.Description,
				}
			}
		}
	}

	// Convert map to slice
	var groups []map[string]interface{}
	for _, chat := range uniqueChats {
		groups = append(groups, chat)
	}

	return groups, nil
}

func (s *TelegramService) TestBot() error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", s.botToken)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to test bot: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var result struct {
		OK          bool        `json:"ok"`
		Error       interface{} `json:"error_code,omitempty"`
		Description string      `json:"description,omitempty"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if !result.OK {
		return fmt.Errorf("bot test failed: %s", result.Description)
	}

	return nil
}

func (s *TelegramService) GetBotUsername() string {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", s.botToken)
	resp, err := http.Get(url)
	if err != nil {
		return "unknown"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "unknown"
	}

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "unknown"
	}

	if !result.OK {
		return "unknown"
	}

	return result.Result.Username
}

// New method to set user authentication mode
func (s *TelegramService) SetUserAuth(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userAuth = enabled
	s.client = nil // Force client recreation
}

func (s *TelegramService) SetUserCredentials(userID int64, hash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userID = userID
	s.userHash = hash
	s.userAuth = true
	s.client = nil // Force client recreation
}
