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
	"github.com/gotd/td/tg"
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
	return fmt.Sprintf("https://oauth.telegram.org/auth?bot_id=%s&origin=%s&return_to=%s",
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

	// Initialize client if not already initialized
	if s.client == nil {
		// Create options with system logger
		options := telegram.Options{
			Logger: s.logger,
		}

		// Create client from environment variables
		client, err := telegram.ClientFromEnvironment(options)
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %v", err)
		}
		s.client = client
	}

	var result []map[string]interface{}

	if err := s.client.Run(ctx, func(ctx context.Context) error {
		api := s.client.API()

		// Get all dialogs (chats and channels)
		dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetPeer: &tg.InputPeerEmpty{},
			Limit:      100,
		})
		if err != nil {
			return fmt.Errorf("failed to get dialogs: %v", err)
		}

		switch d := dialogs.(type) {
		case *tg.MessagesDialogs:
			for _, chat := range d.Chats {
				switch c := chat.(type) {
				case *tg.Channel:
					// Get channel full info
					channelFull, err := api.ChannelsGetFullChannel(ctx, &tg.InputChannel{
						ChannelID:  c.ID,
						AccessHash: c.AccessHash,
					})
					if err != nil {
						s.logger.Warn("Failed to get channel full info", zap.Error(err))
						continue
					}

					var description string
					if full, ok := channelFull.FullChat.(*tg.ChannelFull); ok {
						description = full.About
					}

					result = append(result, map[string]interface{}{
						"id":          c.ID,
						"title":       c.Title,
						"username":    c.Username,
						"type":        "channel",
						"members":     c.ParticipantsCount,
						"description": description,
					})
				case *tg.Chat:
					result = append(result, map[string]interface{}{
						"id":          c.ID,
						"title":       c.Title,
						"type":        "group",
						"members":     c.ParticipantsCount,
						"description": "",
					})
				}
			}
		}

		return nil
	}); err != nil {
		if strings.Contains(err.Error(), "AUTH_KEY_UNREGISTERED") || strings.Contains(err.Error(), "API_ID_INVALID") {
			s.client = nil
		}
		return nil, fmt.Errorf("failed to run client: %v", err)
	}

	return result, nil
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
