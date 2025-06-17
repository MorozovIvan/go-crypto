package telegram

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
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
	"golang.org/x/crypto/pbkdf2"
)

// Add these constants at the top of the file
const (
	minPasswordAttemptInterval = 30 * time.Second
	maxPasswordAttempts        = 5
	passwordAttemptWindow      = 24 * time.Hour
	floodWaitTime              = 24 * time.Hour // Default flood wait time
)

// Add this struct to track password attempts
type PasswordAttempt struct {
	Timestamp time.Time
	Success   bool
}

// Reinstated AuthSession and sessions map
type AuthSession struct {
	Hash                string
	LastCodeAttempt     time.Time
	LastPasswordAttempt time.Time
	PhoneNumber         string
	CreatedAt           time.Time
	PhoneCodeHash       string
	PasswordAttempts    []PasswordAttempt // Track password attempts
}

type TelegramService struct {
	client   *telegram.Client
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	userAuth bool
	phone    string // Current phone being processed, can be removed if handlers pass it
	// Removed global hash, lastPasswordAttempt, lastCodeAttempt
	userID      int64
	clientReady chan struct{}
	sessions    map[string]*AuthSession // key: phone number
}

func NewTelegramService() (*TelegramService, error) {
	// Validate API credentials
	if config.GlobalConfig.TelegramAPIID == "" || config.GlobalConfig.TelegramAPIHash == "" {
		return nil, fmt.Errorf("Telegram API credentials not configured")
	}

	// Set environment variables for the Telegram client
	os.Setenv("APP_ID", config.GlobalConfig.TelegramAPIID)
	os.Setenv("APP_HASH", config.GlobalConfig.TelegramAPIHash)

	logFile, err := os.OpenFile("telegram_service.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(multiWriter),
		zapcore.DebugLevel,
	)
	logger := zap.New(core)

	maskedHash := config.GlobalConfig.TelegramAPIHash
	if len(maskedHash) > 8 {
		maskedHash = maskedHash[:4] + "..." + maskedHash[len(maskedHash)-4:]
	}
	logger.Info("Initializing Telegram service",
		zap.String("api_id", config.GlobalConfig.TelegramAPIID),
		zap.String("api_hash", maskedHash),
	)

	ctx, cancel := context.WithCancel(context.Background())
	options := telegram.Options{
		Logger: logger,
		SessionStorage: &telegram.FileSessionStorage{
			Path: "session.json",
		},
		Device: telegram.DeviceConfig{
			DeviceModel:   "Desktop",
			SystemVersion: "Windows 10",
			AppVersion:    "1.0.0",
			LangCode:      "en",
		},
	}

	client, err := telegram.ClientFromEnvironment(options)
	if err != nil {
		logger.Error("Failed to create Telegram client", zap.Error(err))
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
		sessions:    make(map[string]*AuthSession),
	}

	// Start the client in a separate goroutine
	go func() {
		logger.Info("Starting Telegram client")
		err := client.Run(ctx, func(ctx context.Context) error {
			// Test the connection
			api := client.API()
			_, err := api.HelpGetNearestDC(ctx)
			if err != nil {
				logger.Error("Failed to connect to Telegram", zap.Error(err))
				return fmt.Errorf("failed to connect to Telegram: %v", err)
			}

			close(service.clientReady)
			logger.Info("Telegram client is ready and connected")
			<-ctx.Done()
			return nil
		})
		if err != nil {
			logger.Error("Client run error", zap.Error(err))
		}
	}()

	// Wait for client to be ready with timeout
	select {
	case <-service.clientReady:
		logger.Info("Telegram client initialized successfully")

		// Check if we have an existing authenticated session
		go service.checkExistingSession()
	case <-time.After(10 * time.Second):
		cancel() // Cancel the context if client doesn't initialize in time
		return nil, fmt.Errorf("client initialization timeout")
	}

	return service, nil
}

// checkExistingSession checks if there's an existing authenticated session and restores the state
func (s *TelegramService) checkExistingSession() {
	s.logger.Info("Checking for existing authenticated session")

	// Check if session file exists
	if _, err := os.Stat("session.json"); os.IsNotExist(err) {
		s.logger.Info("No session file found")
		return
	}

	s.logger.Info("Session file found, checking authentication status")

	// Check if we have a persistent auth state file
	authStateFile := "auth_state.json"
	if authData, err := os.ReadFile(authStateFile); err == nil {
		var authState struct {
			UserAuth bool      `json:"user_auth"`
			UserID   int64     `json:"user_id"`
			Phone    string    `json:"phone"`
			SavedAt  time.Time `json:"saved_at"`
			Version  string    `json:"version"`
		}
		if err := json.Unmarshal(authData, &authState); err == nil {
			// Check if the auth state is too old (older than 7 days)
			if !authState.SavedAt.IsZero() && time.Since(authState.SavedAt) > 7*24*time.Hour {
				s.logger.Warn("Auth state is too old, ignoring",
					zap.Time("saved_at", authState.SavedAt),
					zap.Duration("age", time.Since(authState.SavedAt)))
				os.Remove(authStateFile)
				return
			}

			s.logger.Info("Found persistent auth state",
				zap.Bool("user_auth", authState.UserAuth),
				zap.Int64("user_id", authState.UserID),
				zap.String("phone", authState.Phone),
				zap.Time("saved_at", authState.SavedAt),
				zap.String("version", authState.Version))

			// Restore authentication state
			s.mu.Lock()
			s.userAuth = authState.UserAuth
			s.userID = authState.UserID
			s.phone = authState.Phone
			s.mu.Unlock()

			// Verify the session is still valid by trying to get current user
			if authState.UserAuth {
				go func() {
					// Wait for client to be ready
					select {
					case <-s.clientReady:
						// Give the client some time to fully initialize
						time.Sleep(2 * time.Second)

						// Try to get current user to verify session is still valid
						if user, err := s.GetCurrentUser(); err == nil {
							s.logger.Info("Session restored successfully from persistent state",
								zap.Any("user", user))
						} else {
							s.logger.Warn("Session validation failed, but keeping auth state for retry",
								zap.Error(err))
							// Don't immediately clear auth state - it might be a temporary network issue
							// The GetCurrentUser method will handle clearing auth state if needed
						}
					case <-time.After(15 * time.Second):
						s.logger.Warn("Client not ready for session validation timeout")
					}
				}()
			}
			return
		}
	}

	s.logger.Info("No valid persistent auth state found, user will need to authenticate")
}

// attemptSessionRecovery attempts to recover a session by recreating the client
func (s *TelegramService) attemptSessionRecovery() error {
	s.logger.Info("Attempting session recovery")

	// Check if session file exists
	if _, err := os.Stat("session.json"); os.IsNotExist(err) {
		return fmt.Errorf("no session file to recover from")
	}

	// Try to reinitialize the client to reload the session
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()

	// Wait for current client to close
	time.Sleep(2 * time.Second)

	// Reinitialize client
	s.reinitializeClient()

	// Wait for client to be ready
	select {
	case <-s.clientReady:
		s.logger.Info("Client reinitialized for session recovery")

		// Test the recovered session
		if user, err := s.GetCurrentUser(); err == nil {
			s.logger.Info("Session recovery successful", zap.Any("user", user))
			return nil
		} else {
			s.logger.Warn("Session recovery failed", zap.Error(err))
			return err
		}
	case <-time.After(20 * time.Second):
		s.logger.Error("Session recovery timeout")
		return fmt.Errorf("session recovery timeout")
	}
}

// saveAuthState saves the current authentication state to a persistent file
func (s *TelegramService) saveAuthState() {
	s.mu.Lock()
	authState := struct {
		UserAuth bool      `json:"user_auth"`
		UserID   int64     `json:"user_id"`
		Phone    string    `json:"phone"`
		SavedAt  time.Time `json:"saved_at"`
		Version  string    `json:"version"`
	}{
		UserAuth: s.userAuth,
		UserID:   s.userID,
		Phone:    s.phone,
		SavedAt:  time.Now(),
		Version:  "1.0",
	}
	s.mu.Unlock()

	authData, err := json.Marshal(authState)
	if err != nil {
		s.logger.Error("Failed to marshal auth state", zap.Error(err))
		return
	}

	if err := os.WriteFile("auth_state.json", authData, 0600); err != nil {
		s.logger.Error("Failed to save auth state", zap.Error(err))
		return
	}

	s.logger.Info("Authentication state saved successfully",
		zap.Bool("user_auth", authState.UserAuth),
		zap.Int64("user_id", authState.UserID),
		zap.String("phone", authState.Phone))
}

// ClearSessions clears all session data and resets the service state
func (s *TelegramService) ClearSessions() {
	s.logger.Info("Clearing all sessions and resetting service state")

	// Reset service state
	s.mu.Lock()
	s.userAuth = false
	s.userID = 0
	s.sessions = make(map[string]*AuthSession)

	// Cancel current context to stop the client
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()

	// Remove session file
	if err := os.Remove("session.json"); err != nil && !os.IsNotExist(err) {
		s.logger.Warn("Failed to remove session file", zap.Error(err))
	}

	// Remove auth state file
	if err := os.Remove("auth_state.json"); err != nil && !os.IsNotExist(err) {
		s.logger.Warn("Failed to remove auth state file", zap.Error(err))
	}

	// Wait a moment for the client to fully close
	time.Sleep(1 * time.Second)

	// Reinitialize the service synchronously
	s.reinitializeClient()

	s.logger.Info("Sessions cleared successfully")
}

// reinitializeClient reinitializes the Telegram client after logout
func (s *TelegramService) reinitializeClient() {
	s.logger.Info("Reinitializing Telegram client after logout")

	// Create new context
	ctx, cancel := context.WithCancel(context.Background())

	// Create new client ready channel
	clientReady := make(chan struct{})

	// Create new client
	options := telegram.Options{
		Logger: s.logger,
		SessionStorage: &telegram.FileSessionStorage{
			Path: "session.json",
		},
		Device: telegram.DeviceConfig{
			DeviceModel:   "Desktop",
			SystemVersion: "Windows 10",
			AppVersion:    "1.0.0",
			LangCode:      "en",
		},
	}

	client, err := telegram.ClientFromEnvironment(options)
	if err != nil {
		s.logger.Error("Failed to create new Telegram client", zap.Error(err))
		return
	}

	// Update service state with new client and context
	s.mu.Lock()
	s.ctx = ctx
	s.cancel = cancel
	s.client = client
	s.clientReady = clientReady
	s.mu.Unlock()

	// Start the client in a separate goroutine
	go func() {
		s.logger.Info("Starting new Telegram client")
		err := client.Run(ctx, func(ctx context.Context) error {
			// Test the connection
			api := client.API()
			_, err := api.HelpGetNearestDC(ctx)
			if err != nil {
				s.logger.Error("Failed to connect to Telegram", zap.Error(err))
				return fmt.Errorf("failed to connect to Telegram: %v", err)
			}

			close(clientReady)
			s.logger.Info("New Telegram client is ready and connected")
			<-ctx.Done()
			return nil
		})
		if err != nil {
			s.logger.Error("New client run error", zap.Error(err))
		}
	}()

	// Wait for the client to be ready before returning
	select {
	case <-clientReady:
		s.logger.Info("Client reinitialization completed successfully")
		// Give the client a moment to fully stabilize
		time.Sleep(1 * time.Second)
	case <-time.After(15 * time.Second):
		s.logger.Error("Client reinitialization timeout")
	}
}

// GetPhone, SetPhone, SetCode, SetPassword now operate on the global s.phone, s.code, s.password
// These might need adjustment if we want SetPhone to initiate a session for that phone.
// For now, they set the *current* phone/code/password the service is globally focused on.

func (s *TelegramService) GetPhone() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.phone
}

func (s *TelegramService) SetPhone(phone string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if phone == "" {
		s.phone = config.GlobalConfig.DefaultPhoneNumber
	} else {
		s.phone = phone
	}
	// Ensure a session exists for this phone when it's set
	if _, ok := s.sessions[s.phone]; !ok {
		s.sessions[s.phone] = &AuthSession{}
	}
}

// SetHash now operates on the session for the *current* s.phone
func (s *TelegramService) SetHash(hash string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.phone == "" {
		s.logger.Warn("SetHash called with no phone set in service")
		return
	}
	sess, ok := s.sessions[s.phone]
	if !ok {
		s.logger.Info("SetHash: session not found for phone, creating new one", zap.String("phone", s.phone))
		sess = &AuthSession{}
		s.sessions[s.phone] = sess
	} else {
		s.logger.Info("SetHash: updating hash for existing session", zap.String("phone", s.phone), zap.String("old_hash", sess.Hash), zap.String("new_hash", hash))
	}
	sess.Hash = hash
	s.logger.Info("Hash set for phone", zap.String("phone", s.phone), zap.String("stored_hash", sess.Hash))
}

func (s *TelegramService) GetUserGroups(userID int64) ([]map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("GetUserGroups called", zap.Int64("userID", userID))

	// Use the existing authenticated client instead of creating a new one
	if s.client == nil {
		s.logger.Error("No authenticated client available")
		return nil, fmt.Errorf("no authenticated client available")
	}

	var result []map[string]interface{}

	// Create a new context for this operation
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	// Use the existing client API directly instead of calling Run()
	api := s.client.API()

	s.logger.Info("Attempting to get dialogs from Telegram API")

	// First, check if we're authorized by trying to get dialogs
	dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      100,
	})
	if err != nil {
		s.logger.Error("Failed to get dialogs", zap.Error(err))
		if strings.Contains(err.Error(), "AUTH_KEY_UNREGISTERED") {
			// Clear the session file if authentication is invalid
			os.Remove("session.json")
			s.client = nil // Clear the client
			return nil, fmt.Errorf("session expired, please re-authenticate")
		}
		return nil, fmt.Errorf("failed to get dialogs: %v", err)
	}

	s.logger.Info("Successfully got dialogs", zap.String("type", fmt.Sprintf("%T", dialogs)))

	switch d := dialogs.(type) {
	case *tg.MessagesDialogs:
		s.logger.Info("Processing MessagesDialogs", zap.Int("chat_count", len(d.Chats)))
		for i, chat := range d.Chats {
			s.logger.Info("Processing chat", zap.Int("index", i), zap.String("type", fmt.Sprintf("%T", chat)))
			switch c := chat.(type) {
			case *tg.Channel:
				s.logger.Info("Found channel", zap.Int64("id", c.ID), zap.String("title", c.Title))

				// Get channel full info
				channelFull, err := api.ChannelsGetFullChannel(ctx, &tg.InputChannel{
					ChannelID:  c.ID,
					AccessHash: c.AccessHash,
				})
				if err != nil {
					s.logger.Warn("Failed to get channel full info", zap.Error(err))
					// Continue without full info
				}

				var description string
				if channelFull != nil {
					if full, ok := channelFull.FullChat.(*tg.ChannelFull); ok {
						description = full.About
					}
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
				s.logger.Info("Found chat", zap.Int64("id", c.ID), zap.String("title", c.Title))
				result = append(result, map[string]interface{}{
					"id":          c.ID,
					"title":       c.Title,
					"type":        "group",
					"members":     c.ParticipantsCount,
					"description": "",
				})
			default:
				s.logger.Info("Skipping unknown chat type", zap.String("type", fmt.Sprintf("%T", chat)))
			}
		}
	case *tg.MessagesDialogsSlice:
		s.logger.Info("Processing MessagesDialogsSlice", zap.Int("chat_count", len(d.Chats)))
		for i, chat := range d.Chats {
			s.logger.Info("Processing chat", zap.Int("index", i), zap.String("type", fmt.Sprintf("%T", chat)))
			switch c := chat.(type) {
			case *tg.Channel:
				s.logger.Info("Found channel", zap.Int64("id", c.ID), zap.String("title", c.Title))

				// Get channel full info
				channelFull, err := api.ChannelsGetFullChannel(ctx, &tg.InputChannel{
					ChannelID:  c.ID,
					AccessHash: c.AccessHash,
				})
				if err != nil {
					s.logger.Warn("Failed to get channel full info", zap.Error(err))
					// Continue without full info
				}

				var description string
				if channelFull != nil {
					if full, ok := channelFull.FullChat.(*tg.ChannelFull); ok {
						description = full.About
					}
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
				s.logger.Info("Found chat", zap.Int64("id", c.ID), zap.String("title", c.Title))
				result = append(result, map[string]interface{}{
					"id":          c.ID,
					"title":       c.Title,
					"type":        "group",
					"members":     c.ParticipantsCount,
					"description": "",
				})
			default:
				s.logger.Info("Skipping unknown chat type", zap.String("type", fmt.Sprintf("%T", chat)))
			}
		}
	default:
		s.logger.Warn("Unexpected dialogs type", zap.String("type", fmt.Sprintf("%T", dialogs)))
	}

	s.logger.Info("Successfully retrieved user groups", zap.Int("count", len(result)))
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
					waitTime = formatWaitTime(secs)
				} else {
					waitTime = seconds
				}
			}
		}
		return fmt.Errorf("too many attempts. Please wait %s before trying again. This is a security measure to protect your account", waitTime)
	case strings.Contains(errStr, "FLOOD_WAIT"):
		waitTime := "24 hours"
		if strings.Contains(errStr, "FLOOD_WAIT_") {
			parts := strings.Split(errStr, "FLOOD_WAIT_")
			if len(parts) > 1 {
				seconds := parts[1]
				if secs, err := strconv.Atoi(seconds); err == nil {
					waitTime = formatWaitTime(secs)
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

// Helper function to format wait time
func formatWaitTime(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	} else if seconds < 3600 {
		minutes := seconds / 60
		remainingSecs := seconds % 60
		if remainingSecs == 0 {
			return fmt.Sprintf("%d minutes", minutes)
		}
		return fmt.Sprintf("%d minutes and %d seconds", minutes, remainingSecs)
	} else {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		if minutes == 0 {
			return fmt.Sprintf("%d hours", hours)
		}
		return fmt.Sprintf("%d hours and %d minutes", hours, minutes)
	}
}

func (s *TelegramService) GetCurrentUserID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Use the existing authenticated client instead of creating a new one
	if s.client == nil {
		return 0, fmt.Errorf("no authenticated client available")
	}

	var userID int64

	// Create a new context for this operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := s.client.Run(ctx, func(ctx context.Context) error {
		api := s.client.API()

		// Get current user directly without checking authorization
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
		if strings.Contains(err.Error(), "AUTH_KEY_UNREGISTERED") {
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

func (s *TelegramService) AuthenticateUser(phoneNumber string) error {
	s.logger.Info("Starting authentication process", zap.String("phone", phoneNumber))

	// Clear any existing session for this phone number
	s.mu.Lock()
	delete(s.sessions, phoneNumber)
	s.mu.Unlock()

	// Create a new session
	session := &AuthSession{
		PhoneNumber: phoneNumber,
		CreatedAt:   time.Now(),
	}
	s.mu.Lock()
	s.sessions[phoneNumber] = session
	s.mu.Unlock()

	// Wait for client to be ready
	select {
	case <-s.clientReady:
		s.logger.Info("Client is ready, proceeding with authentication")
	case <-time.After(5 * time.Second):
		return fmt.Errorf("client not ready after timeout")
	}

	// Get the API client
	api := s.client.API()

	// Create a new context for this authentication attempt
	authCtx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	// Convert API ID to integer
	apiID, err := strconv.Atoi(config.GlobalConfig.TelegramAPIID)
	if err != nil {
		s.logger.Error("Failed to convert API ID", zap.Error(err))
		return fmt.Errorf("invalid API ID: %v", err)
	}

	s.logger.Info("Sending code request to Telegram",
		zap.String("phone", phoneNumber),
		zap.Int("api_id", apiID))

	// Send code request using the API directly (reuse the api variable)
	sentCode, err := api.AuthSendCode(authCtx, &tg.AuthSendCodeRequest{
		PhoneNumber: phoneNumber,
		APIID:       apiID,
		APIHash:     config.GlobalConfig.TelegramAPIHash,
		Settings: tg.CodeSettings{
			AllowFlashcall: false,
			CurrentNumber:  true,
			AllowAppHash:   true,
		},
	})
	if err != nil {
		s.logger.Error("Failed to send code",
			zap.String("phone", phoneNumber),
			zap.Error(err))

		// Handle flood wait error
		if strings.Contains(err.Error(), "FLOOD_WAIT_") {
			parts := strings.Split(err.Error(), "FLOOD_WAIT_")
			if len(parts) > 1 {
				seconds := parts[1]
				if secs, err := strconv.Atoi(seconds); err == nil {
					waitTime := formatWaitTime(secs)
					return fmt.Errorf("too many attempts. Please wait %s before trying again", waitTime)
				}
			}
			return fmt.Errorf("too many attempts. Please wait before trying again")
		}

		// If we get AUTH_RESTART, try to clear the session file and retry once
		if strings.Contains(err.Error(), "AUTH_RESTART") {
			s.logger.Info("Received AUTH_RESTART, clearing session and retrying")
			os.Remove("session.json")

			// Retry the code request
			sentCode, err = api.AuthSendCode(authCtx, &tg.AuthSendCodeRequest{
				PhoneNumber: phoneNumber,
				APIID:       apiID,
				APIHash:     config.GlobalConfig.TelegramAPIHash,
				Settings: tg.CodeSettings{
					AllowFlashcall: false,
					CurrentNumber:  true,
					AllowAppHash:   true,
				},
			})
			if err != nil {
				s.logger.Error("Failed to send code on retry",
					zap.String("phone", phoneNumber),
					zap.Error(err))
				return fmt.Errorf("failed to send code: %v", err)
			}
		} else {
			return fmt.Errorf("failed to send code: %v", err)
		}
	}

	// Type assert the response
	code, ok := sentCode.(*tg.AuthSentCode)
	if !ok {
		s.logger.Error("Unexpected response type from AuthSendCode",
			zap.String("type", fmt.Sprintf("%T", sentCode)))
		return fmt.Errorf("unexpected response type from AuthSendCode")
	}

	// Store the phone code hash
	s.mu.Lock()
	session.PhoneCodeHash = code.PhoneCodeHash
	s.mu.Unlock()

	s.logger.Info("Successfully sent verification code",
		zap.String("phone", phoneNumber),
		zap.String("type", code.Type.String()))

	return nil
}

func (s *TelegramService) GetPhoneCodeHash(phoneForHash string) string { // Takes phone as arg
	s.mu.Lock()
	defer s.mu.Unlock()
	if phoneForHash == "" {
		s.logger.Warn("GetPhoneCodeHash called with empty phone")
		return ""
	}
	sess, ok := s.sessions[phoneForHash]
	if !ok {
		s.logger.Warn("GetPhoneCodeHash: no session for phone", zap.String("phone", phoneForHash))
		return ""
	}
	s.logger.Info("Retrieved phone code hash for phone",
		zap.String("phone", phoneForHash),
		zap.String("hash", sess.PhoneCodeHash))
	return sess.PhoneCodeHash
}

// VerifyCode now takes phone as an argument to fetch the correct session
func (s *TelegramService) VerifyCode(phoneForVerify, code string) error {
	// phoneForVerify is the key for the session
	s.mu.Lock()
	sess, ok := s.sessions[phoneForVerify]
	if !ok {
		s.mu.Unlock()
		s.logger.Error("VerifyCode: no session for phone", zap.String("phone", phoneForVerify))
		return fmt.Errorf("no session for phone number %s", phoneForVerify)
	}

	s.logger.Info("Starting code verification for phone",
		zap.String("phone", phoneForVerify),
		zap.String("stored_hash_in_session", sess.PhoneCodeHash),
		zap.String("input_code", code),
	)

	if !sess.LastCodeAttempt.IsZero() {
		timeSinceLastAttempt := time.Since(sess.LastCodeAttempt)
		if timeSinceLastAttempt < 30*time.Second { // Rate limit
			waitTime := 30*time.Second - timeSinceLastAttempt
			s.mu.Unlock()
			s.logger.Info("Rate limiting code attempt for phone", zap.String("phone", phoneForVerify))
			return fmt.Errorf("please wait %v before trying again", waitTime.Round(time.Second))
		}
	}
	sess.LastCodeAttempt = time.Now()
	s.mu.Unlock() // Unlock before blocking calls

	select {
	case <-s.clientReady:
		s.logger.Info("Client is ready for AuthSignIn")
	case <-time.After(10 * time.Second):
		s.logger.Error("Client initialization timeout for AuthSignIn")
		return fmt.Errorf("client initialization timeout")
	}

	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		api := s.client.API()
		// Critical: Use sess.PhoneCodeHash which is specific to phoneForVerify
		auth, err := api.AuthSignIn(ctx, &tg.AuthSignInRequest{
			PhoneNumber:   phoneForVerify,     // Use the phone number passed to this method
			PhoneCodeHash: sess.PhoneCodeHash, // Use hash from the session specific to this phone
			PhoneCode:     code,
		})
		if err != nil {
			s.logger.Error("AuthSignIn error for phone", zap.Error(err), zap.String("phone", phoneForVerify))
			if strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED") {
				done <- fmt.Errorf("SESSION_PASSWORD_NEEDED")
				return
			}
			if strings.Contains(err.Error(), "PHONE_CODE_EXPIRED") {
				s.logger.Info("Code expired for phone, requesting new code", zap.String("phone", phoneForVerify))
				apiID, _ := strconv.Atoi(config.GlobalConfig.TelegramAPIID)
				sentCode, sendErr := api.AuthSendCode(ctx, &tg.AuthSendCodeRequest{
					PhoneNumber: phoneForVerify,
					APIID:       apiID,
					APIHash:     config.GlobalConfig.TelegramAPIHash,
					Settings:    tg.CodeSettings{ /* ... */ },
				})
				if sendErr != nil {
					s.logger.Error("Failed to send new code after expiry for phone", zap.Error(sendErr), zap.String("phone", phoneForVerify))
					done <- fmt.Errorf("code expired, failed to send new: %v", sendErr)
					return
				}
				if newCode, ok := sentCode.(*tg.AuthSentCode); ok {
					s.mu.Lock()
					sess.PhoneCodeHash = newCode.PhoneCodeHash // Update session with new hash
					s.mu.Unlock()
					s.logger.Info("New code sent after expiry for phone, updated session hash", zap.String("phone", phoneForVerify), zap.String("new_hash", sess.PhoneCodeHash))
					done <- fmt.Errorf("CODE_EXPIRED_NEW_SENT:%s", sess.PhoneCodeHash) // Return new hash
					return
				}
				done <- fmt.Errorf("code expired, new code sent, but unexpected response type")
				return
			}
			done <- s.formatError(err) // Format other errors
			return
		}
		if auth == nil {
			s.logger.Error("AuthSignIn returned nil auth result")
			done <- fmt.Errorf("authentication failed: nil result")
			return
		}

		// Handle successful authentication
		if authUser, ok := auth.(*tg.AuthAuthorization); ok {
			if user, ok := authUser.User.(*tg.User); ok {
				s.mu.Lock()
				s.userID = user.ID
				s.userAuth = true
				s.phone = phoneForVerify
				s.mu.Unlock()

				// Save authentication state persistently
				s.saveAuthState()

				s.logger.Info("Successfully authenticated with code",
					zap.Int64("userID", user.ID),
					zap.String("username", user.Username),
					zap.String("firstName", user.FirstName),
					zap.String("lastName", user.LastName),
					zap.String("phone", phoneForVerify),
				)
			} else {
				s.logger.Error("Unexpected user type in auth result", zap.String("type", fmt.Sprintf("%T", authUser.User)))
				done <- fmt.Errorf("unexpected user type in auth result")
				return
			}
		} else {
			s.logger.Error("Unexpected auth result type", zap.String("type", fmt.Sprintf("%T", auth)))
			done <- fmt.Errorf("unexpected auth result type")
			return
		}

		done <- nil // Success
	}()

	select {
	case err := <-done:
		s.logger.Info("Verification attempt completed for phone", zap.String("phone", phoneForVerify), zap.Error(err))
		return err
	case <-ctx.Done():
		return fmt.Errorf("code verification for %s timed out", phoneForVerify)
	}
}

// Verify2FA now also takes phone as an argument
func (s *TelegramService) Verify2FA(phoneForVerify, password string) error {
	s.mu.Lock()
	sess, ok := s.sessions[phoneForVerify]
	if !ok {
		s.mu.Unlock()
		s.logger.Error("Verify2FA: no session for phone", zap.String("phone", phoneForVerify))
		return fmt.Errorf("no session for phone number %s", phoneForVerify)
	}

	// Initialize password attempts if not exists
	if sess.PasswordAttempts == nil {
		sess.PasswordAttempts = make([]PasswordAttempt, 0)
	}

	// Clean up old attempts
	now := time.Now()
	validAttempts := make([]PasswordAttempt, 0)
	for _, attempt := range sess.PasswordAttempts {
		if now.Sub(attempt.Timestamp) <= passwordAttemptWindow {
			validAttempts = append(validAttempts, attempt)
		}
	}
	sess.PasswordAttempts = validAttempts

	// Check if we're rate limited
	if !sess.LastPasswordAttempt.IsZero() {
		timeSinceLastAttempt := time.Since(sess.LastPasswordAttempt)
		if timeSinceLastAttempt < minPasswordAttemptInterval {
			waitTime := minPasswordAttemptInterval - timeSinceLastAttempt
			s.mu.Unlock()
			s.logger.Info("Rate limiting password attempt for phone",
				zap.String("phone", phoneForVerify),
				zap.Duration("wait_time", waitTime),
			)
			return fmt.Errorf("please wait %s before trying again", formatWaitTime(int(waitTime.Seconds())))
		}
	}

	// Check if we've exceeded the maximum number of attempts
	if len(sess.PasswordAttempts) >= maxPasswordAttempts {
		oldestAttempt := sess.PasswordAttempts[0].Timestamp
		timeUntilReset := passwordAttemptWindow - now.Sub(oldestAttempt)
		s.mu.Unlock()
		s.logger.Info("Maximum password attempts reached for phone",
			zap.String("phone", phoneForVerify),
			zap.Duration("time_until_reset", timeUntilReset),
		)
		return fmt.Errorf("too many failed attempts. Please wait %s before trying again", formatWaitTime(int(timeUntilReset.Seconds())))
	}

	// Update last attempt time and add new attempt
	sess.LastPasswordAttempt = now
	sess.PasswordAttempts = append(sess.PasswordAttempts, PasswordAttempt{
		Timestamp: now,
		Success:   false, // Will be updated if successful
	})
	s.mu.Unlock()

	select {
	case <-s.clientReady:
		s.logger.Info("Client is ready for AccountGetPassword")
	case <-time.After(10 * time.Second):
		s.logger.Error("Client initialization timeout for AccountGetPassword")
		return fmt.Errorf("client initialization timeout")
	}

	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		defer close(done)
		api := s.client.API()

		// Get password settings
		passwordSettings, err := api.AccountGetPassword(ctx)
		if err != nil {
			s.logger.Error("Failed to get password settings", zap.Error(err))
			done <- fmt.Errorf("failed to get password settings: %v", err)
			return
		}

		s.logger.Info("Got password settings",
			zap.Int64("srpID", passwordSettings.SRPID),
			zap.String("srpB", hex.EncodeToString(passwordSettings.SRPB)),
		)

		// Type assert the password algorithm
		algo, ok := passwordSettings.CurrentAlgo.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
		if !ok {
			s.logger.Error("Unexpected password algorithm type", zap.String("type", fmt.Sprintf("%T", passwordSettings.CurrentAlgo)))
			done <- fmt.Errorf("unexpected password algorithm type")
			return
		}

		s.logger.Info("Current algorithm parameters",
			zap.String("salt1", hex.EncodeToString(algo.Salt1)),
			zap.String("salt2", hex.EncodeToString(algo.Salt2)),
		)

		// Type assert the new algorithm
		newAlgo, ok := passwordSettings.NewAlgo.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
		if !ok {
			s.logger.Error("Unexpected new algorithm type", zap.String("type", fmt.Sprintf("%T", passwordSettings.NewAlgo)))
			done <- fmt.Errorf("unexpected new algorithm type")
			return
		}

		s.logger.Info("New algorithm parameters",
			zap.String("p", hex.EncodeToString(newAlgo.P)),
			zap.Int("g", newAlgo.G),
		)

		// Convert SRP parameters
		p := new(big.Int).SetBytes(newAlgo.P)
		g := big.NewInt(int64(newAlgo.G))
		srpB := new(big.Int).SetBytes(passwordSettings.SRPB)

		// Calculate x = PH2(password, salt1, salt2)
		x := s.calculatePH2([]byte(password), algo.Salt1, algo.Salt2)
		s.logger.Info("SRP Debug: Calculated x", zap.String("x", x.Text(16)))

		// Calculate v = pow(g, x) mod p
		v := new(big.Int).Exp(g, x, p)
		s.logger.Info("SRP Debug: Calculated v", zap.String("v", v.Text(16)))

		// Calculate k = H(p | g)
		k := s.calculateK(p, g)
		s.logger.Info("SRP Debug: Calculated k", zap.String("k", k.Text(16)))

		// Calculate k_v = (k * v) mod p
		kv := new(big.Int).Mul(k, v)
		kv.Mod(kv, p)
		s.logger.Info("SRP Debug: Calculated k_v", zap.String("k_v", kv.Text(16)))

		// Generate client's private key (a)
		aBytes := make([]byte, 256) // 2048 bits
		_, err = io.ReadFull(rand.Reader, aBytes)
		if err != nil {
			s.logger.Error("Failed to generate random private key", zap.Error(err))
			done <- fmt.Errorf("failed to generate random private key: %v", err)
			return
		}
		a := new(big.Int).SetBytes(aBytes)
		a.Mod(a, p) // Ensure a is in the range [0, p-1]
		s.logger.Info("SRP Debug: Generated client private key", zap.String("a", a.Text(16)))

		// Calculate client's public key (A = g^a mod p)
		A := new(big.Int).Exp(g, a, p)
		s.logger.Info("SRP Debug: Calculated client public key", zap.String("A", A.Text(16)))

		// Calculate u = H(A | B)
		u := s.calculateU(A, srpB)
		s.logger.Info("SRP Debug: Calculated u", zap.String("u", u.Text(16)))

		// Calculate t = (g_b - k_v) mod p
		t := new(big.Int).Sub(srpB, kv)
		if t.Sign() < 0 {
			t.Add(t, p)
		}
		t.Mod(t, p)
		s.logger.Info("SRP Debug: Calculated t", zap.String("t", t.Text(16)))

		// Calculate s_a = pow(t, a + u * x) mod p
		ux := new(big.Int).Mul(u, x)
		aux := new(big.Int).Add(a, ux)
		aux.Mod(aux, p) // Ensure aux is in the range [0, p-1]
		s_a := new(big.Int).Exp(t, aux, p)
		s.logger.Info("SRP Debug: Calculated s_a", zap.String("s_a", s_a.Text(16)))

		// Calculate k_a = H(s_a)
		// Ensure s_a is properly padded to 256 bytes
		s_aBytes := make([]byte, 256)
		s_aBytesPadded := s_a.Bytes()
		copy(s_aBytes[256-len(s_aBytesPadded):], s_aBytesPadded)

		h := sha256.New()
		h.Write(s_aBytes)
		k_a := h.Sum(nil)
		s.logger.Info("SRP Debug: Calculated k_a", zap.String("k_a", hex.EncodeToString(k_a)))

		// Calculate M1 = H(H(p) xor H(g) | H(salt1) | H(salt2) | g_a | g_b | k_a)
		m1 := s.calculateM1(p, g, algo.Salt1, algo.Salt2, A, srpB, k_a)
		s.logger.Info("SRP Debug: Calculated M1", zap.String("m1", hex.EncodeToString(m1)))

		// Ensure A is properly padded to 256 bytes
		ABytes := make([]byte, 256)
		aBytesPadded := A.Bytes()
		copy(ABytes[256-len(aBytesPadded):], aBytesPadded)

		// Sign in with 2FA
		authResult, err := api.AuthCheckPassword(ctx, &tg.InputCheckPasswordSRP{
			SRPID: passwordSettings.SRPID,
			A:     ABytes,
			M1:    m1,
		})
		if err != nil {
			s.logger.Error("2FA verification failed", zap.Error(err))
			if strings.Contains(err.Error(), "PHONE_PASSWORD_FLOOD") {
				// Extract wait time from error if available
				waitTime := floodWaitTime
				if parts := strings.Split(err.Error(), "FLOOD_WAIT_"); len(parts) > 1 {
					if secs, err := strconv.Atoi(parts[1]); err == nil {
						waitTime = time.Duration(secs) * time.Second
					}
				}
				done <- fmt.Errorf("too many attempts. Please wait %s before trying again", formatWaitTime(int(waitTime.Seconds())))
			} else {
				done <- s.formatError(err)
			}
			return
		}

		s.logger.Info("2FA verification successful")

		// Get user ID immediately after successful authentication
		// while we're still in the authenticated context
		if authUser, ok := authResult.(*tg.AuthAuthorization); ok {
			if user, ok := authUser.User.(*tg.User); ok {
				s.mu.Lock()
				s.userID = user.ID
				s.userAuth = true
				s.phone = phoneForVerify
				s.mu.Unlock()

				// Save authentication state persistently
				s.saveAuthState()

				s.logger.Info("Successfully authenticated with 2FA",
					zap.Int64("userID", user.ID),
					zap.String("username", user.Username),
					zap.String("firstName", user.FirstName),
					zap.String("lastName", user.LastName),
					zap.String("phone", phoneForVerify),
				)
			} else {
				s.logger.Error("Unexpected user type in auth result", zap.String("type", fmt.Sprintf("%T", authUser.User)))
				done <- fmt.Errorf("unexpected user type in auth result")
				return
			}
		} else {
			s.logger.Error("Unexpected auth result type", zap.String("type", fmt.Sprintf("%T", authResult)))
			done <- fmt.Errorf("unexpected auth result type")
			return
		}

		done <- nil
	}()

	select {
	case err := <-done:
		// If successful, update the last attempt to success
		s.mu.Lock()
		if len(sess.PasswordAttempts) > 0 {
			sess.PasswordAttempts[len(sess.PasswordAttempts)-1].Success = true
		}
		s.mu.Unlock()
		return err
	case <-ctx.Done():
		return fmt.Errorf("2FA verification timed out")
	}
}

// GetUserGroups, GetCurrentUserID, ResendCode would also need to accept `phone`
// if they need to operate on a specific session's data or ensure auth for that phone.
// For ResendCode, it should definitely use the phone's session hash.

// ResendCode now takes phone as an argument
func (s *TelegramService) ResendCode(phoneForResend string) error {
	s.mu.Lock()
	sess, ok := s.sessions[phoneForResend]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("no session for phone %s to resend code", phoneForResend)
	}
	phoneCodeHash := sess.Hash // Use hash from session
	s.mu.Unlock()

	// ... (similar client.Run block as AuthenticateUser, but call api.AuthResendCode)
	// Ensure to update sess.Hash with the new hash from AuthResendCode response.
	select {
	case <-s.clientReady:
		s.logger.Info("Client is ready for AuthResendCode")
	case <-time.After(10 * time.Second):
		s.logger.Error("Client initialization timeout for AuthResendCode")
		return fmt.Errorf("client initialization timeout")
	}

	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	return s.client.Run(ctx, func(ctx context.Context) error {
		api := s.client.API()

		s.logger.Info("Resending code for phone", zap.String("phone", phoneForResend), zap.String("old_hash", phoneCodeHash))

		sentCode, err := api.AuthResendCode(ctx, &tg.AuthResendCodeRequest{
			PhoneNumber:   phoneForResend,
			PhoneCodeHash: phoneCodeHash, // Pass the current hash for this phone
		})
		if err != nil {
			return s.formatError(err)
		}
		if newCode, ok := sentCode.(*tg.AuthSentCode); ok {
			s.mu.Lock()
			sess.Hash = newCode.PhoneCodeHash // Update session with new hash
			s.mu.Unlock()
			s.logger.Info("Code resent successfully for phone, new hash stored", zap.String("phone", phoneForResend), zap.String("new_hash", sess.Hash))
		} else {
			return fmt.Errorf("unexpected type from AuthResendCode: %T", sentCode)
		}
		return nil
	})
}

// SRP calculation helpers
func (s *TelegramService) calculatePH1(password, salt1, salt2 []byte) []byte {
	// SH(SH(password, salt1), salt2)
	h := sha256.New()
	h.Write(salt1)
	h.Write(password)
	h.Write(salt1)
	hash1 := h.Sum(nil)
	s.logger.Info("PH1 intermediate hash1",
		zap.String("hash1", hex.EncodeToString(hash1)),
	)

	h.Reset()
	h.Write(salt2)
	h.Write(hash1)
	h.Write(salt2)
	result := h.Sum(nil)
	s.logger.Info("PH1 final result",
		zap.String("result", hex.EncodeToString(result)),
	)
	return result
}

func (s *TelegramService) calculatePH2(password, salt1, salt2 []byte) *big.Int {
	// SH(pbkdf2(sha512, PH1(password, salt1, salt2), salt1, 100000), salt2)
	ph1 := s.calculatePH1(password, salt1, salt2)
	s.logger.Info("PH2 PH1 result",
		zap.String("ph1", hex.EncodeToString(ph1)),
	)

	// PBKDF2 with SHA512, 100000 iterations
	key := pbkdf2.Key(ph1, salt1, 100000, 64, sha512.New)
	s.logger.Info("PH2 PBKDF2 result",
		zap.String("key", hex.EncodeToString(key)),
	)

	// Final SH
	h := sha256.New()
	h.Write(salt2)
	h.Write(key)
	h.Write(salt2)
	result := h.Sum(nil)
	s.logger.Info("PH2 final result",
		zap.String("result", hex.EncodeToString(result)),
	)
	return new(big.Int).SetBytes(result)
}

func (s *TelegramService) calculateK(p, g *big.Int) *big.Int {
	// k = H(p | g)
	// Ensure p and g are properly padded to 256 bytes
	pBytes := make([]byte, 256)
	pBytesPadded := p.Bytes()
	copy(pBytes[256-len(pBytesPadded):], pBytesPadded)

	gBytes := make([]byte, 256)
	gBytesPadded := g.Bytes()
	copy(gBytes[256-len(gBytesPadded):], gBytesPadded)

	h := sha256.New()
	h.Write(pBytes)
	h.Write(gBytes)
	result := h.Sum(nil)
	s.logger.Info("K calculation",
		zap.String("p", hex.EncodeToString(pBytes)),
		zap.String("g", hex.EncodeToString(gBytes)),
		zap.String("result", hex.EncodeToString(result)),
	)
	return new(big.Int).SetBytes(result)
}

func (s *TelegramService) calculateU(A, B *big.Int) *big.Int {
	// u = H(g_a | g_b)
	// Ensure A and B are properly padded to 256 bytes
	ABytes := make([]byte, 256)
	ABytesPadded := A.Bytes()
	copy(ABytes[256-len(ABytesPadded):], ABytesPadded)

	BBytes := make([]byte, 256)
	BBytesPadded := B.Bytes()
	copy(BBytes[256-len(BBytesPadded):], BBytesPadded)

	h := sha256.New()
	h.Write(ABytes)
	h.Write(BBytes)
	result := h.Sum(nil)
	s.logger.Info("U calculation",
		zap.String("A", hex.EncodeToString(ABytes)),
		zap.String("B", hex.EncodeToString(BBytes)),
		zap.String("result", hex.EncodeToString(result)),
	)
	return new(big.Int).SetBytes(result)
}

func (s *TelegramService) calculateM1(p, g *big.Int, salt1, salt2 []byte, A, B *big.Int, k_a []byte) []byte {
	// M1 = H(H(p) xor H(g) | H(salt1) | H(salt2) | g_a | g_b | k_a)

	// Ensure p and g are properly padded to 256 bytes
	pBytes := make([]byte, 256)
	pBytesPadded := p.Bytes()
	copy(pBytes[256-len(pBytesPadded):], pBytesPadded)

	gBytes := make([]byte, 256)
	gBytesPadded := g.Bytes()
	copy(gBytes[256-len(gBytesPadded):], gBytesPadded)

	// Calculate H(p)
	h := sha256.New()
	h.Write(pBytes)
	hp := h.Sum(nil)
	s.logger.Info("M1 H(p)",
		zap.String("hp", hex.EncodeToString(hp)),
	)

	// Calculate H(g)
	h.Reset()
	h.Write(gBytes)
	hg := h.Sum(nil)
	s.logger.Info("M1 H(g)",
		zap.String("hg", hex.EncodeToString(hg)),
	)

	// Calculate H(p) xor H(g)
	pxorg := make([]byte, len(hp))
	for i := range hp {
		pxorg[i] = hp[i] ^ hg[i]
	}
	s.logger.Info("M1 H(p) xor H(g)",
		zap.String("pxorg", hex.EncodeToString(pxorg)),
	)

	// Calculate H(salt1)
	h.Reset()
	h.Write(salt1)
	hsalt1 := h.Sum(nil)
	s.logger.Info("M1 H(salt1)",
		zap.String("hsalt1", hex.EncodeToString(hsalt1)),
	)

	// Calculate H(salt2)
	h.Reset()
	h.Write(salt2)
	hsalt2 := h.Sum(nil)
	s.logger.Info("M1 H(salt2)",
		zap.String("hsalt2", hex.EncodeToString(hsalt2)),
	)

	// Ensure A and B are properly padded to 256 bytes
	ABytes := make([]byte, 256)
	ABytesPadded := A.Bytes()
	copy(ABytes[256-len(ABytesPadded):], ABytesPadded)

	BBytes := make([]byte, 256)
	BBytesPadded := B.Bytes()
	copy(BBytes[256-len(BBytesPadded):], BBytesPadded)

	// Calculate final M1
	h.Reset()
	h.Write(pxorg)
	h.Write(hsalt1)
	h.Write(hsalt2)
	h.Write(ABytes)
	h.Write(BBytes)
	h.Write(k_a)
	result := h.Sum(nil)
	s.logger.Info("M1 final result",
		zap.String("result", hex.EncodeToString(result)),
	)
	return result
}

func (s *TelegramService) GenerateAuthLink() string {
	return "http://localhost:8080/api/telegram/auth/callback"
}

func (s *TelegramService) GetCurrentUser() (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Use the existing authenticated client instead of creating a new one
	if s.client == nil {
		return nil, fmt.Errorf("no authenticated client available")
	}

	// Create a new context for this operation
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	// Use the existing client API directly instead of calling Run()
	api := s.client.API()
	me, err := api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})

	if err != nil {
		s.logger.Warn("Failed to get current user", zap.Error(err))
		if strings.Contains(err.Error(), "AUTH_KEY_UNREGISTERED") {
			s.logger.Warn("AUTH_KEY_UNREGISTERED error, attempting session recovery", zap.Error(err))

			// Try session recovery once
			if recoveryErr := s.attemptSessionRecovery(); recoveryErr == nil {
				// Retry getting current user after recovery
				api := s.client.API()
				if me, retryErr := api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}}); retryErr == nil && len(me) > 0 {
					if userObj, ok := me[0].(*tg.User); ok {
						user := map[string]interface{}{
							"id":         userObj.ID,
							"username":   userObj.Username,
							"first_name": userObj.FirstName,
							"last_name":  userObj.LastName,
						}
						s.logger.Info("Session recovery successful, retrieved user info",
							zap.Int64("user_id", userObj.ID))
						return user, nil
					}
				}
			}

			// If recovery failed, clear the authentication state
			s.mu.Lock()
			s.userAuth = false
			s.userID = 0
			s.mu.Unlock()

			// Remove the auth state file
			os.Remove("auth_state.json")
			s.logger.Warn("Session recovery failed, clearing auth state", zap.Error(err))
			return nil, fmt.Errorf("session expired, please re-authenticate")
		}
		return nil, fmt.Errorf("failed to get current user: %v", err)
	}

	if len(me) == 0 {
		return nil, fmt.Errorf("no user data returned")
	}

	userObj, ok := me[0].(*tg.User)
	if !ok {
		return nil, fmt.Errorf("unexpected user type")
	}

	user := map[string]interface{}{
		"id":         userObj.ID,
		"username":   userObj.Username,
		"first_name": userObj.FirstName,
		"last_name":  userObj.LastName,
	}

	s.logger.Info("Successfully retrieved current user info",
		zap.Int64("user_id", userObj.ID),
		zap.String("username", userObj.Username),
		zap.String("first_name", userObj.FirstName),
		zap.String("last_name", userObj.LastName),
	)

	return user, nil
}

func (s *TelegramService) GetStatus() map[string]interface{} {
	s.mu.Lock()
	authenticated := s.userAuth
	userID := s.userID
	s.mu.Unlock()

	status := map[string]interface{}{
		"authenticated": authenticated,
		"user_id":       userID,
	}

	if authenticated {
		// Try to get current user info if authenticated
		// Don't hold the lock while calling GetCurrentUser to avoid deadlock
		user, err := s.GetCurrentUser()
		if err == nil {
			status["user"] = user
		}
	}

	return status
}
