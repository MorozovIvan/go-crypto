package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// API Keys
	smartShyftAPIKey   = "CEhOUGrWGGh7zX7i"
	smartMoralisAPIKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJub25jZSI6ImE0NDA4ZjE4LWU5MGItNDI0ZC1hY2JkLTFkY2NhMzEwMWE1MiIsIm9yZ0lkIjoiNDU1MjMyIiwidXNlcklkIjoiNDY4MzczIiwidHlwZUlkIjoiYzM5MGUwZWEtMmRiYy00ZWE3LTgxYTAtMWRmOGM3YjBkN2M1IiwidHlwZSI6IlBST0pFQ1QiLCJpYXQiOjE3NTA2MDU1NTIsImV4cCI6NDkwNjM2NTU1Mn0.Qn4lfrL4xqnemPNSYS9daMSXTQO35ojT48qbpLxexB0"

	// API Endpoints
	smartShyftEndpoint  = "https://rpc.shyft.to/?api_key=" + smartShyftAPIKey
	smartMoralisBaseURL = "https://solana-gateway.moralis.io"
	smartDexScreenerURL = "https://api.dexscreener.com/latest/dex"
)

type TimeFrame string

const (
	TimeFrame1S  TimeFrame = "1s"
	TimeFrame2S  TimeFrame = "2s"
	TimeFrame3S  TimeFrame = "3s"
	TimeFrame1M  TimeFrame = "1m"
	TimeFrame5M  TimeFrame = "5m"
	TimeFrame15M TimeFrame = "15m"
	TimeFrame1H  TimeFrame = "1h"
	TimeFrame4H  TimeFrame = "4h"
	TimeFrame1D  TimeFrame = "1d"
)

type Period string

const (
	Period24H Period = "24h"
	Period3D  Period = "3d"
	Period7D  Period = "7d"
	Period30D Period = "30d"
)

type MarketCapData struct {
	Timestamp  int64   `json:"timestamp"`
	Price      float64 `json:"price"`
	MarketCap  float64 `json:"marketCap"`
	Volume24h  float64 `json:"volume24h"`
	Supply     float64 `json:"supply"`
	Source     string  `json:"source"`
	Confidence float64 `json:"confidence"` // 0-1, how reliable the data is
}

type TokenAnalysis struct {
	TokenMint     string          `json:"tokenMint"`
	TokenName     string          `json:"tokenName"`
	TokenSymbol   string          `json:"tokenSymbol"`
	IsPumpFun     bool            `json:"isPumpFun"`
	MarketCapData []MarketCapData `json:"marketCapData"`
	TimeFrame     TimeFrame       `json:"timeFrame"`
	Period        Period          `json:"period"`
	LastUpdated   time.Time       `json:"lastUpdated"`
}

// Smart API types (renamed to avoid conflicts)
type SmartShyftSignatureResp struct {
	Result []struct {
		Signature string `json:"signature"`
		BlockTime int64  `json:"blockTime"`
	} `json:"result"`
}

type SmartShyftTransactionResp struct {
	Result struct {
		Meta struct {
			LogMessages []string    `json:"logMessages"`
			Err         interface{} `json:"err"`
		} `json:"meta"`
		BlockTime int64 `json:"blockTime"`
	} `json:"result"`
}

// Moralis API integration
type MoralisTokenResponse struct {
	Address     string `json:"address"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    int    `json:"decimals"`
	Logo        string `json:"logo"`
	LogoHash    string `json:"logoHash"`
	Thumbnail   string `json:"thumbnail"`
	BlockNumber string `json:"blockNumber"`
	Validated   int    `json:"validated"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Supply      string `json:"supply"`
}

type MoralisPairsResponse struct {
	Pairs []struct {
		PairAddress string  `json:"pairAddress"`
		BaseToken   string  `json:"baseToken"`
		QuoteToken  string  `json:"quoteToken"`
		PriceUsd    float64 `json:"priceUsd,string"`
		Volume24h   float64 `json:"volume24h,string"`
		Liquidity   float64 `json:"liquidity,string"`
	} `json:"pairs"`
}

// DexScreener API integration
type DexScreenerResponse struct {
	Pairs []struct {
		ChainId     string `json:"chainId"`
		DexId       string `json:"dexId"`
		PairAddress string `json:"pairAddress"`
		BaseToken   struct {
			Address string `json:"address"`
			Name    string `json:"name"`
			Symbol  string `json:"symbol"`
		} `json:"baseToken"`
		QuoteToken struct {
			Address string `json:"address"`
			Name    string `json:"name"`
			Symbol  string `json:"symbol"`
		} `json:"quoteToken"`
		PriceUsd  string `json:"priceUsd"`
		Volume24h string `json:"volume24h"`
		Liquidity struct {
			Usd float64 `json:"usd"`
		} `json:"liquidity"`
		Fdv float64 `json:"fdv"`
	} `json:"pairs"`
}

type SmartAPICombiner struct {
	httpClient *http.Client
}

func NewSmartAPICombiner() *SmartAPICombiner {
	return &SmartAPICombiner{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}



func (c *SmartAPICombiner) AnalyzeToken(tokenAddress string, period Period, timeFrame TimeFrame) (*TokenAnalysis, error) {
	fmt.Printf("🔍 Analyzing token: %s (Period: %s, TimeFrame: %s)\n", tokenAddress, period, timeFrame)

	analysis := &TokenAnalysis{
		TokenMint:   tokenAddress,
		TimeFrame:   timeFrame,
		Period:      period,
		LastUpdated: time.Now(),
	}

	// Step 1: Detect if it's a pump.fun token
	analysis.IsPumpFun = strings.Contains(tokenAddress, "pump")

	// Step 2: Get token metadata (try Moralis first, then DexScreener)
	if err := c.getTokenMetadata(analysis); err != nil {
		fmt.Printf("⚠️  Failed to get token metadata: %v\n", err)
	}

	// Step 3: Get market cap data from multiple sources
	marketCapData, err := c.getMarketCapData(tokenAddress, period, timeFrame)
	if err != nil {
		return nil, fmt.Errorf("failed to get market cap data: %v", err)
	}

	analysis.MarketCapData = marketCapData

	return analysis, nil
}

func (c *SmartAPICombiner) getTokenMetadata(analysis *TokenAnalysis) error {
	// Try Moralis first
	moralisData, err := c.getMoralisTokenMetadata(analysis.TokenMint)
	if err == nil && moralisData != nil {
		analysis.TokenName = moralisData.Name
		analysis.TokenSymbol = moralisData.Symbol
		return nil
	}

	// Fallback to DexScreener
	dexData, err := c.getDexScreenerTokenData(analysis.TokenMint)
	if err == nil && len(dexData.Pairs) > 0 {
		pair := dexData.Pairs[0]
		analysis.TokenName = pair.BaseToken.Name
		analysis.TokenSymbol = pair.BaseToken.Symbol
		return nil
	}

	return fmt.Errorf("failed to get token metadata from both sources")
}

func (c *SmartAPICombiner) getMoralisTokenMetadata(tokenMint string) (*MoralisTokenResponse, error) {
	url := fmt.Sprintf("%s/token/mainnet/%s", smartMoralisBaseURL, tokenMint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", smartMoralisAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var result MoralisTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *SmartAPICombiner) getDexScreenerTokenData(tokenMint string) (*DexScreenerResponse, error) {
	url := fmt.Sprintf("%s/tokens/%s", smartDexScreenerURL, tokenMint)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var result DexScreenerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *SmartAPICombiner) getMarketCapData(tokenMint string, period Period, timeFrame TimeFrame) ([]MarketCapData, error) {
	var allData []MarketCapData

	// Calculate time range
	toDate := time.Now().Unix() * 1000
	var fromDate int64

	switch period {
	case Period24H:
		fromDate = toDate - (24 * 60 * 60 * 1000)
	case Period3D:
		fromDate = toDate - (3 * 24 * 60 * 60 * 1000)
	case Period7D:
		fromDate = toDate - (7 * 24 * 60 * 60 * 1000)
	case Period30D:
		fromDate = toDate - (30 * 24 * 60 * 60 * 1000)
	}

	// Strategy 1: Try Shyft for real-time OHLCV data
	fmt.Println("📊 Strategy 1: Fetching real-time data from Shyft...")
	shyftData, err := c.getShyftOHLCVData(tokenMint, fromDate, toDate, timeFrame)
	if err == nil && len(shyftData) > 0 {
		fmt.Printf("✅ Got %d data points from Shyft\n", len(shyftData))
		allData = append(allData, shyftData...)
	} else {
		fmt.Printf("❌ Shyft data failed: %v\n", err)
	}

	// Strategy 2: Try Moralis for historical data
	fmt.Println("📊 Strategy 2: Fetching historical data from Moralis...")
	moralisData, err := c.getMoralisOHLCVData(tokenMint, fromDate, toDate, timeFrame)
	if err == nil && len(moralisData) > 0 {
		fmt.Printf("✅ Got %d data points from Moralis\n", len(moralisData))
		allData = append(allData, moralisData...)
	} else {
		fmt.Printf("❌ Moralis data failed: %v\n", err)
	}

	// Strategy 3: Try DexScreener for current market cap
	fmt.Println("📊 Strategy 3: Fetching current data from DexScreener...")
	dexData, err := c.getDexScreenerMarketCap(tokenMint)
	if err == nil && len(dexData) > 0 {
		fmt.Printf("✅ Got %d data points from DexScreener\n", len(dexData))
		allData = append(allData, dexData...)
	} else {
		fmt.Printf("❌ DexScreener data failed: %v\n", err)
	}

	// Merge and deduplicate data
	mergedData := c.mergeAndDeduplicateData(allData)

	// Sort by timestamp
	c.sortByTimestamp(mergedData)

	return mergedData, nil
}

func (c *SmartAPICombiner) getShyftOHLCVData(tokenMint string, fromDate, toDate int64, timeFrame TimeFrame) ([]MarketCapData, error) {
	// This would implement the Shyft transaction parsing logic we developed earlier
	// For now, return empty slice as placeholder
	return []MarketCapData{}, nil
}

func (c *SmartAPICombiner) getMoralisOHLCVData(tokenMint string, fromDate, toDate int64, timeFrame TimeFrame) ([]MarketCapData, error) {
	// Try to get pairs first
	pairs, err := c.getMoralisPairs(tokenMint)
	if err != nil || len(pairs.Pairs) == 0 {
		return []MarketCapData{}, fmt.Errorf("no pairs found")
	}

	// Use the first pair for OHLCV data
	pairAddress := pairs.Pairs[0].PairAddress

	url := fmt.Sprintf("%s/token/mainnet/pairs/%s/ohlcv?timeframe=%s&currency=usd&fromDate=%d&toDate=%d",
		smartMoralisBaseURL, pairAddress, timeFrame, fromDate, toDate)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", smartMoralisAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return []MarketCapData{}, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Parse OHLCV response and convert to MarketCapData
	// This is a simplified version - you'd need to implement the full parsing
	return []MarketCapData{}, nil
}

func (c *SmartAPICombiner) getMoralisPairs(tokenMint string) (*MoralisPairsResponse, error) {
	url := fmt.Sprintf("%s/token/mainnet/%s/pairs", smartMoralisBaseURL, tokenMint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", smartMoralisAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var result MoralisPairsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *SmartAPICombiner) getDexScreenerMarketCap(tokenMint string) ([]MarketCapData, error) {
	dexData, err := c.getDexScreenerTokenData(tokenMint)
	if err != nil {
		return nil, err
	}

	var marketCapData []MarketCapData

	for _, pair := range dexData.Pairs {
		price, _ := strconv.ParseFloat(pair.PriceUsd, 64)
		volume, _ := strconv.ParseFloat(pair.Volume24h, 64)

		data := MarketCapData{
			Timestamp:  time.Now().Unix() * 1000,
			Price:      price,
			MarketCap:  pair.Fdv,
			Volume24h:  volume,
			Source:     "DexScreener",
			Confidence: 0.8, // High confidence for current data
		}

		marketCapData = append(marketCapData, data)
	}

	return marketCapData, nil
}

func (c *SmartAPICombiner) mergeAndDeduplicateData(data []MarketCapData) []MarketCapData {
	// Implementation for merging and deduplicating data points
	// This would remove duplicates and combine data from different sources
	return data
}

func (c *SmartAPICombiner) sortByTimestamp(data []MarketCapData) {
	// Implementation for sorting by timestamp
}

func testSmartAPICombiner() {
	combiner := NewSmartAPICombiner()

	// Test with the pump.fun token
	tokenAddress := "7UUUHUxGrDunmbL5g3veDMtb5ai7gCexiNLZtBJgpump"

	analysis, err := combiner.AnalyzeToken(tokenAddress, Period24H, TimeFrame1S)
	if err != nil {
		fmt.Printf("❌ Analysis failed: %v\n", err)
		return
	}

	fmt.Printf("\n📈 Analysis Results:\n")
	fmt.Printf("Token: %s (%s)\n", analysis.TokenName, analysis.TokenSymbol)
	fmt.Printf("Is Pump.fun: %t\n", analysis.IsPumpFun)
	fmt.Printf("Data points: %d\n", len(analysis.MarketCapData))
	fmt.Printf("Timeframe: %s\n", analysis.TimeFrame)
	fmt.Printf("Period: %s\n", analysis.Period)

	if len(analysis.MarketCapData) > 0 {
		fmt.Printf("\n📊 Sample Data Points:\n")
		for i, data := range analysis.MarketCapData {
			if i >= 5 { // Show first 5 data points
				break
			}
			timestamp := time.Unix(data.Timestamp/1000, 0).Format("2006-01-02 15:04:05")
			fmt.Printf("  %s | Price: $%.8f | Market Cap: $%.2f | Source: %s (%.1f%% confidence)\n",
				timestamp, data.Price, data.MarketCap, data.Source, data.Confidence*100)
		}
	}
}
