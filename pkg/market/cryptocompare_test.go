package market

import (
	"fmt"
	"testing"
	"time"
)

func TestSolanaContractPerformance(t *testing.T) {
	// Create a new Solana service
	service := NewSolanaService()

	// Test contracts
	contracts := []struct {
		address string
		token   string
	}{
		{"65c1fQCGScocK8rWTEPXBagXFXXzMaetS3L2rU5Epump", "TINYTIM"},
		{"9bAjHaaZmuikoWp18E3dxdNX1haSPXFev4X94m6Wpump", "Tard"},
		{"pwQm6YAqktQcaYHyEXxvZXUvH9MAZ9a7RuTrHgTpump", "GorbaganAI"},
		{"2H2eFHH41VCyaGHut4YbvCbXjJ6pBH4YqRd3ccxapump", "FrogHat"},
	}

	for _, contract := range contracts {
		t.Run(fmt.Sprintf("Test %s performance", contract.token), func(t *testing.T) {
			// Get token data using the contract address
			data, err := service.GetTokenData(contract.address)
			if err != nil {
				t.Logf("Warning: Failed to get data for %s (%s): %v", contract.token, contract.address, err)
				return
			}

			t.Logf("Token: %s (%s)", contract.token, contract.address)
			t.Logf("Price: %.8f", data.Price)
			t.Logf("Volume: %.2f", data.Volume)
			t.Logf("Performance: %.2f%%", data.Performance)
			t.Logf("Max Performance: %.2f%%", data.MaxPerformance)
			t.Logf("Time: %v", data.Time.Format(time.RFC3339))
			t.Logf("---")
		})
	}
}
