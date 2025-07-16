package market

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RPCSwapEvent represents a parsed swap event from transaction logs via RPC
type RPCSwapEvent struct {
	Signature    string                 `json:"signature"`
	Timestamp    int64                  `json:"timestamp"`
	Slot         uint64                 `json:"slot"`
	AmountIn     float64                `json:"amount_in"`
	AmountOut    float64                `json:"amount_out"`
	Price        float64                `json:"price"`
	Direction    string                 `json:"direction"` // "buy" or "sell"
	TokenIn      string                 `json:"token_in"`
	TokenOut     string                 `json:"token_out"`
	FeeAmount    float64                `json:"fee_amount"`
	FeeToken     string                 `json:"fee_token"`
	RawLog       string                 `json:"raw_log"`
	ParsedFields map[string]interface{} `json:"parsed_fields"`
}

// OHLCVCandle represents aggregated OHLCV data
type OHLCVCandle struct {
	Timestamp  int64   `json:"timestamp"`
	Open       float64 `json:"open"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Close      float64 `json:"close"`
	Volume     float64 `json:"volume"`
	TradeCount int     `json:"trade_count"`
}

// SolanaRPCRequest represents a JSON-RPC request
type SolanaRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// SolanaRPCResponse represents a JSON-RPC response
type SolanaRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// PumpFunAMMProgram is the program ID for pump.fun AMM
const PumpFunAMMProgram = "PFundAMM111111111111111111111111111111111111"

// SwapIndexerService handles RPC-based swap indexing
type SwapIndexerService struct {
	rpcEndpoints []string
	client       *http.Client
}

// NewSwapIndexerService creates a new swap indexer service
func NewSwapIndexerService() *SwapIndexerService {
	return &SwapIndexerService{
		rpcEndpoints: []string{
			"https://rpc.helius.xyz/?api-key=" + getHeliusAPIKey(),
			"https://solana-mainnet.rpc.extrnode.com",
			"https://api.mainnet-beta.solana.com",
			"https://solana-api.projectserum.com",
		},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// getHeliusAPIKey returns the Helius API key from environment
func getHeliusAPIKey() string {
	return os.Getenv("HELIUS_API_KEY")
}

// GetSwapEvents fetches and parses swap events for a mint address
func (s *SwapIndexerService) GetSwapEvents(mintAddress string, limit int) ([]RPCSwapEvent, error) {
	rpcURL := s.getBestRPCURL()

	log.Printf("Indexing swap events for mint: %s using RPC: %s", mintAddress, rpcURL)

	// Use reasonable limit for faster processing
	if limit > 500 {
		limit = 500 // Maximum 500 transactions for faster response
	}
	if limit < 100 {
		limit = 100 // Minimum 100 transactions for some data
	}

	// Get recent signatures for the mint address
	signatures, err := s.getSignaturesForAddress(mintAddress, rpcURL, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get signatures: %v", err)
	}

	log.Printf("Found %d recent transactions", len(signatures))

	var allSwapEvents []RPCSwapEvent
	var processedCount int

	// Process each transaction with minimal rate limiting
	for i, sig := range signatures {
		if i >= limit {
			break
		}

		events, err := s.parseTransactionForSwaps(sig, mintAddress, rpcURL)
		if err != nil {
			log.Printf("Error parsing transaction %s: %v", sig, err)
			continue
		}

		if len(events) > 0 {
			allSwapEvents = append(allSwapEvents, events...)
			log.Printf("Found %d swap events in transaction %s", len(events), sig)
		}

		processedCount++
		if processedCount%100 == 0 {
			log.Printf("Processed %d/%d transactions", processedCount, len(signatures))
		}

		// Minimal rate limiting - only 10ms to avoid overwhelming RPC
		time.Sleep(10 * time.Millisecond)
	}

	log.Printf("Total swap events found: %d", len(allSwapEvents))
	return allSwapEvents, nil
}

// GetOHLCVData fetches swap events and aggregates them into OHLCV candles
func (s *SwapIndexerService) GetOHLCVData(mintAddress, timeframe string, limit int) ([]OHLCVCandle, error) {
	// Use reasonable limits for faster processing
	switch timeframe {
	case "1m", "5m":
		limit = 300 // Reduced for faster response
	case "15m", "1h":
		limit = 200
	case "4h", "1d":
		limit = 100
	default:
		limit = 200
	}

	events, err := s.GetSwapEvents(mintAddress, limit)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return []OHLCVCandle{}, nil
	}

	return s.aggregateToOHLCV(events, timeframe), nil
}

// getBestRPCURL returns the best available RPC URL
func (s *SwapIndexerService) getBestRPCURL() string {
	// Check for Helius API key first
	if getHeliusAPIKey() != "" {
		return fmt.Sprintf("https://rpc.helius.xyz/?api-key=%s", getHeliusAPIKey())
	}

	// Use fallback endpoints
	return s.rpcEndpoints[2] // Default to Solana mainnet
}

// getSignaturesForAddress fetches recent transaction signatures for a mint address
func (s *SwapIndexerService) getSignaturesForAddress(address, rpcURL string, limit int) ([]string, error) {
	// Validate the address format first
	if len(address) < 32 || len(address) > 44 {
		return nil, fmt.Errorf("invalid address format: %s", address)
	}

	// For pump.fun tokens, we need to search for transactions involving the token
	// Use a more conservative limit to avoid RPC errors
	if limit > 1000 {
		limit = 1000
	}

	req := SolanaRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "getSignaturesForAddress",
		Params: []interface{}{
			address,
			map[string]interface{}{
				"limit":      limit,
				"commitment": "confirmed",
			},
		},
	}

	resp, err := s.makeRPCRequestWithRetry(rpcURL, req)
	if err != nil {
		return nil, err
	}

	var result []struct {
		Signature string `json:"signature"`
		Slot      uint64 `json:"slot"`
		BlockTime *int64 `json:"blockTime,omitempty"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal signatures: %v", err)
	}

	var signatures []string
	for _, r := range result {
		signatures = append(signatures, r.Signature)
	}

	log.Printf("Retrieved %d transaction signatures for %s", len(signatures), address)
	return signatures, nil
}

// parseTransactionForSwaps fetches transaction details and parses swap events
func (s *SwapIndexerService) parseTransactionForSwaps(signature, mint, rpcURL string) ([]RPCSwapEvent, error) {
	req := SolanaRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "getTransaction",
		Params: []interface{}{
			signature,
			map[string]interface{}{
				"encoding":                       "json",
				"maxSupportedTransactionVersion": 0,
			},
		},
	}

	resp, err := s.makeRPCRequestWithRetry(rpcURL, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction %s: %v", signature, err)
	}

	var txResult struct {
		Meta struct {
			LogMessages      []string `json:"logMessages"`
			PreTokenBalances []struct {
				AccountIndex  int    `json:"accountIndex"`
				Mint          string `json:"mint"`
				Owner         string `json:"owner"`
				ProgramId     string `json:"programId"`
				UiTokenAmount struct {
					Amount   string  `json:"amount"`
					Decimals int     `json:"decimals"`
					UiAmount float64 `json:"uiAmount"`
				} `json:"uiTokenAmount"`
			} `json:"preTokenBalances"`
			PostTokenBalances []struct {
				AccountIndex  int    `json:"accountIndex"`
				Mint          string `json:"mint"`
				Owner         string `json:"owner"`
				ProgramId     string `json:"programId"`
				UiTokenAmount struct {
					Amount   string  `json:"amount"`
					Decimals int     `json:"decimals"`
					UiAmount float64 `json:"uiAmount"`
				} `json:"uiTokenAmount"`
			} `json:"postTokenBalances"`
		} `json:"meta"`
		BlockTime int64  `json:"blockTime"`
		Slot      uint64 `json:"slot"`
	}

	if err := json.Unmarshal(resp.Result, &txResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %v", err)
	}

	if txResult.Meta.LogMessages == nil {
		return []RPCSwapEvent{}, nil
	}

	var events []RPCSwapEvent

	// Parse swap events from logs
	for _, logLine := range txResult.Meta.LogMessages {
		event := s.ParseSwapFromLog(logLine, signature, txResult.Slot, txResult.BlockTime, mint, txResult.Meta)
		if event != nil {
			events = append(events, *event)
		}
	}

	// If no events found in logs, try to infer from balance changes
	if len(events) == 0 {
		event := s.InferSwapFromBalances(signature, txResult.Slot, txResult.BlockTime, mint, txResult.Meta)
		if event != nil {
			events = append(events, *event)
		}
	}

	return events, nil
}

// ParseSwapFromLog parses a swap event from a log line
func (s *SwapIndexerService) ParseSwapFromLog(logLine, signature string, slot uint64, blockTime int64, mint string, meta struct {
	LogMessages      []string `json:"logMessages"`
	PreTokenBalances []struct {
		AccountIndex  int    `json:"accountIndex"`
		Mint          string `json:"mint"`
		Owner         string `json:"owner"`
		ProgramId     string `json:"programId"`
		UiTokenAmount struct {
			Amount   string  `json:"amount"`
			Decimals int     `json:"decimals"`
			UiAmount float64 `json:"uiAmount"`
		} `json:"uiTokenAmount"`
	} `json:"preTokenBalances"`
	PostTokenBalances []struct {
		AccountIndex  int    `json:"accountIndex"`
		Mint          string `json:"mint"`
		Owner         string `json:"owner"`
		ProgramId     string `json:"programId"`
		UiTokenAmount struct {
			Amount   string  `json:"amount"`
			Decimals int     `json:"decimals"`
			UiAmount float64 `json:"uiAmount"`
		} `json:"uiTokenAmount"`
	} `json:"postTokenBalances"`
}) *RPCSwapEvent {
	// Look for pump.fun AMM program logs
	if !strings.Contains(logLine, PumpFunAMMProgram) {
		return nil
	}

	// Parse amount patterns
	amountPatterns := []*regexp.Regexp{
		regexp.MustCompile(`amount_in: (\d+)`),
		regexp.MustCompile(`amount_out: (\d+)`),
		regexp.MustCompile(`Amount In: (\d+)`),
		regexp.MustCompile(`Amount Out: (\d+)`),
		regexp.MustCompile(`input_amount: (\d+)`),
		regexp.MustCompile(`output_amount: (\d+)`),
	}

	var amountIn, amountOut float64
	var direction string

	for _, pattern := range amountPatterns {
		matches := pattern.FindStringSubmatch(logLine)
		if len(matches) > 1 {
			if strings.Contains(pattern.String(), "in") || strings.Contains(pattern.String(), "input") {
				amountIn, _ = strconv.ParseFloat(matches[1], 64)
			} else if strings.Contains(pattern.String(), "out") || strings.Contains(pattern.String(), "output") {
				amountOut, _ = strconv.ParseFloat(matches[1], 64)
			}
		}
	}

	// Determine direction based on which token is the mint
	if amountIn > 0 && amountOut > 0 {
		// This is a simplified logic - in reality, you'd need to check which token is which
		direction = "swap"
	}

	// Calculate price if possible
	var price float64
	if amountOut > 0 {
		price = amountIn / amountOut
	}

	if amountIn > 0 || amountOut > 0 {
		return &RPCSwapEvent{
			Signature: signature,
			Timestamp: blockTime,
			Slot:      slot,
			AmountIn:  amountIn,
			AmountOut: amountOut,
			Price:     price,
			Direction: direction,
			TokenIn:   mint, // Simplified
			TokenOut:  mint, // Simplified
			RawLog:    logLine,
			ParsedFields: map[string]interface{}{
				"amount_in":  amountIn,
				"amount_out": amountOut,
				"price":      price,
			},
		}
	}

	return nil
}

// InferSwapFromBalances infers swap events from token balance changes
func (s *SwapIndexerService) InferSwapFromBalances(signature string, slot uint64, blockTime int64, mint string, meta struct {
	LogMessages      []string `json:"logMessages"`
	PreTokenBalances []struct {
		AccountIndex  int    `json:"accountIndex"`
		Mint          string `json:"mint"`
		Owner         string `json:"owner"`
		ProgramId     string `json:"programId"`
		UiTokenAmount struct {
			Amount   string  `json:"amount"`
			Decimals int     `json:"decimals"`
			UiAmount float64 `json:"uiAmount"`
		} `json:"uiTokenAmount"`
	} `json:"preTokenBalances"`
	PostTokenBalances []struct {
		AccountIndex  int    `json:"accountIndex"`
		Mint          string `json:"mint"`
		Owner         string `json:"owner"`
		ProgramId     string `json:"programId"`
		UiTokenAmount struct {
			Amount   string  `json:"amount"`
			Decimals int     `json:"decimals"`
			UiAmount float64 `json:"uiAmount"`
		} `json:"uiTokenAmount"`
	} `json:"postTokenBalances"`
}) *RPCSwapEvent {
	// Create maps for easy lookup
	preBalances := make(map[string]float64)
	postBalances := make(map[string]float64)

	for _, balance := range meta.PreTokenBalances {
		if balance.Mint == mint {
			preBalances[balance.Owner] = balance.UiTokenAmount.UiAmount
		}
	}

	for _, balance := range meta.PostTokenBalances {
		if balance.Mint == mint {
			postBalances[balance.Owner] = balance.UiTokenAmount.UiAmount
		}
	}

	// Find significant balance changes
	var totalChange float64
	for owner, postBalance := range postBalances {
		preBalance := preBalances[owner]
		change := postBalance - preBalance
		totalChange += math.Abs(change)
	}

	// Only consider significant changes
	if totalChange < 0.001 {
		return nil
	}

	// Determine direction based on net change
	var direction string
	var amountIn, amountOut float64

	for owner, postBalance := range postBalances {
		preBalance := preBalances[owner]
		change := postBalance - preBalance

		if change > 0 {
			amountOut += change
		} else {
			amountIn += math.Abs(change)
		}
	}

	if amountIn > 0 && amountOut > 0 {
		direction = "swap"
	} else if amountIn > 0 {
		direction = "sell"
	} else if amountOut > 0 {
		direction = "buy"
	}

	// Calculate price
	var price float64
	if amountOut > 0 {
		price = amountIn / amountOut
	}

	return &RPCSwapEvent{
		Signature: signature,
		Timestamp: blockTime,
		Slot:      slot,
		AmountIn:  amountIn,
		AmountOut: amountOut,
		Price:     price,
		Direction: direction,
		TokenIn:   mint,
		TokenOut:  mint,
		RawLog:    "inferred from balance changes",
		ParsedFields: map[string]interface{}{
			"amount_in":    amountIn,
			"amount_out":   amountOut,
			"price":        price,
			"total_change": totalChange,
		},
	}
}

// aggregateToOHLCV aggregates swap events into OHLCV candles
func (s *SwapIndexerService) aggregateToOHLCV(events []RPCSwapEvent, timeframe string) []OHLCVCandle {
	if len(events) == 0 {
		return []OHLCVCandle{}
	}

	// Group events by timeframe
	timeframeMap := make(map[int64][]RPCSwapEvent)

	var interval int64
	switch timeframe {
	case "1m":
		interval = 60
	case "5m":
		interval = 300
	case "15m":
		interval = 900
	case "1h":
		interval = 3600
	case "4h":
		interval = 14400
	case "1d":
		interval = 86400
	default:
		interval = 60 // Default to 1 minute
	}

	for _, event := range events {
		// Round timestamp to interval
		roundedTime := (event.Timestamp / interval) * interval
		timeframeMap[roundedTime] = append(timeframeMap[roundedTime], event)
	}

	// Convert to candles
	var candles []OHLCVCandle
	for timestamp, events := range timeframeMap {
		candle := s.calculateOHLCV(events)
		candle.Timestamp = timestamp
		candles = append(candles, candle)
	}

	return candles
}

// calculateOHLCV calculates OHLCV values for a group of events
func (s *SwapIndexerService) calculateOHLCV(events []RPCSwapEvent) OHLCVCandle {
	if len(events) == 0 {
		return OHLCVCandle{}
	}

	var prices []float64
	var totalVolume float64

	for _, event := range events {
		if event.Price > 0 {
			prices = append(prices, event.Price)
		}
		totalVolume += event.AmountIn + event.AmountOut
	}

	if len(prices) == 0 {
		return OHLCVCandle{
			TradeCount: len(events),
			Volume:     totalVolume,
		}
	}

	// Calculate OHLC
	open := prices[0]
	close := prices[len(prices)-1]
	high := prices[0]
	low := prices[0]

	for _, price := range prices {
		if price > high {
			high = price
		}
		if price < low {
			low = price
		}
	}

	return OHLCVCandle{
		Open:       open,
		High:       high,
		Low:        low,
		Close:      close,
		Volume:     totalVolume,
		TradeCount: len(events),
	}
}

// makeRPCRequestWithRetry makes an RPC request with retry logic
func (s *SwapIndexerService) makeRPCRequestWithRetry(rpcURL string, req SolanaRPCRequest) (*SolanaRPCResponse, error) {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err := s.makeRPCRequest(rpcURL, req)
		if err == nil {
			return resp, nil
		}

		if i < maxRetries-1 {
			log.Printf("RPC request failed (attempt %d/%d): %v, retrying...", i+1, maxRetries, err)
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return nil, fmt.Errorf("all retries failed")
}

// makeRPCRequest makes a single RPC request
func (s *SwapIndexerService) makeRPCRequest(rpcURL string, req SolanaRPCRequest) (*SolanaRPCResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", rpcURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rpcResp SolanaRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	return &rpcResp, nil
}
