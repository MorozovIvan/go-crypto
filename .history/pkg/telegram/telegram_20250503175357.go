package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-vue/pkg/config"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

type TelegramService struct {
	client   *telegram.Client
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	userAuth bool
	phone    string
	code     string
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
	// Set environment variables for the Telegram client
	os.Setenv("APP_ID", config.GlobalConfig.TelegramAPIID)
	os.Setenv("APP_HASH", config.GlobalConfig.TelegramAPIHash)

	logger, _ := zap.NewDevelopment()

	ctx, cancel := context.WithCancel(context.Background())

	return &TelegramService{
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
		userAuth: false,
	}, nil
}

func (s *TelegramService) SetPhone(phone string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.phone = phone
}

func (s *TelegramService) SetCode(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.code = code
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

// Custom authenticator that implements auth.UserAuthenticator
type noSignUp struct{}

func (c noSignUp) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, fmt.Errorf("not implemented")
}

func (c noSignUp) AcceptTermsOfService(ctx context.Context, tos auth.TermsOfService) error {
	return fmt.Errorf("not implemented")
}

func (s *TelegramService) AuthenticateUser(phone string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store phone number
	s.phone = phone

	// Create a new context for this operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize client if not already initialized
	if s.client == nil {
		options := telegram.Options{
			Logger: s.logger,
		}

		client, err := telegram.ClientFromEnvironment(options)
		if err != nil {
			return fmt.Errorf("failed to create client: %v", err)
		}
		s.client = client
	}

	return s.client.Run(ctx, func(ctx context.Context) error {
		api := s.client.API()

		// Convert API ID to int
		apiID, err := strconv.Atoi(config.GlobalConfig.TelegramAPIID)
		if err != nil {
			return fmt.Errorf("invalid API ID: %v", err)
		}

		// Send code request
		_, err = api.AuthSendCode(ctx, &tg.AuthSendCodeRequest{
			PhoneNumber: phone,
			APIID:       apiID,
			APIHash:     config.GlobalConfig.TelegramAPIHash,
			Settings:    tg.CodeSettings{},
		})
		if err != nil {
			return fmt.Errorf("failed to send code: %v", err)
		}

		s.userAuth = true
		return nil
	})
}

func (s *TelegramService) VerifyCode(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return fmt.Errorf("client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.client.Run(ctx, func(ctx context.Context) error {
		api := s.client.API()

		// Sign in with code
		_, err := api.AuthSignIn(ctx, &tg.AuthSignInRequest{
			PhoneNumber:   s.phone,
			PhoneCodeHash: s.userHash,
			PhoneCode:     code,
		})
		if err != nil {
			return fmt.Errorf("failed to verify code: %v", err)
		}

		return nil
	})
}
