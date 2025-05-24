package telegram

import (
	"context"
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

// Reinstated AuthSession and sessions map
type AuthSession struct {
	Hash                string
	LastCodeAttempt     time.Time
	LastPasswordAttempt time.Time
	PhoneNumber         string
	CreatedAt           time.Time
	PhoneCodeHash       string
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
	case <-time.After(10 * time.Second):
		cancel() // Cancel the context if client doesn't initialize in time
		return nil, fmt.Errorf("client initialization timeout")
	}

	return service, nil
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

	// Send code request using the API directly
	api := s.client.API()
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
		if auth == nil { /* ... */
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

	if !sess.LastPasswordAttempt.IsZero() {
		timeSinceLastAttempt := time.Since(sess.LastPasswordAttempt)
		if timeSinceLastAttempt < 30*time.Second {
			waitTime := 30*time.Second - timeSinceLastAttempt
			s.mu.Unlock()
			s.logger.Info("Rate limiting password attempt for phone", zap.String("phone", phoneForVerify))
			return fmt.Errorf("please wait %v before trying again", waitTime.Round(time.Second))
		}
	}
	sess.LastPasswordAttempt = time.Now()
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
		api := s.client.API()

		// Get current password settings
		passwordSettings, err := api.AccountGetPassword(ctx)
		if err != nil {
			s.logger.Error("Failed to get password settings", zap.Error(err))
			done <- fmt.Errorf("failed to get password settings: %v", err)
			return
		}

		// Type assert the password algorithm
		algo, ok := passwordSettings.CurrentAlgo.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
		if !ok {
			s.logger.Error("Unexpected password algorithm type", zap.String("type", fmt.Sprintf("%T", passwordSettings.CurrentAlgo)))
			done <- fmt.Errorf("unexpected password algorithm type")
			return
		}

		// Type assert the new algorithm
		newAlgo, ok := passwordSettings.NewAlgo.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
		if !ok {
			s.logger.Error("Unexpected new algorithm type", zap.String("type", fmt.Sprintf("%T", passwordSettings.NewAlgo)))
			done <- fmt.Errorf("unexpected new algorithm type")
			return
		}

		// Calculate password hash
		passwordHash := s.calculatePasswordHash(
			algo.Salt1,
			[]byte(password),
			algo.Salt2,
		)

		// Convert SRP parameters
		p := new(big.Int).SetBytes(newAlgo.P)
		g := big.NewInt(int64(newAlgo.G))
		srpB := new(big.Int).SetBytes(passwordSettings.SRPB)

		// Calculate M1
		m1 := s.calculateM1(p, g, algo.Salt1, srpB, srpB, passwordHash)

		// Sign in with 2FA
		_, err = api.AuthCheckPassword(ctx, &tg.InputCheckPasswordSRP{
			SRPID: passwordSettings.SRPID,
			A:     passwordSettings.SRPB,
			M1:    m1,
		})
		if err != nil {
			s.logger.Error("2FA verification failed", zap.Error(err))
			done <- s.formatError(err)
			return
		}

		s.logger.Info("2FA verification successful")
		done <- nil
	}()

	select {
	case err := <-done:
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

//SRP calculation helpers (calculatePasswordHash, calculateK, calculateU, calculateM1) remain the same.

// SRP calculation helpers (calculatePasswordHash, calculateK, calculateU, calculateM1) remain the same.
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

func (s *TelegramService) GetCurrentUser() (map[string]interface{}, error) {
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

	var user map[string]interface{}

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

		user = map[string]interface{}{
			"id":         userObj.ID,
			"username":   userObj.Username,
			"first_name": userObj.FirstName,
			"last_name":  userObj.LastName,
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
