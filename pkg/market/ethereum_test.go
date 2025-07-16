package market

import (
	"fmt"
	"testing"
)

func TestEthereumTokenAnalysis(t *testing.T) {
	// Create a new Ethereum service
	service := NewEthereumService()

	// Test cases for Ethereum tokens
	testCases := []struct {
		address     string
		token       string
		signalTime  string
		description string
		expectedROI float64
	}{
		{
			address:     "0x8123f5bd81805fc38fcc4431286aad86c1091f15",
			token:       "Ichigo",
			signalTime:  "2025-06-21T16:21:21+02:00",
			description: "Ichigo token analysis",
			expectedROI: 380.0,
		},
		{
			address:     "0x1234567890abcdef1234567890abcdef12345678",
			token:       "TestToken",
			signalTime:  "2025-06-20T10:15:00Z",
			description: "Test Ethereum token analysis",
			expectedROI: 0.0, // Will be generated
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			// Analyze token performance
			result, err := service.AnalyzeTokenPerformance(testCase.address, testCase.signalTime)
			if err != nil {
				t.Fatalf("Failed to analyze token %s: %v", testCase.token, err)
			}

			// Verify the result
			if result.Error != "" {
				t.Errorf("Analysis returned error for %s: %s", testCase.token, result.Error)
			}

			// For known tokens, verify expected ROI
			if testCase.expectedROI > 0 {
				if result.ROIPercent != testCase.expectedROI {
					t.Errorf("Expected ROI %.2f%% for %s, got %.2f%%",
						testCase.expectedROI, testCase.token, result.ROIPercent)
				}
			}

			// Verify that useful signals have ROI > 15%
			if result.IsUseful && result.ROIPercent <= 15.0 {
				t.Errorf("Token %s marked as useful but ROI is only %.2f%%",
					testCase.token, result.ROIPercent)
			}

			// Verify basic data integrity
			if result.EntryPrice <= 0 {
				t.Errorf("Entry price should be positive for %s, got %.8f",
					testCase.token, result.EntryPrice)
			}

			if result.MaxPrice < result.EntryPrice {
				t.Errorf("Max price should be >= entry price for %s, got max=%.8f, entry=%.8f",
					testCase.token, result.MaxPrice, result.EntryPrice)
			}

			if result.MaxX <= 0 {
				t.Errorf("Max X should be positive for %s, got %.2f",
					testCase.token, result.MaxX)
			}

			fmt.Printf("✓ %s Analysis Results:\n", testCase.token)
			fmt.Printf("  Address: %s\n", testCase.address)
			fmt.Printf("  Entry Price: $%.8f\n", result.EntryPrice)
			fmt.Printf("  Max Price: $%.8f\n", result.MaxPrice)
			fmt.Printf("  ROI: %.2f%%\n", result.ROIPercent)
			fmt.Printf("  Max X: %.2fx\n", result.MaxX)
			fmt.Printf("  Is Useful: %t\n", result.IsUseful)
			fmt.Printf("  Optimal Exit: %s\n", result.OptimalExit)
			fmt.Printf("\n")
		})
	}
}

func TestEthereumServiceInterface(t *testing.T) {
	// Test that the service implements the expected interface
	service := NewEthereumService()

	// Test GetTokenPerformance method
	perf, maxPerf, exit, useful := service.GetTokenPerformance(
		"0x8123f5bd81805fc38fcc4431286aad86c1091f15",
		"2025-06-21T16:21:21+02:00",
	)

	// Verify return types and values
	if perfStr, ok := perf.(string); !ok || perfStr == "N/A" {
		t.Errorf("Expected valid performance string, got %v", perf)
	}

	if maxPerfStr, ok := maxPerf.(string); !ok || maxPerfStr == "N/A" {
		t.Errorf("Expected valid max performance string, got %v", maxPerf)
	}

	if exit == "N/A" {
		t.Errorf("Expected valid optimal exit time, got %s", exit)
	}

	if !useful {
		t.Errorf("Ichigo should be marked as useful signal")
	}

	fmt.Printf("✓ Interface Test Results:\n")
	fmt.Printf("  Performance: %v\n", perf)
	fmt.Printf("  Max Performance: %v\n", maxPerf)
	fmt.Printf("  Optimal Exit: %s\n", exit)
	fmt.Printf("  Is Useful: %t\n", useful)
}

func TestEthereumAddressDetection(t *testing.T) {
	testCases := []struct {
		address    string
		isEthereum bool
	}{
		{"0x8123f5bd81805fc38fcc4431286aad86c1091f15", true},    // Ethereum
		{"65c1fQCGScocK8rWTEPXBagXFXXzMaetS3L2rU5Epump", false}, // Solana
		{"0x123", false}, // Too short
		{"8123f5bd81805fc38fcc4431286aad86c1091f15", false}, // No 0x prefix
	}

	for _, testCase := range testCases {
		isEth := len(testCase.address) == 42 && testCase.address[:2] == "0x"
		if isEth != testCase.isEthereum {
			t.Errorf("Address %s: expected isEthereum=%t, got %t",
				testCase.address, testCase.isEthereum, isEth)
		}
	}
}
