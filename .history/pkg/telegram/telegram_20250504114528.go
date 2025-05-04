package telegram

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
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

// DefaultPhoneNumber is the default phone number to use
const DefaultPhoneNumber = "+380500232802"

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
		phone:    DefaultPhoneNumber, // Set default phone number
	}, nil
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
		s.phone = DefaultPhoneNumber
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

// formatError formats Telegram API errors into user-friendly messages
func (s *TelegramService) formatError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "PHONE_PASSWORD_FLOOD"):
		return fmt.Errorf("Too many failed attempts. For security reasons, please wait 24 hours before trying again. You can also try using the official Telegram app to reset your password.")
	case strings.Contains(errStr, "FLOOD_WAIT"):
		waitTime := "24 hours"
		if strings.Contains(errStr, "FLOOD_WAIT_") {
			parts := strings.Split(errStr, "FLOOD_WAIT_")
			if len(parts) > 1 {
				waitTime = parts[1]
			}
		}
		return fmt.Errorf("Too many attempts. Please wait %s before trying again. This is a security measure to protect your account.", waitTime)
	case strings.Contains(errStr, "PHONE_NUMBER_INVALID"):
		return fmt.Errorf("Invalid phone number format. Please check the number and try again. Make sure to include the country code (e.g., +1 for US).")
	case strings.Contains(errStr, "PHONE_CODE_INVALID"):
		return fmt.Errorf("Invalid verification code. Please check the code and try again. Make sure you're using the most recent code sent to your phone.")
	case strings.Contains(errStr, "PHONE_CODE_EXPIRED"):
		return fmt.Errorf("Verification code has expired. Please request a new code by clicking the 'Send Code' button again.")
	case strings.Contains(errStr, "PASSWORD_HASH_INVALID"):
		return fmt.Errorf("Incorrect 2FA password. Please check your password and try again. If you've forgotten your password, you can reset it in the official Telegram app.")
	case strings.Contains(errStr, "SESSION_PASSWORD_NEEDED"):
		return fmt.Errorf("2FA password required. Please enter your 2FA password to continue. If you haven't set up 2FA, please do so in the official Telegram app first.")
	case strings.Contains(errStr, "AUTH_RESTART"):
		return fmt.Errorf("Authentication session expired. Please start the authentication process again.")
	default:
		return fmt.Errorf("Authentication failed: %v", err)
	}
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
			// Handle specific error cases
			switch {
			case strings.Contains(err.Error(), "AUTH_RESTART"):
				// If we get AUTH_RESTART, try to resend the code
				sentCode, err = api.AuthResendCode(ctx, &tg.AuthResendCodeRequest{
					PhoneNumber:   phone,
					PhoneCodeHash: s.hash,
				})
				if err != nil {
					return s.formatError(err)
				}
			default:
				return s.formatError(err)
			}
		}

		// Store the phone code hash
		if code, ok := sentCode.(*tg.AuthSentCode); ok {
			s.hash = code.PhoneCodeHash
			s.logger.Info("Stored phone code hash", zap.String("hash", s.hash))
		} else {
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

	s.logger.Info("Starting code verification",
		zap.String("phone", s.phone),
		zap.String("code", code),
		zap.String("hash", s.hash))

	// Validate required fields
	if s.phone == "" {
		return s.formatError(fmt.Errorf("phone number is required"))
	}
	if s.hash == "" {
		return s.formatError(fmt.Errorf("phone code hash is required"))
	}
	if code == "" {
		return s.formatError(fmt.Errorf("verification code is required"))
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
		return s.formatError(fmt.Errorf("failed to create client: %v", err))
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
					return s.formatError(fmt.Errorf("2FA password is required"))
				}

				s.logger.Info("Attempting to sign in with 2FA password")

				// Get password settings
				settings, err := api.AccountGetPassword(ctx)
				if err != nil {
					s.logger.Error("Failed to get password settings", zap.Error(err))
					return s.formatError(err)
				}

				// Check if 2FA is actually enabled
				if !settings.HasPassword {
					s.logger.Info("2FA is not enabled for this account")
					return s.formatError(fmt.Errorf("2FA is not enabled for this account. Please enable 2FA in the Telegram app first"))
				}

				// Get the algorithm
				algo, ok := settings.CurrentAlgo.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
				if !ok {
					return s.formatError(fmt.Errorf("unsupported password algorithm"))
				}

				// Generate random 'a' value (256-bit)
				a := make([]byte, 32)
				if _, err := rand.Read(a); err != nil {
					return s.formatError(fmt.Errorf("failed to generate random value: %v", err))
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
				B := new(big.Int).SetBytes(algo.Salt1)

				// Calculate u = H(A, B)
				u := s.calculateU(A, B)

				// Calculate S = (B - k*g^x)^(a + u*x) mod p
				gx := new(big.Int).Exp(g, x, p)
				kgx := new(big.Int).Mul(k, gx)
				diff := new(big.Int).Sub(B, kgx)
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
				_, err = api.AuthCheckPassword(ctx, input)
				if err != nil {
					s.logger.Error("Failed to verify 2FA password", zap.Error(err))
					return s.formatError(err)
				}
				return nil
			}
			return s.formatError(err)
		}

		// Check if authentication was successful
		if auth != nil {
			s.logger.Info("Successfully signed in")
			return nil
		}

		return s.formatError(fmt.Errorf("authentication failed: unexpected response"))
	})
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

	return new(big.Int).SetBytes(hash2)
}

func (s *TelegramService) calculateK(p, g *big.Int) *big.Int {
	h := sha256.New()
	h.Write(p.Bytes())
	h.Write([]byte{byte(g.Int64())})
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
	h.Write([]byte{byte(g.Int64())})
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
