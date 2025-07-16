package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	moralisAPIKey  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJub25jZSI6ImE0NDA4ZjE4LWU5MGItNDI0ZC1hY2JkLTFkY2NhMzEwMWE1MiIsIm9yZ0lkIjoiNDU1MjMyIiwidXNlcklkIjoiNDY4MzczIiwidHlwZUlkIjoiYzM5MGUwZWEtMmRiYy00ZWE3LTgxYTAtMWRmOGM3YjBkN2M1IiwidHlwZSI6IlBST0pFQ1QiLCJpYXQiOjE3NTA2MDU1NTIsImV4cCI6NDkwNjM2NTU1Mn0.Qn4lfrL4xqnemPNSYS9daMSXTQO35ojT48qbpLxexB0"
	moralisBaseURL = "https://solana-gateway.moralis.io"
	tokenMint      = "7UUUHUxGrDunmbL5g3veDMtb5ai7gCexiNLZtBJgpump"
)

type TokenPairsResponse struct {
	Pairs []struct {
		PairAddress string  `json:"pairAddress"`
		BaseToken   string  `json:"baseToken"`
		QuoteToken  string  `json:"quoteToken"`
		PriceUsd    float64 `json:"priceUsd,string"`
		Volume24h   float64 `json:"volume24h,string"`
		Liquidity   float64 `json:"liquidity,string"`
	} `json:"pairs"`
}

type OHLCVResponse struct {
	Result []struct {
		Timestamp int64   `json:"timestamp"`
		Open      float64 `json:"open,string"`
		High      float64 `json:"high,string"`
		Low       float64 `json:"low,string"`
		Close     float64 `json:"close,string"`
		Volume    float64 `json:"volume,string"`
	} `json:"result"`
}

func makeMoralisRequest(endpoint string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", moralisAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func testTokenPairs() error {
	fmt.Println("🔍 Testing Moralis Token Pairs API...")

	endpoint := fmt.Sprintf("%s/token/mainnet/%s/pairs", moralisBaseURL, tokenMint)
	fmt.Printf("Endpoint: %s\n", endpoint)

	data, err := makeMoralisRequest(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch token pairs: %v", err)
	}

	var response TokenPairsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("failed to parse token pairs response: %v", err)
	}

	fmt.Printf("✅ Found %d trading pairs\n", len(response.Pairs))

	for i, pair := range response.Pairs {
		fmt.Printf("  %d. Pair: %s\n", i+1, pair.PairAddress)
		fmt.Printf("     Base: %s\n", pair.BaseToken)
		fmt.Printf("     Quote: %s\n", pair.QuoteToken)
		fmt.Printf("     Price USD: $%.8f\n", pair.PriceUsd)
		fmt.Printf("     Volume 24h: $%.2f\n", pair.Volume24h)
		fmt.Printf("     Liquidity: $%.2f\n", pair.Liquidity)
		fmt.Println()
	}

	return nil
}

func testOHLCV() error {
	fmt.Println("📈 Testing Moralis OHLCV API...")

	// Calculate time range for last 24 hours
	toDate := time.Now().Unix() * 1000
	fromDate := toDate - (24 * 60 * 60 * 1000) // 24 hours ago

	fmt.Printf("Time range: %s to %s\n",
		time.Unix(fromDate/1000, 0).Format("2006-01-02 15:04:05"),
		time.Unix(toDate/1000, 0).Format("2006-01-02 15:04:05"))

	// Use the first pair address if available, otherwise use token mint
	pairAddress := tokenMint // fallback to token mint

	endpoint := fmt.Sprintf("%s/token/mainnet/pairs/%s/ohlcv?timeframe=1s&currency=usd&fromDate=%d&toDate=%d",
		moralisBaseURL, pairAddress, fromDate, toDate)

	fmt.Printf("Endpoint: %s\n", endpoint)

	data, err := makeMoralisRequest(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch OHLCV data: %v", err)
	}

	var response OHLCVResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("failed to parse OHLCV response: %v", err)
	}

	fmt.Printf("✅ Found %d OHLCV candles (1-second intervals)\n", len(response.Result))

	// Show first 5 and last 5 candles
	if len(response.Result) > 0 {
		fmt.Println("\n📊 First 5 candles:")
		for i := 0; i < 5 && i < len(response.Result); i++ {
			candle := response.Result[i]
			timestamp := time.Unix(candle.Timestamp/1000, 0).Format("15:04:05")
			fmt.Printf("  %s | O: %.8f | H: %.8f | L: %.8f | C: %.8f | V: %.2f\n",
				timestamp, candle.Open, candle.High, candle.Low, candle.Close, candle.Volume)
		}

		if len(response.Result) > 5 {
			fmt.Println("\n📊 Last 5 candles:")
			start := len(response.Result) - 5
			for i := start; i < len(response.Result); i++ {
				candle := response.Result[i]
				timestamp := time.Unix(candle.Timestamp/1000, 0).Format("15:04:05")
				fmt.Printf("  %s | O: %.8f | H: %.8f | L: %.8f | C: %.8f | V: %.2f\n",
					timestamp, candle.Open, candle.High, candle.Low, candle.Close, candle.Volume)
			}
		}

		// Calculate some basic stats
		if len(response.Result) > 0 {
			first := response.Result[0]
			last := response.Result[len(response.Result)-1]
			priceChange := ((last.Close - first.Open) / first.Open) * 100

			fmt.Printf("\n📊 Summary:\n")
			fmt.Printf("  Price change: %.2f%%\n", priceChange)
			fmt.Printf("  First price: $%.8f\n", first.Open)
			fmt.Printf("  Last price: $%.8f\n", last.Close)
			fmt.Printf("  Total candles: %d\n", len(response.Result))
		}
	}

	return nil
}

func main() {
	fmt.Printf("🚀 Testing Moralis API for token: %s\n", tokenMint)
	fmt.Println(strings.Repeat("=", 60))

	// Test token pairs first
	if err := testTokenPairs(); err != nil {
		fmt.Printf("❌ Token pairs test failed: %v\n", err)
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 60))

	// Test OHLCV data
	if err := testOHLCV(); err != nil {
		fmt.Printf("❌ OHLCV test failed: %v\n", err)
		return
	}

	fmt.Println("\n✅ All tests completed successfully!")
}
