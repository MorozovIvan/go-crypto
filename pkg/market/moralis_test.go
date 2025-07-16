package market

import (
	"fmt"
	"testing"
	"time"
)

func TestMoralisTokenAnalysis(t *testing.T) {
	// Create a new Moralis service
	service := NewMoralisService()

	// Test cases for different token types
	testCases := []struct {
		address       string
		token         string
		signalTime    string
		description   string
		expectedType  TokenAddressType
		shouldAnalyze bool
	}{
		{
			address:       "0x8123f5bd81805fc38fcc4431286aad86c1091f15",
			token:         "Ichigo",
			signalTime:    "2025-06-21T16:21:21+02:00",
			description:   "Ethereum token address",
			expectedType:  Ethereum,
			shouldAnalyze: false, // Should return NON_SOLANA error
		},
		{
			address:       "65c1fQCGScocK8rWTEPXBagXFXXzMaetS3L2rU5Epump",
			token:         "TestPumpFun",
			signalTime:    "2025-06-22T08:00:00Z",
			description:   "Pump.fun token address",
			expectedType:  PumpFun,
			shouldAnalyze: true,
		},
		{
			address:       "9CdCXL6gX9m83tSwqNo8KNHQMio26UtK5RFDog7Hpump",
			token:         "PumpFun",
			signalTime:    "2025-06-22T10:00:00Z",
			description:   "Pump.fun token address",
			expectedType:  PumpFun,
			shouldAnalyze: true,
		},
		{
			address:       "invalid_address",
			token:         "Invalid",
			signalTime:    "2025-06-22T12:00:00Z",
			description:   "Invalid address format",
			expectedType:  Unknown,
			shouldAnalyze: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			// Test address type detection
			detectedType := service.DetectAddressType(testCase.address)
			if detectedType != testCase.expectedType {
				t.Errorf("Expected address type %v for %s, got %v",
					testCase.expectedType, testCase.address, detectedType)
			}

			// Test token analysis
			result, err := service.AnalyzeMaxGrowth(testCase.address, testCase.signalTime)

			// For Solana/PumpFun tokens, API errors are expected in tests since the tokens might not exist
			if err != nil && testCase.shouldAnalyze {
				// API errors are acceptable for Solana tokens in tests
				t.Logf("API error for %s (expected in tests): %v", testCase.token, err)
			}

			if result == nil {
				t.Errorf("Result should not be nil for %s", testCase.token)
				return
			}

			// Verify the result structure
			if testCase.expectedType == Ethereum {
				if result.Error != "Non-Solana token - should be marked on bar chart" {
					t.Errorf("Expected non-Solana error for %s, got: %s",
						testCase.token, result.Error)
				}
				if result.TokenType != "ethereum" {
					t.Errorf("Expected token type 'ethereum' for %s, got: %s",
						testCase.token, result.TokenType)
				}
			} else if testCase.expectedType == Unknown {
				if result.Error != "Unknown token address type" {
					t.Errorf("Expected unknown address type error for %s, got: %s",
						testCase.token, result.Error)
				}
				if result.TokenType != "unknown" {
					t.Errorf("Expected token type 'unknown' for %s, got: %s",
						testCase.token, result.TokenType)
				}
			} else if testCase.shouldAnalyze {
				// For Solana tokens, we expect either valid data or API errors (both are acceptable)
				if result.TokenType != "solana" {
					t.Errorf("Expected token type 'solana' for %s, got: %s",
						testCase.token, result.TokenType)
				}
				// API errors are expected for non-existent test tokens
				if result.Error != "" {
					t.Logf("Expected API error for test token %s: %s", testCase.token, result.Error)
				}
			}

			fmt.Printf("✓ %s Analysis Results:\n", testCase.token)
			fmt.Printf("  Address: %s\n", testCase.address)
			fmt.Printf("  Type: %s\n", result.TokenType)
			if result.Error != "" {
				fmt.Printf("  Error: %s\n", result.Error)
			} else {
				fmt.Printf("  Entry Price: $%.8f\n", result.EntryPrice)
				fmt.Printf("  Max Price: $%.8f\n", result.MaxPrice)
				fmt.Printf("  ROI: %.2f%%\n", result.ROIPercent)
				fmt.Printf("  Max X: %.2fx\n", result.MaxX)
				fmt.Printf("  Is Useful: %t\n", result.IsUseful)
				fmt.Printf("  Optimal Exit: %s\n", result.OptimalExit)
			}
			fmt.Printf("\n")
		})
	}
}

func TestAddressTypeDetection(t *testing.T) {
	service := NewMoralisService()

	// Test cases for address type detection
	testCases := []struct {
		address      string
		expectedType TokenAddressType
		description  string
	}{
		{
			address:      "0x8123f5bd81805fc38fcc4431286aad86c1091f15",
			expectedType: Ethereum,
			description:  "Valid Ethereum address",
		},
		{
			address:      "0x5655a136511686388B4a2dc054Ca5f260AFE268d",
			expectedType: Ethereum,
			description:  "Another valid Ethereum address",
		},
		{
			address:      "65c1fQCGScocK8rWTEPXBagXFXXzMaetS3L2rU5Epump",
			expectedType: PumpFun,
			description:  "Pump.fun token address (ends with pump and 44 chars)",
		},
		{
			address:      "So11111111111111111111111111111111111111112",
			expectedType: Solana,
			description:  "Valid Solana address (SOL token)",
		},
		{
			address:      "9CdCXL6gX9m83tSwqNo8KNHQMio26UtK5RFDog7Hpump",
			expectedType: PumpFun,
			description:  "Pump.fun token address",
		},
		{
			address:      "DNrmDMs2czDaAwgzg2BmvM7Jn5ZqA6VN5huRqCrSpump",
			expectedType: PumpFun,
			description:  "Another Pump.fun token address",
		},
		{
			address:      "invalid_address",
			expectedType: Unknown,
			description:  "Invalid address format",
		},
		{
			address:      "0x123", // Too short
			expectedType: Unknown,
			description:  "Too short Ethereum address",
		},
		{
			address:      "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghij", // Too long for Solana
			expectedType: Unknown,
			description:  "Too long address",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			result := service.DetectAddressType(testCase.address)
			if result != testCase.expectedType {
				t.Errorf("Expected %v for address %s, got %v",
					testCase.expectedType, testCase.address, result)
			}
		})
	}
}

func TestGetTokenPerformanceWrapper(t *testing.T) {
	service := NewMoralisService()

	// Test the wrapper function that matches the expected interface
	testCases := []struct {
		address     string
		signalTime  string
		description string
	}{
		{
			address:     "0x8123f5bd81805fc38fcc4431286aad86c1091f15",
			signalTime:  "2025-06-21T16:21:21+02:00",
			description: "Ethereum token should return NON_SOLANA",
		},
		{
			address:     "65c1fQCGScocK8rWTEPXBagXFXXzMaetS3L2rU5Epump",
			signalTime:  "2025-06-22T08:00:00Z",
			description: "Solana token should attempt analysis",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			performance, maxPerformance, optimalExit, isUseful := service.GetTokenPerformance(testCase.address, testCase.signalTime)

			// Verify the return values are not nil
			if performance == nil {
				t.Error("Performance should not be nil")
			}
			if maxPerformance == nil {
				t.Error("MaxPerformance should not be nil")
			}
			if optimalExit == "" {
				t.Error("OptimalExit should not be empty")
			}

			fmt.Printf("✓ %s Performance Results:\n", testCase.description)
			fmt.Printf("  Address: %s\n", testCase.address)
			fmt.Printf("  Performance: %v\n", performance)
			fmt.Printf("  Max Performance: %v\n", maxPerformance)
			fmt.Printf("  Optimal Exit: %s\n", optimalExit)
			fmt.Printf("  Is Useful: %t\n", isUseful)
			fmt.Printf("\n")
		})
	}
}

func TestMoralisAPIIntegration(t *testing.T) {
	// Skip this test if we don't want to make real API calls
	if testing.Short() {
		t.Skip("Skipping API integration test in short mode")
	}

	service := NewMoralisService()

	// Test with a known Solana token (this might fail if the token doesn't exist)
	testAddress := "65c1fQCGScocK8rWTEPXBagXFXXzMaetS3L2rU5Epump"
	signalTime := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)

	t.Run("Real API Call", func(t *testing.T) {
		result, err := service.AnalyzeMaxGrowth(testAddress, signalTime)

		// We expect either success or a specific API error
		if err != nil {
			t.Logf("API call failed (expected for test): %v", err)
		}

		if result != nil {
			fmt.Printf("Real API Test Results:\n")
			fmt.Printf("  Address: %s\n", testAddress)
			fmt.Printf("  Token Type: %s\n", result.TokenType)
			if result.Error != "" {
				fmt.Printf("  Error: %s\n", result.Error)
			} else {
				fmt.Printf("  Entry Price: $%.8f\n", result.EntryPrice)
				fmt.Printf("  Max Price: $%.8f\n", result.MaxPrice)
				fmt.Printf("  ROI: %.2f%%\n", result.ROIPercent)
				fmt.Printf("  Max X: %.2fx\n", result.MaxX)
				fmt.Printf("  Is Useful: %t\n", result.IsUseful)
				fmt.Printf("  Optimal Exit: %s\n", result.OptimalExit)
			}
		}
	})
}

func TestMoralisServiceMethods(t *testing.T) {
	service := NewMoralisService()

	// Test with a known token address
	tokenAddress := "65c1fQCGScocK8rWTEPXBagXFXXzMaetS3L2rU5Epump"

	t.Run("Test getPairAddress", func(t *testing.T) {
		pairAddress, err := service.getPairAddress(tokenAddress)
		if err != nil {
			t.Logf("getPairAddress error (expected for some tokens): %v", err)
			return
		}

		t.Logf("Pair address for %s: %s", tokenAddress, pairAddress)

		if pairAddress == "" {
			t.Error("Pair address should not be empty")
		}
	})

	t.Run("Test getOHLCVData", func(t *testing.T) {
		// First get pair address
		pairAddress, err := service.getPairAddress(tokenAddress)
		if err != nil {
			t.Skip("Skipping OHLCV test due to pair address error")
		}

		// Test OHLCV data retrieval
		startTime := "2024-06-21T14:35:00Z"
		ohlcvData, err := service.getOHLCVData(pairAddress, startTime, "1h")
		if err != nil {
			t.Logf("getOHLCVData error (may be expected): %v", err)
			return
		}

		t.Logf("Retrieved %d OHLCV data points", len(ohlcvData))

		if len(ohlcvData) > 0 {
			first := ohlcvData[0]
			t.Logf("First candle - Time: %s, Open: %f, High: %f, Low: %f, Close: %f",
				first.Timestamp, first.Open, first.High, first.Low, first.Close)
		}
	})
}

func TestMoralisErrorHandling(t *testing.T) {
	service := NewMoralisService()

	t.Run("Test invalid token address", func(t *testing.T) {
		result, err := service.AnalyzeMaxGrowth("invalid_address", "2024-06-21T14:35:00Z")
		if err != nil {
			t.Logf("Expected error for invalid address: %v", err)
		}

		if result != nil && result.Error == "" {
			t.Error("Should have error for invalid token address")
		}
	})

	t.Run("Test invalid time format", func(t *testing.T) {
		result, err := service.AnalyzeMaxGrowth("65c1fQCGScocK8rWTEPXBagXFXXzMaetS3L2rU5Epump", "invalid_time")
		if err != nil {
			t.Logf("Expected error for invalid time: %v", err)
		}

		if result != nil && result.Error == "" {
			t.Error("Should have error for invalid time format")
		}
	})
}

func TestMoralisRealTimeAnalysis(t *testing.T) {
	service := NewMoralisService()

	// Test with recent signal time (last 24 hours)
	recentSignalTime := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)

	testCases := []string{
		"65c1fQCGScocK8rWTEPXBagXFXXzMaetS3L2rU5Epump",
		"9bAjHaaZmuikoWp18E3dxdNX1haSPXFev4X94m6Wpump",
	}

	for _, address := range testCases {
		t.Run(fmt.Sprintf("Real-time analysis for %s", address[:8]), func(t *testing.T) {
			result, err := service.AnalyzeMaxGrowth(address, recentSignalTime)
			if err != nil {
				t.Logf("Error in real-time analysis: %v", err)
				return
			}

			if result.Error != "" {
				t.Logf("Analysis error: %s", result.Error)
				return
			}

			t.Logf("Real-time analysis results for %s:", address[:8])
			t.Logf("- ROI: %.2f%%", result.ROIPercent)
			t.Logf("- Max X: %.2fx", result.MaxX)
			t.Logf("- Optimal Exit: %s", result.OptimalExit)
			t.Logf("- Is Useful: %t", result.IsUseful)
		})
	}
}
