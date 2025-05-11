package telegram

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MockTelegramService is a test version of TelegramService that simulates the authentication flow
type MockTelegramService struct {
	mu                  sync.Mutex
	logger              *zap.Logger
	phone               string
	code                string
	hash                string
	password            string
	lastPasswordAttempt time.Time
	lastCodeAttempt     time.Time
	userID              int64
	clientReady         chan struct{}
	// Test configuration
	shouldFailCodeVerification bool
	shouldFail2FA              bool
	simulateCodeExpiration     bool
	simulateFloodWait          bool
}

// NewMockTelegramService creates a new mock Telegram service for testing
func NewMockTelegramService() *MockTelegramService {
	logger, _ := zap.NewDevelopment()
	service := &MockTelegramService{
		logger:      logger,
		clientReady: make(chan struct{}),
	}
	close(service.clientReady) // Mock client is always ready
	return service
}

// SetTestConfig allows configuring test behavior
func (s *MockTelegramService) SetTestConfig(config map[string]bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if val, ok := config["shouldFailCodeVerification"]; ok {
		s.shouldFailCodeVerification = val
	}
	if val, ok := config["shouldFail2FA"]; ok {
		s.shouldFail2FA = val
	}
	if val, ok := config["simulateCodeExpiration"]; ok {
		s.simulateCodeExpiration = val
	}
	if val, ok := config["simulateFloodWait"]; ok {
		s.simulateFloodWait = val
	}
}

// AuthenticateUser simulates sending a verification code
func (s *MockTelegramService) AuthenticateUser(phone string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.phone = phone
	s.hash = "mock_hash_" + fmt.Sprintf("%d", time.Now().Unix())
	s.logger.Info("Mock: Sent verification code", zap.String("phone", phone), zap.String("hash", s.hash))
	return nil
}

// VerifyCode simulates code verification
func (s *MockTelegramService) VerifyCode(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check rate limiting
	if !s.lastCodeAttempt.IsZero() {
		timeSinceLastAttempt := time.Since(s.lastCodeAttempt)
		if timeSinceLastAttempt < 30*time.Second {
			waitTime := 30*time.Second - timeSinceLastAttempt
			return fmt.Errorf("please wait %v before trying again", waitTime.Round(time.Second))
		}
	}

	s.lastCodeAttempt = time.Now()

	// Simulate code expiration
	if s.simulateCodeExpiration {
		s.hash = "mock_hash_" + fmt.Sprintf("%d", time.Now().Unix())
		return fmt.Errorf("CODE_EXPIRED_NEW_SENT")
	}

	// Simulate flood wait
	if s.simulateFloodWait {
		return fmt.Errorf("too many attempts. Please wait 24 hours before trying again. This is a security measure to protect your account")
	}

	// Simulate failed verification
	if s.shouldFailCodeVerification {
		return fmt.Errorf("invalid verification code")
	}

	// Simulate successful verification
	if code == "123456" { // Use a fixed test code
		return fmt.Errorf("SESSION_PASSWORD_NEEDED")
	}

	return fmt.Errorf("invalid verification code")
}

// Verify2FA simulates 2FA verification
func (s *MockTelegramService) Verify2FA(password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check rate limiting
	if !s.lastPasswordAttempt.IsZero() {
		timeSinceLastAttempt := time.Since(s.lastPasswordAttempt)
		if timeSinceLastAttempt < 30*time.Second {
			waitTime := 30*time.Second - timeSinceLastAttempt
			return fmt.Errorf("please wait %v before trying again", waitTime.Round(time.Second))
		}
	}

	s.lastPasswordAttempt = time.Now()

	// Simulate flood wait
	if s.simulateFloodWait {
		return fmt.Errorf("too many attempts. Please wait 24 hours before trying again. This is a security measure to protect your account")
	}

	// Simulate failed 2FA
	if s.shouldFail2FA {
		return fmt.Errorf("incorrect 2FA password")
	}

	// Simulate successful 2FA
	if password == "test123" { // Use a fixed test password
		s.userID = 12345
		return nil
	}

	return fmt.Errorf("incorrect 2FA password")
}

// GetCurrentUserID returns the mock user ID
func (s *MockTelegramService) GetCurrentUserID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.userID == 0 {
		return 0, fmt.Errorf("not authenticated")
	}
	return s.userID, nil
}

// GetPhone returns the current phone number
func (s *MockTelegramService) GetPhone() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.phone
}

// SetPhone sets a new phone number
func (s *MockTelegramService) SetPhone(phone string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.phone = phone
}

// GetPhoneCodeHash returns the stored phone code hash
func (s *MockTelegramService) GetPhoneCodeHash() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hash
}
