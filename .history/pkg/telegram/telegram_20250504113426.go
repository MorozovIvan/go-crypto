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
	"github.com/gotd/td/telegram/auth/srp"
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
	password string
}

func NewTelegramService() (*TelegramService, error) {
	// Set environment variables for the Telegram client
	os.Setenv("APP_ID", config.GlobalConfig.TelegramAPIID)
	os.Setenv("APP_HASH", config.GlobalConfig.TelegramAPIHash)

	logger, _ := zap.NewDevelopment()

	ctx, cancel := context.WithCancel(context.Background())

	// Create options with system logger and session storage
	options := telegram.Options{
		Logger: logger,
		SessionStorage: &telegram.FileSessionStorage{
			Path: "session.json",
		},
	}

	// Create client from environment variables
	client, err := telegram.ClientFromEnvironment(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return &TelegramService{
		client:   client,
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

func (s *TelegramService) SetPassword(password string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.password = password
}

func (s *TelegramService) SetHash(hash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hash = hash
}

func (s *TelegramService) GetUserGroups(userID int64) ([]map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a new context for this operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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

	// Create a new client for this operation
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: s.logger,
		SessionStorage: &telegram.FileSessionStorage{
			Path: "session.json",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	// Create a new context for this operation
	ctx := context.Background()

	return client.Run(ctx, func(ctx context.Context) error {
		api := client.API()

		// Convert API ID to int
		apiID, err := strconv.Atoi(config.GlobalConfig.TelegramAPIID)
		if err != nil {
			return fmt.Errorf("invalid API ID: %v", err)
		}

		// Try to log out first to ensure a clean state
		_, _ = api.AuthLogOut(ctx)

		// Send code request using MTProto API
		sentCode, err := api.AuthSendCode(ctx, &tg.AuthSendCodeRequest{
			PhoneNumber: phone,
			APIID:       apiID,
			APIHash:     config.GlobalConfig.TelegramAPIHash,
			Settings: tg.CodeSettings{
				AllowFlashcall:  false,
				CurrentNumber:   false,
				AllowAppHash:    false,
				AllowMissedCall: false,
				LogoutTokens:    nil,
				Token:           "",
				AppSandbox:      false,
			},
		})
		if err != nil {
			if strings.Contains(err.Error(), "AUTH_RESTART") {
				// If we get AUTH_RESTART, try to resend the code
				sentCode, err = api.AuthResendCode(ctx, &tg.AuthResendCodeRequest{
					PhoneNumber:   phone,
					PhoneCodeHash: s.hash,
				})
				if err != nil {
					return fmt.Errorf("failed to resend code: %v", err)
				}
			} else {
				return fmt.Errorf("failed to send code: %v", err)
			}
		}

		// Store the phone code hash
		if code, ok := sentCode.(*tg.AuthSentCode); ok {
			s.hash = code.PhoneCodeHash
			s.logger.Info("Stored phone code hash", zap.String("hash", s.hash))
		} else {
			return fmt.Errorf("unexpected response type from AuthSendCode")
		}

		s.userAuth = true
		return nil
	})
}

// GetPhoneCodeHash returns the stored phone code hash
func (s *TelegramService) GetPhoneCodeHash() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hash
}

func (s *TelegramService) VerifyCode(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Starting code verification",
		zap.String("phone", s.phone),
		zap.String("code", code),
		zap.String("hash", s.hash))

	// Validate required fields
	if s.phone == "" {
		return fmt.Errorf("phone number is required")
	}
	if s.hash == "" {
		return fmt.Errorf("phone code hash is required")
	}
	if code == "" {
		return fmt.Errorf("verification code is required")
	}

	// Store the code
	s.code = code

	// Create a new client for this operation
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: s.logger,
		SessionStorage: &telegram.FileSessionStorage{
			Path: "session.json",
		},
	})
	if err != nil {
		s.logger.Error("Failed to create client", zap.Error(err))
		return fmt.Errorf("failed to create client: %v", err)
	}

	// Create a new context for this operation
	ctx := context.Background()

	return client.Run(ctx, func(ctx context.Context) error {
		api := client.API()

		s.logger.Info("Attempting to sign in with code",
			zap.String("phone", s.phone),
			zap.String("code", s.code),
			zap.String("hash", s.hash))

		// Sign in with code
		auth, err := api.AuthSignIn(ctx, &tg.AuthSignInRequest{
			PhoneNumber:   s.phone,
			PhoneCodeHash: s.hash,
			PhoneCode:     s.code,
		})
		if err != nil {
			s.logger.Error("Failed to sign in", zap.Error(err))
			// Check if 2FA is required
			if strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED") {
				s.logger.Info("2FA password required")

				// Use stored password or prompt for it
				password := s.password
				if password == "" {
					password = prompt("Please enter your 2FA password:")
					if password == "" {
						return fmt.Errorf("2FA password is required")
					}
				}

				s.logger.Info("Attempting to sign in with 2FA password")

				// Get password settings
				settings, err := api.AccountGetPassword(ctx)
				if err != nil {
					s.logger.Error("Failed to get password settings", zap.Error(err))
					return fmt.Errorf("failed to get password settings: %v", err)
				}

				// Create SRP instance
				srpInstance := srp.NewSRP(nil)

				// Create input for SRP
				input := srp.Input{
					Salt1: settings.CurrentSalt,
					Salt2: settings.NewAlgo.Salt2,
					G:     settings.NewAlgo.G,
					P:     settings.NewAlgo.P,
				}

				// Calculate SRP answer
				answer, err := srpInstance.Hash([]byte(password), settings.SrpB, nil, input)
				if err != nil {
					s.logger.Error("Failed to calculate SRP hash", zap.Error(err))
					return fmt.Errorf("failed to calculate SRP hash: %v", err)
				}

				// Sign in with password
				_, err = api.AuthCheckPassword(ctx, &tg.InputCheckPasswordSRP{
					SRPID: settings.SRPID,
					A:     answer.A,
					M1:    answer.M1,
				})
				if err != nil {
					s.logger.Error("Failed to verify 2FA password", zap.Error(err))
					return fmt.Errorf("failed to verify 2FA password: %v", err)
				}
				return nil
			}
			return fmt.Errorf("failed to verify code: %v", err)
		}

		s.logger.Info("Successfully signed in")
		return nil
	})
}

// Helper function to prompt for input
func prompt(message string) string {
	fmt.Print(message + " ")
	var input string
	fmt.Scanln(&input)
	return input
}

func (s *TelegramService) GenerateAuthLink() string {
	return "http://localhost:8080/api/telegram/auth/callback"
}
