package telegram

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-vue/pkg/config"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type TelegramService struct {
	client              *telegram.Client
	logger              *zap.Logger
	ctx                 context.Context
	cancel              context.CancelFunc
	mu                  sync.Mutex
	userAuth            bool
	phone               string
	code                string
	hash                string
	password            string
	lastPasswordAttempt time.Time
	lastCodeAttempt     time.Time
	userID              int64
	clientReady         chan struct{}
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func NewTelegramService() (*TelegramService, error) {
	// Set environment variables for the Telegram client
	os.Setenv("APP_ID", config.GlobalConfig.TelegramAPIID)
	os.Setenv("APP_HASH", config.GlobalConfig.TelegramAPIHash)

	// Create a file logger
	logFile, err := os.OpenFile("telegram_service.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// Create a multi-writer to log to both file and console
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Create logger with custom encoder config
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Create core with file output
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(multiWriter),
		zapcore.DebugLevel,
	)

	logger := zap.New(core)

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

	service := &TelegramService{
		client:      client,
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		userAuth:    false,
		phone:       config.GlobalConfig.DefaultPhoneNumber,
		clientReady: make(chan struct{}),
	}

	// Start the client in a goroutine
	go func() {
		err := client.Run(ctx, func(ctx context.Context) error {
			// Signal that the client is ready
			close(service.clientReady)
			// Keep the client running
			<-ctx.Done()
			return nil
		})
		if err != nil {
			logger.Error("Client run error", zap.Error(err))
		}
	}()

	return service, nil
}

// GetPhone returns the current phone number
func (s *TelegramService) GetPhone() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.phone
}

// SetPhone sets a new phone number
func (s *TelegramService) SetPhone(phone string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if phone == "" {
		s.phone = config.GlobalConfig.DefaultPhoneNumber
	} else {
		s.phone = phone
	}
}

func (s *TelegramService) SetCode(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.code = code
}

func (s *TelegramService) SetPassword(password string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add detailed password debugging
	s.logger.Info("Setting 2FA password",
		zap.Int("length", len(password)),
		zap.Bool("hasSpaces", strings.Contains(password, " ")),
		zap.Bool("hasSpecialChars", strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?")),
		zap.String("firstChar", string(password[0])),
		zap.String("lastChar", string(password[len(password)-1])),
		zap.Bool("isEmptyAfterTrim", strings.TrimSpace(password) == ""),
		zap.String("rawPassword", password),
		zap.String("passwordHex", hex.EncodeToString([]byte(password))),
		zap.Any("passwordBytes", []byte(password)),
	)

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

	// Use the existing client if available
	client := s.client
	if client == nil {
		var err error
		client, err = telegram.ClientFromEnvironment(telegram.Options{
			Logger: s.logger,
			SessionStorage: &telegram.FileSessionStorage{
				Path: "session.json",
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %v", err)
		}
	}

	var result []map[string]interface{}

	// Create a new context for this operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := client.Run(ctx, func(ctx context.Context) error {
		api := client.API()

		// First, check if we're authorized
		_, err := api.AuthExportAuthorization(ctx, 2) // Use DC ID 2 (default)
		if err != nil {
			if strings.Contains(err.Error(), "AUTH_KEY_UNREGISTERED") {
				// Clear the session file if authentication is invalid
				os.Remove("session.json")
				s.client = nil // Clear the client
				return fmt.Errorf("session expired, please re-authenticate")
			}
			return fmt.Errorf("failed to check authorization: %v", err)
		}

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
	})

	if err != nil {
		if strings.Contains(err.Error(), "AUTH_KEY_UNREGISTERED") {
			// Clear the session file if authentication is invalid
			os.Remove("session.json")
			s.client = nil // Clear the client
			return nil, fmt.Errorf("session expired, please re-authenticate")
		}
		return nil, fmt.Errorf("failed to run client: %v", err)
	}

	return result, nil
}

// formatError formats Telegram API errors into user-friendly messages
func (s *TelegramService) formatError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "PHONE_PASSWORD_FLOOD"):
		// Extract wait time from error message
		waitTime := "24 hours"
		if strings.Contains(errStr, "FLOOD_WAIT_") {
			parts := strings.Split(errStr, "FLOOD_WAIT_")
			if len(parts) > 1 {
				seconds := parts[1]
				if secs, err := strconv.Atoi(seconds); err == nil {
					minutes := secs / 60
					remainingSecs := secs % 60
					waitTime = fmt.Sprintf("%d minutes and %d seconds", minutes, remainingSecs)
				} else {
					waitTime = seconds
				}
			}
		}
		return fmt.Errorf("too many failed attempts. Please wait %s before trying again. This is a security measure to protect your account", waitTime)
	case strings.Contains(errStr, "FLOOD_WAIT"):
		waitTime := "24 hours"
		if strings.Contains(errStr, "FLOOD_WAIT_") {
			parts := strings.Split(errStr, "FLOOD_WAIT_")
			if len(parts) > 1 {
				seconds := parts[1]
				if secs, err := strconv.Atoi(seconds); err == nil {
					minutes := secs / 60
					remainingSecs := secs % 60
					waitTime = fmt.Sprintf("%d minutes and %d seconds", minutes, remainingSecs)
				} else {
					waitTime = seconds
				}
			}
		}
		return fmt.Errorf("too many attempts. Please wait %s before trying again. This is a security measure to protect your account", waitTime)
	case strings.Contains(errStr, "PHONE_NUMBER_INVALID"):
		return fmt.Errorf("invalid phone number format. Please check the number and try again. Make sure to include the country code (e.g., +1 for US)")
	case strings.Contains(errStr, "PHONE_CODE_INVALID"):
		return fmt.Errorf("invalid verification code. Please check the code and try again. Make sure you're using the most recent code sent to your phone")
	case strings.Contains(errStr, "PHONE_CODE_EXPIRED"):
		return fmt.Errorf("verification code has expired. Please request a new code by clicking the 'Send Code' button again")
	case strings.Contains(errStr, "PASSWORD_HASH_INVALID"):
		return fmt.Errorf("incorrect 2FA password. Please check your password and try again. If you've forgotten your password, you can reset it in the official Telegram app")
	case strings.Contains(errStr, "SESSION_PASSWORD_NEEDED"):
		return fmt.Errorf("2FA password required. Please enter your 2FA password to continue. If you haven't set up 2FA, please do so in the official Telegram app first")
	case strings.Contains(errStr, "AUTH_RESTART"):
		return fmt.Errorf("authentication session expired. Please start the authentication process again")
	case strings.Contains(errStr, "PASSWORD_REQUIRED"):
		return fmt.Errorf("2FA password is required. Please enter your 2FA password to continue")
	case strings.Contains(errStr, "PASSWORD_INVALID"):
		return fmt.Errorf("incorrect 2FA password. Please check your password and try again. If you've forgotten your password, you can reset it in the official Telegram app")
	default:
		return fmt.Errorf("authentication failed: %v", err)
	}
}

func (s *TelegramService) GetCurrentUserID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a new client for this operation
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: s.logger,
		SessionStorage: &telegram.FileSessionStorage{
			Path: "session.json",
		},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create client: %v", err)
	}

	var userID int64

	// Create a new context for this operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Run(ctx, func(ctx context.Context) error {
		api := client.API()

		// First, check if we're authorized
		_, err := api.AuthExportAuthorization(ctx, 2) // Use DC ID 2 (default)
		if err != nil {
			if strings.Contains(err.Error(), "AUTH_KEY_UNREGISTERED") {
				// Clear the session file if authentication is invalid
				os.Remove("session.json")
				s.client = nil // Clear the client
				return fmt.Errorf("session expired, please re-authenticate")
			}
			return fmt.Errorf("failed to check authorization: %v", err)
		}

		// Get current user
		user, err := api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})
		if err != nil {
			s.logger.Error("Failed to get current user", zap.Error(err))
			return fmt.Errorf("failed to get current user: %v", err)
		}

		if len(user) == 0 {
			s.logger.Error("No user found in response")
			return fmt.Errorf("no user found")
		}

		// Get user ID from the response
		switch u := user[0].(type) {
		case *tg.User:
			if u.ID == 0 {
				s.logger.Error("Invalid user ID received", zap.Any("user", u))
				return fmt.Errorf("invalid user ID received")
			}
			userID = u.ID
			s.logger.Info("Got user ID",
				zap.Int64("userID", userID),
				zap.String("username", u.Username),
				zap.String("firstName", u.FirstName),
				zap.String("lastName", u.LastName),
			)
		default:
			s.logger.Error("Unexpected user type", zap.String("type", fmt.Sprintf("%T", user[0])))
			return fmt.Errorf("unexpected user type: %T", user[0])
		}

		return nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "session expired") {
			// Clear the session file if authentication is invalid
			os.Remove("session.json")
			s.client = nil // Clear the client
			return 0, fmt.Errorf("session expired, please re-authenticate")
		}
		s.logger.Error("Failed to get user ID", zap.Error(err))
		return 0, fmt.Errorf("failed to get user ID: %v", err)
	}

	if userID == 0 {
		s.logger.Error("Invalid user ID received after successful retrieval")
		return 0, fmt.Errorf("invalid user ID received")
	}

	return userID, nil
}

func (s *TelegramService) AuthenticateUser(phone string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store phone number
	s.phone = phone
	s.logger.Info("Starting authentication process", zap.String("phone", phone))

	// Delete existing session file to ensure fresh authentication
	if err := os.Remove("session.json"); err != nil && !os.IsNotExist(err) {
		s.logger.Warn("Failed to remove existing session file", zap.Error(err))
	}

	// Create a new client for this operation
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: s.logger,
		SessionStorage: &telegram.FileSessionStorage{
			Path: "session.json",
		},
	})
	if err != nil {
		return s.formatError(fmt.Errorf("failed to create client: %v", err))
	}

	// Create a new context for this operation
	ctx := context.Background()

	return client.Run(ctx, func(ctx context.Context) error {
		api := client.API()

		// Convert API ID to int
		apiID, err := strconv.Atoi(config.GlobalConfig.TelegramAPIID)
		if err != nil {
			return s.formatError(fmt.Errorf("invalid API ID: %v", err))
		}

		// Try to log out first to ensure a clean state
		s.logger.Info("Logging out any existing session")
		_, _ = api.AuthLogOut(ctx)

		// Send code request using MTProto API
		s.logger.Info("Sending code request",
			zap.String("phone", phone),
			zap.Int("apiID", apiID),
		)

		// First try to cancel any existing code request
		_, _ = api.AuthCancelCode(ctx, &tg.AuthCancelCodeRequest{
			PhoneNumber:   phone,
			PhoneCodeHash: s.hash,
		})

		// Clear the hash before sending new code
		s.hash = ""

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
			s.logger.Error("Failed to send code", zap.Error(err))
			// Handle specific error cases
			switch {
			case strings.Contains(err.Error(), "AUTH_RESTART"):
				s.logger.Info("Auth restart required, resending code")
				// If we get AUTH_RESTART, try to resend the code
				sentCode, err = api.AuthResendCode(ctx, &tg.AuthResendCodeRequest{
					PhoneNumber:   phone,
					PhoneCodeHash: s.hash,
				})
				if err != nil {
					s.logger.Error("Failed to resend code", zap.Error(err))
					return s.formatError(err)
				}
			default:
				return s.formatError(err)
			}
		}

		// Store the phone code hash
		if code, ok := sentCode.(*tg.AuthSentCode); ok {
			s.hash = code.PhoneCodeHash
			s.logger.Info("Code sent successfully",
				zap.String("hash", s.hash),
				zap.String("type", fmt.Sprintf("%T", code.Type)),
			)
		} else {
			s.logger.Error("Unexpected response type from AuthSendCode",
				zap.String("type", fmt.Sprintf("%T", sentCode)),
			)
			return s.formatError(fmt.Errorf("unexpected response type from AuthSendCode"))
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

	// Check if we need to wait before trying again
	if !s.lastCodeAttempt.IsZero() {
		timeSinceLastAttempt := time.Since(s.lastCodeAttempt)
		if timeSinceLastAttempt < 30*time.Second {
			waitTime := 30*time.Second - timeSinceLastAttempt
			s.logger.Info("Rate limiting code attempt",
				zap.Duration("timeSinceLastAttempt", timeSinceLastAttempt),
				zap.Duration("waitTime", waitTime),
			)
			return fmt.Errorf("please wait %v before trying again", waitTime.Round(time.Second))
		}
	}

	// Update last code attempt time
	s.lastCodeAttempt = time.Now()

	// Wait for client to be ready
	select {
	case <-s.clientReady:
		// Client is ready
	case <-time.After(10 * time.Second):
		return fmt.Errorf("client initialization timeout")
	}

	// Create a new context for this operation
	ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
	defer cancel()

	// Create a channel to signal completion
	done := make(chan error, 1)

	// Run the operation in a goroutine
	go func() {
		api := s.client.API()

		// Sign in with code
		auth, err := api.AuthSignIn(ctx, &tg.AuthSignInRequest{
			PhoneNumber:   s.phone,
			PhoneCodeHash: s.hash,
			PhoneCode:     code,
		})

		if err != nil {
			s.logger.Error("Auth error", zap.Error(err))

			// Check for specific error types
			if strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED") {
				s.logger.Info("2FA password required")
				// Don't close the client when 2FA is required
				done <- fmt.Errorf("SESSION_PASSWORD_NEEDED")
				return
			}
			if strings.Contains(err.Error(), "PHONE_CODE_EXPIRED") {
				s.logger.Info("Code expired, requesting new code")

				// Convert API ID to int
				apiID, err := strconv.Atoi(config.GlobalConfig.TelegramAPIID)
				if err != nil {
					s.logger.Error("Invalid API ID", zap.Error(err))
					done <- fmt.Errorf("invalid API ID: %v", err)
					return
				}

				// Request a completely new code
				sentCode, err := api.AuthSendCode(ctx, &tg.AuthSendCodeRequest{
					PhoneNumber: s.phone,
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
					s.logger.Error("Failed to send new code", zap.Error(err))
					done <- fmt.Errorf("code expired and failed to request new code: %v", err)
					return
				}

				// Store the new phone code hash
				if code, ok := sentCode.(*tg.AuthSentCode); ok {
					s.hash = code.PhoneCodeHash
					s.logger.Info("New code sent successfully",
						zap.String("hash", s.hash),
						zap.String("type", fmt.Sprintf("%T", code.Type)),
					)
					// Return a special error that won't hide the form
					done <- fmt.Errorf("CODE_EXPIRED_NEW_SENT")
					return
				} else {
					s.logger.Error("Unexpected response type from AuthSendCode",
						zap.String("type", fmt.Sprintf("%T", sentCode)),
					)
					done <- fmt.Errorf("unexpected response type from AuthSendCode")
					return
				}
			}
			if strings.Contains(err.Error(), "PHONE_PASSWORD_FLOOD") {
				done <- fmt.Errorf("too many attempts. Please wait 30 seconds before trying again")
				return
			}
			if strings.Contains(err.Error(), "PASSWORD_HASH_INVALID") {
				done <- fmt.Errorf("incorrect 2FA password. Please check your password and try again")
				return
			}
			if strings.Contains(err.Error(), "PASSWORD_INVALID") {
				done <- fmt.Errorf("incorrect 2FA password. Please check your password and try again")
				return
			}

			// Clean up the error message
			errMsg := err.Error()
			if strings.HasPrefix(errMsg, "callback: ") {
				errMsg = strings.TrimPrefix(errMsg, "callback: ")
			}
			done <- fmt.Errorf("%s", errMsg)
			return
		}

		// Check authentication result
		if auth == nil {
			done <- fmt.Errorf("authentication failed: no response")
			return
		}

		switch auth.(type) {
		case *tg.AuthAuthorization:
			s.logger.Info("Authentication successful without 2FA")
			done <- nil
		default:
			s.logger.Info("Unexpected auth response type", zap.String("type", fmt.Sprintf("%T", auth)))
			done <- fmt.Errorf("unexpected authentication response")
		}
	}()

	// Wait for either completion or timeout
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("code verification timed out")
	}
}

// Helper functions for SRP calculations
func (s *TelegramService) calculatePasswordHash(salt1, password, salt2 []byte) *big.Int {
	// First hash: H(salt1 + password + salt2)
	h1 := sha256.New()
	h1.Write(salt1)
	h1.Write(password)
	h1.Write(salt2)
	hash1 := h1.Sum(nil)

	// Second hash: H(hash1)
	h2 := sha256.New()
	h2.Write(hash1)
	hash2 := h2.Sum(nil)

	s.logger.Debug("Password hash calculation",
		zap.Int("salt1Length", len(salt1)),
		zap.Int("passwordLength", len(password)),
		zap.Int("salt2Length", len(salt2)),
		zap.Int("hash1Length", len(hash1)),
		zap.Int("hash2Length", len(hash2)),
		zap.String("salt1Hex", hex.EncodeToString(salt1)),
		zap.String("passwordHex", hex.EncodeToString(password)),
		zap.String("salt2Hex", hex.EncodeToString(salt2)),
		zap.String("hash1Hex", hex.EncodeToString(hash1)),
		zap.String("hash2Hex", hex.EncodeToString(hash2)),
		zap.String("rawPassword", string(password)),
		zap.Any("passwordBytes", password),
	)

	return new(big.Int).SetBytes(hash2)
}

func (s *TelegramService) calculateK(p, g *big.Int) *big.Int {
	h := sha256.New()
	h.Write(p.Bytes())
	h.Write(g.Bytes())
	return new(big.Int).SetBytes(h.Sum(nil))
}

func (s *TelegramService) calculateU(A, B *big.Int) *big.Int {
	h := sha256.New()
	h.Write(A.Bytes())
	h.Write(B.Bytes())
	return new(big.Int).SetBytes(h.Sum(nil))
}

func (s *TelegramService) calculateM1(p, g *big.Int, salt1 []byte, A, B, S *big.Int) []byte {
	// Calculate H(p)
	h := sha256.New()
	h.Write(p.Bytes())
	hp := h.Sum(nil)

	// Calculate H(g)
	h.Reset()
	h.Write(g.Bytes())
	hg := h.Sum(nil)

	// Calculate H(p) xor H(g)
	ngxor := make([]byte, len(hp))
	for i := range hp {
		ngxor[i] = hp[i] ^ hg[i]
	}

	// Calculate H(salt1)
	h.Reset()
	h.Write(salt1)
	hsalt1 := h.Sum(nil)

	// Calculate final M1
	h.Reset()
	h.Write(ngxor)
	h.Write(hsalt1)
	h.Write(A.Bytes())
	h.Write(B.Bytes())
	h.Write(S.Bytes())

	return h.Sum(nil)
}

func (s *TelegramService) GenerateAuthLink() string {
	return "http://localhost:8080/api/telegram/auth/callback"
}

func (s *TelegramService) GetCurrentUser() (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a new client for this operation
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: s.logger,
		SessionStorage: &telegram.FileSessionStorage{
			Path: "session.json",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	var user *User

	// Create a new context for this operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Run(ctx, func(ctx context.Context) error {
		api := client.API()

		// First, check if we're authorized
		_, err := api.AuthExportAuthorization(ctx, 2) // Use DC ID 2 (default)
		if err != nil {
			if strings.Contains(err.Error(), "AUTH_KEY_UNREGISTERED") {
				return fmt.Errorf("session expired, please re-authenticate")
			}
			return fmt.Errorf("failed to check authorization: %v", err)
		}

		// Get current user info
		me, err := api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})
		if err != nil {
			return fmt.Errorf("failed to get current user: %v", err)
		}

		if len(me) == 0 {
			return fmt.Errorf("no user data returned")
		}

		userObj, ok := me[0].(*tg.User)
		if !ok {
			return fmt.Errorf("unexpected user type")
		}

		user = &User{
			ID:        userObj.ID,
			Username:  userObj.Username,
			FirstName: userObj.FirstName,
			LastName:  userObj.LastName,
		}

		return nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "session expired") {
			// Clear the session file if authentication is invalid
			os.Remove("session.json")
			return nil, fmt.Errorf("session expired, please re-authenticate")
		}
		return nil, fmt.Errorf("failed to run client: %v", err)
	}

	return user, nil
}

// Add a new method to resend the code
func (s *TelegramService) ResendCode() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.phone == "" {
		return fmt.Errorf("no phone number set")
	}

	// Create a new client for this operation
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: s.logger,
		SessionStorage: &telegram.FileSessionStorage{
			Path: "session.json",
		},
	})
	if err != nil {
		return s.formatError(fmt.Errorf("failed to create client: %v", err))
	}

	// Create a new context for this operation
	ctx := context.Background()

	return client.Run(ctx, func(ctx context.Context) error {
		api := client.API()

		// Cancel any existing code request
		_, _ = api.AuthCancelCode(ctx, &tg.AuthCancelCodeRequest{
			PhoneNumber:   s.phone,
			PhoneCodeHash: s.hash,
		})

		// Clear the hash
		s.hash = ""

		// Convert API ID to int
		apiID, err := strconv.Atoi(config.GlobalConfig.TelegramAPIID)
		if err != nil {
			return s.formatError(fmt.Errorf("invalid API ID: %v", err))
		}

		// Send new code request
		sentCode, err := api.AuthSendCode(ctx, &tg.AuthSendCodeRequest{
			PhoneNumber: s.phone,
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
			s.logger.Error("Failed to resend code", zap.Error(err))
			return s.formatError(err)
		}

		// Store the new phone code hash
		if code, ok := sentCode.(*tg.AuthSentCode); ok {
			s.hash = code.PhoneCodeHash
			s.logger.Info("New code sent successfully",
				zap.String("hash", s.hash),
				zap.String("type", fmt.Sprintf("%T", code.Type)),
			)
		} else {
			s.logger.Error("Unexpected response type from AuthSendCode",
				zap.String("type", fmt.Sprintf("%T", sentCode)),
			)
			return s.formatError(fmt.Errorf("unexpected response type from AuthSendCode"))
		}

		return nil
	})
}

func (s *TelegramService) Verify2FA(password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if we need to wait before trying again
	if !s.lastPasswordAttempt.IsZero() {
		timeSinceLastAttempt := time.Since(s.lastPasswordAttempt)
		if timeSinceLastAttempt < 30*time.Second {
			waitTime := 30*time.Second - timeSinceLastAttempt
			s.logger.Info("Rate limiting password attempt",
				zap.Duration("timeSinceLastAttempt", timeSinceLastAttempt),
				zap.Duration("waitTime", waitTime),
			)
			return fmt.Errorf("please wait %v before trying again", waitTime.Round(time.Second))
		}
	}

	// Update last password attempt time
	s.lastPasswordAttempt = time.Now()

	// Wait for client to be ready
	select {
	case <-s.clientReady:
		// Client is ready
	case <-time.After(10 * time.Second):
		return fmt.Errorf("client initialization timeout")
	}

	// Create a new context for this operation
	ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
	defer cancel()

	// Create a channel to signal completion
	done := make(chan error, 1)

	// Run the operation in a goroutine
	go func() {
		api := s.client.API()

		// Get password settings
		settings, err := api.AccountGetPassword(ctx)
		if err != nil {
			s.logger.Error("Failed to get password settings", zap.Error(err))
			done <- fmt.Errorf("failed to get password settings: %v", err)
			return
		}

		// Check if 2FA is actually enabled
		if !settings.HasPassword {
			done <- fmt.Errorf("2FA is not enabled for this account")
			return
		}

		// Get the algorithm
		algo, ok := settings.CurrentAlgo.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
		if !ok {
			done <- fmt.Errorf("unsupported password algorithm")
			return
		}

		// Generate random 'a' value (256-bit)
		a := make([]byte, 32)
		if _, err := rand.Read(a); err != nil {
			done <- fmt.Errorf("failed to generate random value: %v", err)
			return
		}

		// Convert algorithm parameters to big.Int
		g := new(big.Int).SetInt64(int64(algo.G))
		p := new(big.Int).SetBytes(algo.P)

		// Calculate A = g^a mod p
		A := new(big.Int).Exp(g, new(big.Int).SetBytes(a), p)

		// Calculate password hash
		x := s.calculatePasswordHash(algo.Salt1, []byte(password), algo.Salt2)

		// Calculate k = H(p, g)
		k := s.calculateK(p, g)

		// Calculate B = k*v + g^b mod p
		B := new(big.Int).SetBytes(settings.SRPB)

		// Calculate u = H(A, B)
		u := s.calculateU(A, B)

		// Calculate S = (B - k*g^x)^(a + u*x) mod p
		gx := new(big.Int).Exp(g, x, p)
		kgx := new(big.Int).Mul(k, gx)
		kgx.Mod(kgx, p)
		diff := new(big.Int).Sub(B, kgx)
		diff.Mod(diff, p)
		ux := new(big.Int).Mul(u, x)
		aaux := new(big.Int).Add(new(big.Int).SetBytes(a), ux)
		S := new(big.Int).Exp(diff, aaux, p)

		// Calculate M1 = H(H(p) xor H(g), H(salt1), A, B, S)
		M1 := s.calculateM1(p, g, algo.Salt1, A, B, S)

		// Create input for SRP
		input := &tg.InputCheckPasswordSRP{
			SRPID: settings.SRPID,
			A:     A.Bytes(),
			M1:    M1,
		}

		// Sign in with password
		auth, err := api.AuthCheckPassword(ctx, input)
		if err != nil {
			s.logger.Error("Failed to verify 2FA password",
				zap.Error(err),
				zap.String("passwordLength", fmt.Sprintf("%d", len(password))),
				zap.String("passwordHex", hex.EncodeToString([]byte(password))),
				zap.String("A", hex.EncodeToString(A.Bytes())),
				zap.String("M1", hex.EncodeToString(M1)),
			)
			done <- fmt.Errorf("failed to verify 2FA password: %v", err)
			return
		}

		// Check authentication result
		if auth == nil {
			done <- fmt.Errorf("2FA verification failed: no response")
			return
		}

		switch auth.(type) {
		case *tg.AuthAuthorization:
			s.logger.Info("2FA verification successful")
			done <- nil
		default:
			s.logger.Info("Unexpected auth response type after 2FA", zap.String("type", fmt.Sprintf("%T", auth)))
			done <- fmt.Errorf("unexpected authentication response after 2FA")
		}
	}()

	// Wait for either completion or timeout
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("2FA verification timed out")
	}
}
