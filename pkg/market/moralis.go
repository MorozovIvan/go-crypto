package market

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	MORALIS_API_KEY = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJub25jZSI6ImE0NDA4ZjE4LWU5MGItNDI0ZC1hY2JkLTFkY2NhMzEwMWE1MiIsIm9yZ0lkIjoiNDU1MjMyIiwidXNlcklkIjoiNDY4MzczIiwidHlwZUlkIjoiYzM5MGUwZWEtMmRiYy00ZWE3LTgxYTAtMWRmOGM3YjBkN2M1IiwidHlwZSI6IlBST0pFQ1QiLCJpYXQiOjE3NTA2MDU1NTIsImV4cCI6NDkwNjM2NTU1Mn0.Qn4lfrL4xqnemPNSYS9daMSXTQO35ojT48qbpLxexB0"
	BASE_URL        = "https://solana-gateway.moralis.io"
	MAX_RETRIES     = 3
	RETRY_DELAY     = 2 * time.Second
)

// TokenAddressType represents the type of token address
type TokenAddressType int

const (
	Unknown  TokenAddressType = iota
	Ethereum                  // 0x... addresses
	Solana                    // Solana addresses (32-44 chars)
	PumpFun                   // Pump.fun internal project IDs
)

// String returns the string representation of TokenAddressType
func (t TokenAddressType) String() string {
	switch t {
	case Ethereum:
		return "ethereum"
	case Solana:
		return "solana"
	case PumpFun:
		return "pumpfun"
	default:
		return "unknown"
	}
}

// MoralisService handles interactions with the Moralis API
type MoralisService struct {
	client  *http.Client
	headers map[string]string
}

type PairData struct {
	PairAddress        string  `json:"pairAddress"`
	ExchangeName       string  `json:"exchangeName"`
	PairLabel          string  `json:"pairLabel"`
	UsdPrice           float64 `json:"usdPrice"`
	UsdPrice24hrChange float64 `json:"usdPrice24hrPercentChange"`
	BaseToken          string  `json:"baseToken"`
	QuoteToken         string  `json:"quoteToken"`
	InactivePair       bool    `json:"inactivePair"`
}

type PairsResponse struct {
	Pairs []PairData `json:"pairs"`
}

type OHLCVData struct {
	Timestamp string  `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
	Trades    int     `json:"trades"`
	MarketCap float64 `json:"market_cap,omitempty"` // Calculated market cap
}

type OHLCVResponse struct {
	Result []OHLCVData `json:"result"`
}

type AnalysisResult struct {
	EntryPrice     float64 `json:"entry_price"`
	MaxPrice       float64 `json:"max_price"`
	EntryMarketCap float64 `json:"entry_market_cap"`
	MaxMarketCap   float64 `json:"max_market_cap"`
	ROIPercent     float64 `json:"roi_percent"`
	MaxX           float64 `json:"max_x"`
	Error          string  `json:"error,omitempty"`
	IsUseful       bool    `json:"is_useful"`
	OptimalExit    string  `json:"optimal_exit"`
	TokenType      string  `json:"token_type"`
}

// TokenMetadata represents token information including supply
type TokenMetadata struct {
	Address           string  `json:"address"`
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	Decimals          int     `json:"decimals"`
	Supply            float64 `json:"supply"`
	TotalSupply       float64 `json:"total_supply"`
	CirculatingSupply float64 `json:"circulating_supply"`
}

type TokenMetadataResponse struct {
	Result TokenMetadata `json:"result"`
}

// NewMoralisService creates a new instance of MoralisService
func NewMoralisService() *MoralisService {
	return &MoralisService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: map[string]string{
			"X-API-Key": MORALIS_API_KEY,
			"Accept":    "application/json",
		},
	}
}

// DetectAddressType determines the type of token address
func (m *MoralisService) DetectAddressType(address string) TokenAddressType {
	address = strings.TrimSpace(address)

	// Ethereum addresses start with 0x and are 42 characters long
	if matched, _ := regexp.MatchString(`^0x[a-fA-F0-9]{40}$`, address); matched {
		return Ethereum
	}

	// Solana addresses are base58 encoded, typically 32-44 characters
	// Base58 uses: 123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz (no 0, O, I, l)
	if len(address) >= 32 && len(address) <= 44 {
		// Check if it contains only valid base58 characters
		if matched, _ := regexp.MatchString(`^[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+$`, address); matched {
			// Check if it's a pump.fun token address pattern (more specific check)
			if strings.HasSuffix(address, "pump") && len(address) == 44 {
				return PumpFun
			}
			return Solana
		}
	}

	return Unknown
}

func (m *MoralisService) doRequest(req *http.Request) (*http.Response, error) {
	var lastErr error
	for i := 0; i < MAX_RETRIES; i++ {
		if i > 0 {
			time.Sleep(RETRY_DELAY)
			log.Printf("Retrying request to %s (attempt %d/%d)", req.URL, i+1, MAX_RETRIES)
		}

		resp, err := m.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %v", err)
			continue
		}

		if resp.StatusCode == 429 {
			log.Printf("Rate limit hit, waiting before retry...")
			resp.Body.Close()
			time.Sleep(RETRY_DELAY * 2) // Wait longer for rate limits
			continue
		}

		if resp.StatusCode != 200 {
			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body[:n])
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("all retries failed: %v", lastErr)
}

func (m *MoralisService) GetPairAddress(tokenAddress string) (string, error) {
	url := fmt.Sprintf("%s/token/mainnet/%s/pairs", BASE_URL, tokenAddress)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	for key, value := range m.headers {
		req.Header.Set(key, value)
	}

	resp, err := m.doRequest(req)
	if err != nil {
		return "", fmt.Errorf("failed to get pair data: %v", err)
	}
	defer resp.Body.Close()

	var pairsResp PairsResponse
	if err := json.NewDecoder(resp.Body).Decode(&pairsResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(pairsResp.Pairs) == 0 {
		return "", fmt.Errorf("no pairs found for token %s", tokenAddress)
	}

	return pairsResp.Pairs[0].PairAddress, nil
}

func (m *MoralisService) getOHLCVData(pairAddress string, startTimeISO string, timeframe string) ([]OHLCVData, error) {
	startTime, err := time.Parse(time.RFC3339, startTimeISO)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %v", err)
	}

	endTime := time.Now().UTC()

	// Try different time ranges to get data
	timeRanges := []struct {
		from time.Time
		to   time.Time
		desc string
	}{
		{
			from: startTime,
			to:   endTime,
			desc: "signal time to now",
		},
		{
			from: startTime.Add(-1 * time.Hour), // 1 hour before signal
			to:   endTime,
			desc: "1 hour before signal to now",
		},
		{
			from: startTime.Add(-24 * time.Hour), // 24 hours before signal
			to:   endTime,
			desc: "24 hours before signal to now",
		},
		{
			from: endTime.Add(-24 * time.Hour), // Last 24 hours only
			to:   endTime,
			desc: "last 24 hours",
		},
	}

	for _, timeRange := range timeRanges {
		fromTime := timeRange.from.UnixMilli()
		toTime := timeRange.to.UnixMilli()

		log.Printf("Requesting OHLCV data for pair %s from %s to %s (timeframe: %s) - %s",
			pairAddress, timeRange.from.Format(time.RFC3339), timeRange.to.Format(time.RFC3339), timeframe, timeRange.desc)

		url := fmt.Sprintf("%s/token/mainnet/pairs/%s/ohlcv?timeframe=%s&currency=usd&fromDate=%d&toDate=%d",
			BASE_URL, pairAddress, timeframe, fromTime, toTime)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		for key, value := range m.headers {
			req.Header.Set(key, value)
		}

		resp, err := m.doRequest(req)
		if err != nil {
			log.Printf("Failed to get OHLCV data for %s: %v", timeRange.desc, err)
			continue
		}
		defer resp.Body.Close()

		var ohlcvResp OHLCVResponse
		if err := json.NewDecoder(resp.Body).Decode(&ohlcvResp); err != nil {
			log.Printf("Failed to decode OHLCV response for %s: %v", timeRange.desc, err)
			continue
		}

		log.Printf("Received %d OHLCV candles for pair %s (%s)", len(ohlcvResp.Result), pairAddress, timeRange.desc)

		if len(ohlcvResp.Result) > 0 {
			// Filter data to only include candles after or equal to signal time
			filtered := m.filterOHLCVData(ohlcvResp.Result, startTime)
			log.Printf("After filtering: %d candles remain for pair %s", len(filtered), pairAddress)

			if len(filtered) > 0 {
				return filtered, nil
			} else {
				// Check if we have data but it's all before signal time
				earliestCandle := ohlcvResp.Result[0]
				earliestTime, _ := time.Parse(time.RFC3339, earliestCandle.Timestamp)
				log.Printf("Token %s has data from %s, but signal was at %s (token created after signal)",
					pairAddress, earliestTime.Format(time.RFC3339), startTime.Format(time.RFC3339))
			}
		}
	}

	// If we get here, no data was found in any time range
	return nil, fmt.Errorf("no OHLCV data found in any time range for token %s", pairAddress)
}

func (m *MoralisService) filterOHLCVData(data []OHLCVData, startTime time.Time) []OHLCVData {
	var filtered []OHLCVData
	for _, candle := range data {
		candleTime, err := time.Parse(time.RFC3339, candle.Timestamp)
		if err != nil {
			log.Printf("Failed to parse timestamp %s: %v", candle.Timestamp, err)
			continue
		}

		// Keep only candles strictly after or equal to the signal time
		if candleTime.After(startTime) || candleTime.Equal(startTime) {
			filtered = append(filtered, candle)
		}
	}
	return filtered
}

// AnalyzeMaxGrowth implements the Python logic for analyzing token performance
func (m *MoralisService) AnalyzeMaxGrowth(tokenAddress string, callTimeISO string) (*AnalysisResult, error) {
	log.Printf("Analyzing token %s from time %s", tokenAddress, callTimeISO)

	// Detect address type
	addressType := m.DetectAddressType(tokenAddress)

	switch addressType {
	case Ethereum:
		// For Ethereum tokens, mark as non-Solana and return appropriate message
		return &AnalysisResult{
			Error:     "Non-Solana token - should be marked on bar chart",
			TokenType: "ethereum",
		}, nil

	case Solana, PumpFun:
		return m.analyzeSolanaToken(tokenAddress, callTimeISO, addressType)

	default:
		return &AnalysisResult{
			Error:     "Unknown token address type",
			TokenType: "unknown",
		}, nil
	}
}

func (m *MoralisService) analyzeSolanaToken(tokenAddress string, callTimeISO string, addressType TokenAddressType) (*AnalysisResult, error) {
	// Step 1: Get pair address
	pairAddress, err := m.GetPairAddress(tokenAddress)
	if err != nil {
		return &AnalysisResult{
			Error:     fmt.Sprintf("Failed to get pair address: %v", err),
			TokenType: "solana",
		}, err
	}

	log.Printf("Found pair address for %s: %s", tokenAddress, pairAddress)

	// Step 2: Try different timeframes to get OHLCV data
	timeframes := []string{"1s", "10s", "30s", "1min", "5min", "10min", "30min", "1h", "4h", "12h", "1d", "1w", "1M"}

	for _, timeframe := range timeframes {
		log.Printf("Trying timeframe %s for token %s", timeframe, tokenAddress)

		ohlcvData, err := m.getOHLCVData(pairAddress, callTimeISO, timeframe)
		if err == nil && len(ohlcvData) > 0 {
			log.Printf("Successfully got %d candles with timeframe %s for %s", len(ohlcvData), timeframe, tokenAddress)
			// Step 3: Analyze the data using the Python logic
			return m.analyzeOHLCVData(ohlcvData, callTimeISO, tokenAddress)
		}

		log.Printf("Failed to get OHLCV data with timeframe %s for %s: %v", timeframe, tokenAddress, err)
	}

	// If we get here, no timeframe worked
	// Check if this is a pumpfun token and provide specific error message
	if addressType == PumpFun {
		return &AnalysisResult{
			Error:     "PumpFun token analysis failed: no price data available in any timeframe",
			TokenType: "pumpfun",
		}, fmt.Errorf("no OHLCV data available for pumpfun token %s", tokenAddress)
	}

	return &AnalysisResult{
		Error:     "Failed to get OHLCV data in any timeframe",
		TokenType: "solana",
	}, fmt.Errorf("no OHLCV data available for token %s", tokenAddress)
}

func (m *MoralisService) analyzeOHLCVData(ohlcvData []OHLCVData, callTimeISO, tokenAddress string) (*AnalysisResult, error) {
	callTime, err := time.Parse(time.RFC3339, callTimeISO)
	if err != nil {
		return nil, fmt.Errorf("failed to parse call time: %v", err)
	}

	if len(ohlcvData) == 0 {
		return &AnalysisResult{
			Error:     "No OHLCV data available",
			TokenType: "solana",
		}, nil
	}

	// Calculate market cap for all candles
	ohlcvData, err = m.calculateMarketCap(ohlcvData, tokenAddress)
	if err != nil {
		log.Printf("Warning: Could not calculate market cap for %s: %v", tokenAddress, err)
	}

	// Find entry candle (closest to call time)
	var entryCandle *OHLCVData
	minTimeDiff := time.Duration(1<<63 - 1) // Max duration

	for i := range ohlcvData {
		candleTime, err := time.Parse(time.RFC3339, ohlcvData[i].Timestamp)
		if err != nil {
			continue
		}

		timeDiff := candleTime.Sub(callTime)
		if timeDiff >= 0 && timeDiff < minTimeDiff {
			minTimeDiff = timeDiff
			entryCandle = &ohlcvData[i]
		}
	}

	if entryCandle == nil {
		return &AnalysisResult{
			Error:     "No suitable entry candle found",
			TokenType: "solana",
		}, nil
	}

	entryPrice := entryCandle.Open
	entryMarketCap := entryCandle.MarketCap

	// Find maximum high after signal
	var maxHigh float64
	var maxHighCandle *OHLCVData

	for _, candle := range ohlcvData {
		if candle.High > maxHigh {
			maxHigh = candle.High
			maxHighCandle = &candle
		}
	}

	if maxHigh <= entryPrice {
		return &AnalysisResult{
			EntryPrice:     entryPrice,
			MaxPrice:       entryPrice,
			EntryMarketCap: entryMarketCap,
			MaxMarketCap:   entryMarketCap,
			ROIPercent:     0.0,
			MaxX:           1.0,
			IsUseful:       false,
			OptimalExit:    "No price increase found",
			TokenType:      "solana",
		}, nil
	}

	// Calculate metrics using market cap if available, otherwise use price
	var roi float64
	var maxMarketCap float64

	if entryMarketCap > 0 && maxHighCandle != nil && maxHighCandle.MarketCap > 0 {
		// Use market cap for analysis
		maxMarketCap = maxHighCandle.MarketCap
		roi = (maxMarketCap/entryMarketCap - 1) * 100
		log.Printf("Using market cap analysis for %s: Entry MC=%.2f, Max MC=%.2f, ROI=%.2f%%",
			tokenAddress, entryMarketCap, maxMarketCap, roi)
	} else {
		// Fallback to price analysis
		maxMarketCap = maxHigh * entryMarketCap / entryPrice // Estimate
		roi = (maxHigh/entryPrice - 1) * 100
		log.Printf("Using price analysis for %s: Entry=%.8f, Max=%.8f, ROI=%.2f%%",
			tokenAddress, entryPrice, maxHigh, roi)
	}

	xGrowth := round(maxHigh/entryPrice, 2)

	// Calculate optimal exit time and percentage
	var maxHighTime time.Time
	if maxHighCandle != nil {
		maxHighTime, _ = time.Parse(time.RFC3339, maxHighCandle.Timestamp)
	}
	duration := maxHighTime.Sub(callTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	// The optimal exit is the maximum ROI percentage achieved
	optimalExitPercent := roi

	isUseful := roi > 15.0

	log.Printf("Analysis for %s: Entry=%.8f, Max=%.8f, ROI=%.2f%%, X=%.2fx, OptimalExit=%.2f%% at %dh %dm",
		tokenAddress, entryPrice, maxHigh, roi, xGrowth, optimalExitPercent, hours, minutes)

	return &AnalysisResult{
		EntryPrice:     entryPrice,
		MaxPrice:       maxHigh,
		EntryMarketCap: entryMarketCap,
		MaxMarketCap:   maxMarketCap,
		ROIPercent:     round(roi, 2),
		MaxX:           xGrowth,
		IsUseful:       isUseful,
		OptimalExit:    fmt.Sprintf("%.2f%%", optimalExitPercent),
		TokenType:      "solana",
	}, nil
}

// round rounds a float64 to the specified number of decimal places
func round(num float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Round(num*shift) / shift
}

// GetTokenPerformance is a wrapper that matches the interface expected by the main application
func (m *MoralisService) GetTokenPerformance(contractAddr, signalTimeStr string) (interface{}, interface{}, string, bool) {
	result, err := m.AnalyzeMaxGrowth(contractAddr, signalTimeStr)
	if err != nil {
		log.Printf("Moralis analysis failed for %s: %v", contractAddr, err)
		return "N/A", "N/A", "N/A", false
	}

	if result.Error != "" {
		log.Printf("Moralis analysis error for %s: %s", contractAddr, result.Error)

		// For Ethereum tokens, return special marker
		if result.TokenType == "ethereum" {
			return "NON_SOLANA", "NON_SOLANA", "Mark on bar chart", false
		}

		return "N/A", "N/A", "N/A", false
	}

	// Return performance as percentage for consistency
	perf := result.ROIPercent
	maxPerf := result.ROIPercent

	log.Printf("Moralis performance for %s: %.2f%% (Max: %.2f%%)", contractAddr, perf, maxPerf)

	return perf, maxPerf, result.OptimalExit, result.IsUseful
}

// GetTokenAnalysisData returns the full analysis result for the API
func (m *MoralisService) GetTokenAnalysisData(contractAddr, signalTimeStr string) *AnalysisResult {
	return m.GetTokenAnalysisDataWithPeriod(contractAddr, signalTimeStr, "24h")
}

// GetTokenAnalysisDataWithPeriod returns the full analysis result for the API with a specified period
func (m *MoralisService) GetTokenAnalysisDataWithPeriod(contractAddr, signalTimeStr, period string) *AnalysisResult {
	result, err := m.AnalyzeMaxGrowthWithPeriod(contractAddr, signalTimeStr, period)
	if err != nil {
		log.Printf("Moralis analysis failed for %s: %v", contractAddr, err)
		return &AnalysisResult{
			Error:     fmt.Sprintf("Analysis failed: %v", err),
			TokenType: "unknown",
		}
	}
	return result
}

// AnalyzeMaxGrowthWithPeriod analyzes the token performance with a specified period
func (m *MoralisService) AnalyzeMaxGrowthWithPeriod(tokenAddress string, callTimeISO string, period string) (*AnalysisResult, error) {
	log.Printf("Analyzing token %s from time %s (period: %s)", tokenAddress, callTimeISO, period)
	addressType := m.DetectAddressType(tokenAddress)
	switch addressType {
	case Ethereum:
		return &AnalysisResult{
			Error:     "Non-Solana token - should be marked on bar chart",
			TokenType: "ethereum",
		}, nil
	case Solana, PumpFun:
		return m.analyzeSolanaTokenWithPeriod(tokenAddress, callTimeISO, addressType, period)
	default:
		return &AnalysisResult{
			Error:     "Unknown token address type",
			TokenType: "unknown",
		}, nil
	}
}

// analyzeSolanaTokenWithPeriod analyzes the token performance with a specified period
func (m *MoralisService) analyzeSolanaTokenWithPeriod(tokenAddress string, callTimeISO string, addressType TokenAddressType, period string) (*AnalysisResult, error) {
	pairAddress, err := m.GetPairAddress(tokenAddress)
	if err != nil {
		return &AnalysisResult{
			Error:     fmt.Sprintf("Failed to get pair address: %v", err),
			TokenType: "solana",
		}, err
	}
	log.Printf("Found pair address for %s: %s", tokenAddress, pairAddress)

	// Calculate start time based on period
	endTime := time.Now().UTC()
	var duration time.Duration
	switch period {
	case "3d":
		duration = 72 * time.Hour
	case "7d":
		duration = 168 * time.Hour
	default:
		duration = 24 * time.Hour
	}
	startTime := endTime.Add(-duration)

	timeframes := []string{"1s", "10s", "30s", "1min", "5min", "10min", "30min", "1h", "4h", "12h", "1d", "1w", "1M"}
	for _, timeframe := range timeframes {
		log.Printf("Trying timeframe %s for token %s", timeframe, tokenAddress)
		ohlcvData, err := m.GetOHLCVDataRange(pairAddress, startTime, endTime, timeframe)
		if err == nil && len(ohlcvData) > 0 {
			log.Printf("Successfully got %d candles with timeframe %s for %s", len(ohlcvData), timeframe, tokenAddress)
			return m.analyzeOHLCVDataClosest(ohlcvData, callTimeISO, tokenAddress)
		}
		log.Printf("Failed to get OHLCV data with timeframe %s for %s: %v", timeframe, tokenAddress, err)
	}
	if addressType == PumpFun {
		return &AnalysisResult{
			Error:     "PumpFun token analysis failed: no price data available in any timeframe",
			TokenType: "pumpfun",
		}, fmt.Errorf("no OHLCV data available for pumpfun token %s", tokenAddress)
	}
	return &AnalysisResult{
		Error:     "Failed to get OHLCV data in any timeframe",
		TokenType: "solana",
	}, fmt.Errorf("no OHLCV data available for token %s", tokenAddress)
}

// getOHLCVDataRange gets OHLCV data for a specified range and timeframe
func (m *MoralisService) GetOHLCVDataRange(pairAddress string, startTime, endTime time.Time, timeframe string) ([]OHLCVData, error) {
	fromTime := startTime.UnixMilli()
	toTime := endTime.UnixMilli()
	log.Printf("Requesting OHLCV data for pair %s from %s to %s (timeframe: %s)",
		pairAddress, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), timeframe)
	url := fmt.Sprintf("%s/token/mainnet/pairs/%s/ohlcv?timeframe=%s&currency=usd&fromDate=%d&toDate=%d",
		BASE_URL, pairAddress, timeframe, fromTime, toTime)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	for key, value := range m.headers {
		req.Header.Set(key, value)
	}
	resp, err := m.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get OHLCV data: %v", err)
	}
	defer resp.Body.Close()
	var ohlcvResp OHLCVResponse
	if err := json.NewDecoder(resp.Body).Decode(&ohlcvResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	log.Printf("Received %d OHLCV candles for pair %s", len(ohlcvResp.Result), pairAddress)
	return ohlcvResp.Result, nil
}

// analyzeOHLCVDataClosest analyzes the token performance with the closest candle to the signal time
func (m *MoralisService) analyzeOHLCVDataClosest(ohlcvData []OHLCVData, callTimeISO, tokenAddress string) (*AnalysisResult, error) {
	callTime, err := time.Parse(time.RFC3339, callTimeISO)
	if err != nil {
		return nil, fmt.Errorf("failed to parse call time: %v", err)
	}
	if len(ohlcvData) == 0 {
		return &AnalysisResult{
			Error:     "No OHLCV data available",
			TokenType: "solana",
		}, nil
	}

	// Calculate market cap for all candles
	ohlcvData, err = m.calculateMarketCap(ohlcvData, tokenAddress)
	if err != nil {
		log.Printf("Warning: Could not calculate market cap for %s: %v", tokenAddress, err)
	}

	// Find entry candle (closest to call time, before or after)
	var entryCandle *OHLCVData
	minTimeDiff := time.Duration(1<<63 - 1)
	for i := range ohlcvData {
		candleTime, err := time.Parse(time.RFC3339, ohlcvData[i].Timestamp)
		if err != nil {
			continue
		}
		timeDiff := candleTime.Sub(callTime)
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}
		if timeDiff < minTimeDiff {
			minTimeDiff = timeDiff
			entryCandle = &ohlcvData[i]
		}
	}
	if entryCandle == nil {
		return &AnalysisResult{
			Error:     "No suitable entry candle found",
			TokenType: "solana",
		}, nil
	}
	entryPrice := entryCandle.Open
	entryMarketCap := entryCandle.MarketCap

	// Find maximum high after signal
	var maxHigh float64
	var maxHighCandle *OHLCVData
	for _, candle := range ohlcvData {
		candleTime, _ := time.Parse(time.RFC3339, candle.Timestamp)
		if candleTime.After(callTime) && candle.High > maxHigh {
			maxHigh = candle.High
			maxHighCandle = &candle
		}
	}
	if maxHigh <= entryPrice {
		return &AnalysisResult{
			EntryPrice:     entryPrice,
			MaxPrice:       entryPrice,
			EntryMarketCap: entryMarketCap,
			MaxMarketCap:   entryMarketCap,
			ROIPercent:     0.0,
			MaxX:           1.0,
			IsUseful:       false,
			OptimalExit:    "No price increase found",
			TokenType:      "solana",
		}, nil
	}

	// Calculate metrics using market cap if available, otherwise use price
	var roi float64
	var maxMarketCap float64

	if entryMarketCap > 0 && maxHighCandle != nil && maxHighCandle.MarketCap > 0 {
		// Use market cap for analysis
		maxMarketCap = maxHighCandle.MarketCap
		roi = (maxMarketCap/entryMarketCap - 1) * 100
		log.Printf("Using market cap analysis for %s: Entry MC=%.2f, Max MC=%.2f, ROI=%.2f%%",
			tokenAddress, entryMarketCap, maxMarketCap, roi)
	} else {
		// Fallback to price analysis
		maxMarketCap = maxHigh * entryMarketCap / entryPrice // Estimate
		roi = (maxHigh/entryPrice - 1) * 100
		log.Printf("Using price analysis for %s: Entry=%.8f, Max=%.8f, ROI=%.2f%%",
			tokenAddress, entryPrice, maxHigh, roi)
	}

	xGrowth := round(maxHigh/entryPrice, 2)

	var maxHighTime time.Time
	if maxHighCandle != nil {
		maxHighTime, _ = time.Parse(time.RFC3339, maxHighCandle.Timestamp)
	}
	duration := maxHighTime.Sub(callTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	optimalExitPercent := roi
	isUseful := roi > 15.0
	log.Printf("Analysis for %s: Entry=%.8f, Max=%.8f, ROI=%.2f%%, X=%.2fx, OptimalExit=%.2f%% at %dh %dm",
		tokenAddress, entryPrice, maxHigh, roi, xGrowth, optimalExitPercent, hours, minutes)
	return &AnalysisResult{
		EntryPrice:     entryPrice,
		MaxPrice:       maxHigh,
		EntryMarketCap: entryMarketCap,
		MaxMarketCap:   maxMarketCap,
		ROIPercent:     round(roi, 2),
		MaxX:           xGrowth,
		IsUseful:       isUseful,
		OptimalExit:    fmt.Sprintf("%.2f%%", optimalExitPercent),
		TokenType:      "solana",
	}, nil
}

// getTokenMetadata retrieves token metadata including supply information
func (m *MoralisService) GetTokenMetadata(tokenAddress string) (*TokenMetadata, error) {
	url := fmt.Sprintf("%s/token/mainnet/%s", BASE_URL, tokenAddress)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	for key, value := range m.headers {
		req.Header.Set(key, value)
	}

	resp, err := m.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get token metadata: %v", err)
	}
	defer resp.Body.Close()

	var metadataResp TokenMetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&metadataResp); err != nil {
		return nil, fmt.Errorf("failed to decode metadata response: %v", err)
	}

	return &metadataResp.Result, nil
}

// calculateMarketCap calculates market cap for each OHLCV candle
func (m *MoralisService) calculateMarketCap(ohlcvData []OHLCVData, tokenAddress string) ([]OHLCVData, error) {
	// Get token metadata to get supply information
	metadata, err := m.GetTokenMetadata(tokenAddress)
	if err != nil {
		log.Printf("Failed to get token metadata for %s: %v", tokenAddress, err)
		// Return data without market cap if metadata is unavailable
		return ohlcvData, nil
	}

	// Use circulating supply if available, otherwise use total supply
	supply := metadata.CirculatingSupply
	if supply == 0 {
		supply = metadata.TotalSupply
	}
	if supply == 0 {
		supply = metadata.Supply
	}

	if supply == 0 {
		log.Printf("No supply information available for token %s", tokenAddress)
		return ohlcvData, nil
	}

	// Calculate market cap for each candle using close price
	for i := range ohlcvData {
		ohlcvData[i].MarketCap = ohlcvData[i].Close * supply
	}

	log.Printf("Calculated market cap for %s using supply: %.2f", tokenAddress, supply)
	return ohlcvData, nil
}
