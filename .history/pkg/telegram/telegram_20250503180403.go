package telegram

import (
	"context"
	"fmt"
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
	hash     string
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
		// Create options with system logger
		options := telegram.Options{
			Logger: s.logger,
		}

		// Create client from environment variables
		client, err := telegram.ClientFromEnvironment(options)
		if err != nil {
			return fmt.Errorf("failed to create client: %v", err)
		}
		s.client = client
	}

	// Create a custom authenticator
	authenticator := &phoneAuthenticator{
		phone: phone,
	}

	return s.client.Run(ctx, func(ctx context.Context) error {
		// Start authentication flow
		flow := auth.NewFlow(
			authenticator,
			auth.SendCodeOptions{},
		)

		if err := flow.Run(ctx, s.client); err != nil {
			return fmt.Errorf("failed to authenticate: %v", err)
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

	// Create a custom authenticator
	authenticator := &phoneAuthenticator{
		phone: s.phone,
		code:  code,
	}

	return s.client.Run(ctx, func(ctx context.Context) error {
		// Start authentication flow
		flow := auth.NewFlow(
			authenticator,
			auth.SendCodeOptions{},
		)

		if err := flow.Run(ctx, s.client); err != nil {
			return fmt.Errorf("failed to verify code: %v", err)
		}

		return nil
	})
}

// phoneAuthenticator implements auth.UserAuthenticator
type phoneAuthenticator struct {
	phone string
	code  string
}

func (a *phoneAuthenticator) Phone(_ context.Context) (string, error) {
	return a.phone, nil
}

func (a *phoneAuthenticator) Password(_ context.Context) (string, error) {
	return "", fmt.Errorf("password authentication not supported")
}

func (a *phoneAuthenticator) Code(_ context.Context) (string, error) {
	return a.code, nil
}

func (a *phoneAuthenticator) AcceptTermsOfService(_ context.Context, _ *tg.HelpTermsOfService) error {
	return nil
}
