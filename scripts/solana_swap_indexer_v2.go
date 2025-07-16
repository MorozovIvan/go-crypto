package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SwapEvent represents a parsed swap event from transaction logs
type SwapEvent struct {
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

// RPC endpoints with fallback
var rpcEndpoints = []string{
	"https://rpc.helius.xyz/?api-key=YOUR_API_KEY", // Replace with your Helius API key
	"https://solana-mainnet.rpc.extrnode.com",
	"https://api.mainnet-beta.solana.com",
	"https://solana-api.projectserum.com",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <mint_address> [rpc_url]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s 7UUUHUxGrDunmbL5g3veDMtb5ai7gCexiNLZtBJgpump\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Note: Set HELIUS_API_KEY environment variable for better rate limits\n")
		os.Exit(1)
	}

	mint := os.Args[1]

	// Use provided RPC URL or get from environment or use fallback
	rpcURL := getRPCURL()

	log.Printf("Indexing swap events for mint: %s", mint)
	log.Printf("Using RPC: %s", rpcURL)

	// Get recent signatures for the mint address
	signatures, err := getSignaturesForAddress(mint, rpcURL)
	if err != nil {
		log.Fatalf("Failed to get signatures: %v", err)
	}

	log.Printf("Found %d recent transactions", len(signatures))

	var allSwapEvents []SwapEvent
	var processedCount int

	// Process each transaction with better rate limiting
	for i, sig := range signatures {
		if i >= 20 { // Limit to last 20 transactions for performance
			break
		}

		events, err := parseTransactionForSwaps(sig, mint, rpcURL)
		if err != nil {
			log.Printf("Error parsing transaction %s: %v", sig, err)
			continue
		}

		if len(events) > 0 {
			allSwapEvents = append(allSwapEvents, events...)
			log.Printf("Found %d swap events in transaction %s", len(events), sig)
		}

		processedCount++
		if processedCount%5 == 0 {
			log.Printf("Processed %d/%d transactions", processedCount, len(signatures))
		}

		// More aggressive rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("Total swap events found: %d", len(allSwapEvents))

	// Output raw swap events
	fmt.Println("=== RAW SWAP EVENTS ===")
	for _, event := range allSwapEvents {
		b, _ := json.Marshal(event)
		fmt.Println(string(b))
	}

	// Aggregate into OHLCV candles
	if len(allSwapEvents) > 0 {
		fmt.Println("\n=== OHLCV CANDLES ===")
		candles := aggregateToOHLCV(allSwapEvents, "1m") // 1-minute candles
		for _, candle := range candles {
			b, _ := json.Marshal(candle)
			fmt.Println(string(b))
		}
	}
}

// getRPCURL returns the best available RPC URL
func getRPCURL() string {
	// Check for custom RPC URL in args
	if len(os.Args) > 2 {
		return os.Args[2]
	}

	// Check for Helius API key in environment
	if apiKey := os.Getenv("HELIUS_API_KEY"); apiKey != "" {
		return fmt.Sprintf("https://rpc.helius.xyz/?api-key=%s", apiKey)
	}

	// Use fallback endpoints
	return rpcEndpoints[2] // Default to Solana mainnet
}

// getSignaturesForAddress fetches recent transaction signatures for a mint address
func getSignaturesForAddress(address, rpcURL string) ([]string, error) {
	req := SolanaRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "getSignaturesForAddress",
		Params: []interface{}{
			address,
			map[string]interface{}{
				"limit": 50, // Reduced limit
			},
		},
	}

	resp, err := makeRPCRequestWithRetry(rpcURL, req)
	if err != nil {
		return nil, err
	}

	var result []struct {
		Signature string `json:"signature"`
		Slot      uint64 `json:"slot"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal signatures: %v", err)
	}

	var signatures []string
	for _, r := range result {
		signatures = append(signatures, r.Signature)
	}

	return signatures, nil
}

// parseTransactionForSwaps fetches transaction details and parses swap events
func parseTransactionForSwaps(signature, mint, rpcURL string) ([]SwapEvent, error) {
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

	resp, err := makeRPCRequestWithRetry(rpcURL, req)
	if err != nil {
		return nil, err
	}

	var tx struct {
		Slot      uint64 `json:"slot"`
		BlockTime int64  `json:"blockTime"`
		Meta      struct {
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
		Transaction struct {
			Message struct {
				AccountKeys []string `json:"accountKeys"`
			} `json:"message"`
		} `json:"transaction"`
	}

	if err := json.Unmarshal(resp.Result, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %v", err)
	}

	var events []SwapEvent

	// Check if this transaction involves the pump.fun AMM program
	involvesAMM := false
	for _, key := range tx.Transaction.Message.AccountKeys {
		if key == PumpFunAMMProgram {
			involvesAMM = true
			break
		}
	}

	if !involvesAMM {
		return events, nil
	}

	// Parse logs for swap events
	for _, logLine := range tx.Meta.LogMessages {
		// Look for pump.fun AMM swap-related log messages
		if strings.Contains(logLine, "Program PFundAMM111111111111111111111111111111111111 invoke") ||
			strings.Contains(logLine, "Program log: Instruction: Swap") ||
			strings.Contains(logLine, "Program log: Swap") ||
			strings.Contains(logLine, "amount") ||
			strings.Contains(logLine, "token") {

			// Try to parse swap details from the log
			event := ParseSwapFromLog(logLine, signature, tx.Slot, tx.BlockTime, mint, tx.Meta)
			if event != nil {
				events = append(events, *event)
			}
		}
	}

	// If no events found in logs, try to infer from token balance changes
	if len(events) == 0 {
		event := InferSwapFromBalances(signature, tx.Slot, tx.BlockTime, mint, tx.Meta)
		if event != nil {
			events = append(events, *event)
		}
	}

	return events, nil
}

// parseSwapFromLog attempts to extract swap details from a log message
func ParseSwapFromLog(logLine, signature string, slot uint64, blockTime int64, mint string, meta struct {
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
}) *SwapEvent {

	event := &SwapEvent{
		Signature:    signature,
		Slot:         slot,
		Timestamp:    blockTime,
		RawLog:       logLine,
		ParsedFields: make(map[string]interface{}),
	}

	// Enhanced parsing patterns for pump.fun AMM logs
	patterns := []struct {
		name  string
		regex string
	}{
		{"amount_in", `amount[_\s]?in[:\s]*([0-9]+(?:\.[0-9]+)?)`},
		{"amount_out", `amount[_\s]?out[:\s]*([0-9]+(?:\.[0-9]+)?)`},
		{"token_in", `token[_\s]?in[:\s]*([A-Za-z0-9]{32,44})`},
		{"token_out", `token[_\s]?out[:\s]*([A-Za-z0-9]{32,44})`},
		{"price", `price[:\s]*([0-9]+(?:\.[0-9]+)?)`},
		{"fee", `fee[:\s]*([0-9]+(?:\.[0-9]+)?)`},
		{"swap_amount", `swap[:\s]*([0-9]+(?:\.[0-9]+)?)`},
		{"input_amount", `input[:\s]*([0-9]+(?:\.[0-9]+)?)`},
		{"output_amount", `output[:\s]*([0-9]+(?:\.[0-9]+)?)`},
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern.regex)
		matches := re.FindStringSubmatch(logLine)
		if len(matches) > 1 {
			value := matches[1]
			event.ParsedFields[pattern.name] = value

			// Convert numeric values
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				switch pattern.name {
				case "amount_in", "input_amount", "swap_amount":
					event.AmountIn = num
				case "amount_out", "output_amount":
					event.AmountOut = num
				case "price":
					event.Price = num
				case "fee":
					event.FeeAmount = num
				}
			} else {
				// Handle token addresses
				switch pattern.name {
				case "token_in":
					event.TokenIn = value
				case "token_out":
					event.TokenOut = value
				}
			}
		}
	}

	// Determine swap direction
	if event.TokenIn == mint {
		event.Direction = "sell"
	} else if event.TokenOut == mint {
		event.Direction = "buy"
	}

	// Calculate price if we have both amounts
	if event.AmountIn > 0 && event.AmountOut > 0 && event.Price == 0 {
		if event.Direction == "buy" {
			event.Price = event.AmountIn / event.AmountOut
		} else {
			event.Price = event.AmountOut / event.AmountIn
		}
	}

	// Only return event if we found meaningful data
	if event.AmountIn > 0 || event.AmountOut > 0 || event.Price > 0 {
		return event
	}

	return nil
}

// inferSwapFromBalances attempts to infer swap details from token balance changes
func InferSwapFromBalances(signature string, slot uint64, blockTime int64, mint string, meta struct {
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
}) *SwapEvent {

	// Create maps for easy lookup
	preBalances := make(map[string]float64)
	postBalances := make(map[string]float64)

	for _, balance := range meta.PreTokenBalances {
		preBalances[balance.Mint] = balance.UiTokenAmount.UiAmount
	}

	for _, balance := range meta.PostTokenBalances {
		postBalances[balance.Mint] = balance.UiTokenAmount.UiAmount
	}

	// Find balance changes
	var tokenIn, tokenOut string
	var amountIn, amountOut float64

	for mintAddr, postBalance := range postBalances {
		preBalance := preBalances[mintAddr]
		change := postBalance - preBalance

		if math.Abs(change) > 0.000001 { // Significant change
			if change > 0 {
				// Token received
				if tokenOut == "" {
					tokenOut = mintAddr
					amountOut = change
				}
			} else {
				// Token spent
				if tokenIn == "" {
					tokenIn = mintAddr
					amountIn = math.Abs(change)
				}
			}
		}
	}

	// If we found a swap involving our target mint
	if (tokenIn == mint || tokenOut == mint) && (amountIn > 0 && amountOut > 0) {
		event := &SwapEvent{
			Signature:    signature,
			Slot:         slot,
			Timestamp:    blockTime,
			AmountIn:     amountIn,
			AmountOut:    amountOut,
			TokenIn:      tokenIn,
			TokenOut:     tokenOut,
			ParsedFields: make(map[string]interface{}),
		}

		// Determine direction
		if tokenIn == mint {
			event.Direction = "sell"
			event.Price = amountOut / amountIn
		} else {
			event.Direction = "buy"
			event.Price = amountIn / amountOut
		}

		event.ParsedFields["inferred_from_balances"] = true
		event.ParsedFields["token_in"] = tokenIn
		event.ParsedFields["token_out"] = tokenOut
		event.ParsedFields["amount_in"] = amountIn
		event.ParsedFields["amount_out"] = amountOut

		return event
	}

	return nil
}

// aggregateToOHLCV aggregates swap events into OHLCV candles
func aggregateToOHLCV(events []SwapEvent, timeframe string) []OHLCVCandle {
	if len(events) == 0 {
		return nil
	}

	// Group events by time bucket
	buckets := make(map[int64][]SwapEvent)

	for _, event := range events {
		// Convert to time bucket based on timeframe
		var bucket int64
		switch timeframe {
		case "1m":
			bucket = event.Timestamp - (event.Timestamp % 60)
		case "5m":
			bucket = event.Timestamp - (event.Timestamp % 300)
		case "1h":
			bucket = event.Timestamp - (event.Timestamp % 3600)
		default:
			bucket = event.Timestamp - (event.Timestamp % 60) // Default to 1m
		}

		buckets[bucket] = append(buckets[bucket], event)
	}

	var candles []OHLCVCandle
	for bucket, bucketEvents := range buckets {
		if len(bucketEvents) == 0 {
			continue
		}

		// Calculate OHLCV for this bucket
		candle := calculateOHLCV(bucketEvents)
		candle.Timestamp = bucket
		candles = append(candles, candle)
	}

	return candles
}

// calculateOHLCV calculates OHLCV values for a group of swap events
func calculateOHLCV(events []SwapEvent) OHLCVCandle {
	if len(events) == 0 {
		return OHLCVCandle{}
	}

	var open, high, low, close, volume float64

	// Use the first event's price as open
	open = events[0].Price
	close = events[len(events)-1].Price

	high = open
	low = open

	for _, event := range events {
		if event.Price > high {
			high = event.Price
		}
		if event.Price < low {
			low = event.Price
		}
		volume += event.AmountIn
	}

	return OHLCVCandle{
		Open:       open,
		High:       high,
		Low:        low,
		Close:      close,
		Volume:     volume,
		TradeCount: len(events),
	}
}

// makeRPCRequestWithRetry makes a JSON-RPC request with retry logic
func makeRPCRequestWithRetry(rpcURL string, req SolanaRPCRequest) (*SolanaRPCResponse, error) {
	maxRetries := 3
	baseDelay := 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := makeRPCRequest(rpcURL, req)
		if err != nil {
			// Check if it's a rate limit error
			if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "Too many requests") {
				if attempt < maxRetries-1 {
					delay := time.Duration(attempt+1) * baseDelay
					log.Printf("Rate limited, retrying in %v (attempt %d/%d)", delay, attempt+1, maxRetries)
					time.Sleep(delay)
					continue
				}
			}
			return nil, err
		}
		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded")
}

// makeRPCRequest makes a JSON-RPC request to the Solana RPC endpoint
func makeRPCRequest(rpcURL string, req SolanaRPCRequest) (*SolanaRPCResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Post(rpcURL, "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var rpcResp SolanaRPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s (code: %d)", rpcResp.Error.Message, rpcResp.Error.Code)
	}

	return &rpcResp, nil
}
