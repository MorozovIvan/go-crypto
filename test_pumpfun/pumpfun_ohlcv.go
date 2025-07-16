package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Helius API response structures
type HeliusTransaction struct {
	Signature   string      `json:"signature"`
	Slot        int64       `json:"slot"`
	BlockTime   int64       `json:"blockTime"`
	Meta        Meta        `json:"meta"`
	Transaction Transaction `json:"transaction"`
}

type Meta struct {
	Err               interface{}    `json:"err"`
	Fee               int64          `json:"fee"`
	PreBalances       []int64        `json:"preBalances"`
	PostBalances      []int64        `json:"postBalances"`
	PreTokenBalances  []TokenBalance `json:"preTokenBalances"`
	PostTokenBalances []TokenBalance `json:"postTokenBalances"`
	LogMessages       []string       `json:"logMessages"`
	InnerInstructions []interface{}  `json:"innerInstructions"`
}

type Transaction struct {
	Message Message `json:"message"`
}

type Message struct {
	AccountKeys  []string      `json:"accountKeys"`
	Instructions []Instruction `json:"instructions"`
}

type Instruction struct {
	ProgramIDIndex int    `json:"programIdIndex"`
	Accounts       []int  `json:"accounts"`
	Data           string `json:"data"`
}

type TokenBalance struct {
	AccountIndex  int         `json:"accountIndex"`
	Mint          string      `json:"mint"`
	Owner         string      `json:"owner"`
	ProgramID     string      `json:"programId"`
	UITokenAmount TokenAmount `json:"uiTokenAmount"`
}

type TokenAmount struct {
	TokenAmount string  `json:"tokenAmount"`
	Decimals    int     `json:"decimals"`
	UIAmount    float64 `json:"uiAmount"`
	UIScale     int     `json:"uiScale"`
}

// OHLCV structures
type OHLCVCandle struct {
	Timestamp  time.Time `json:"timestamp"`
	Open       float64   `json:"open"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	Close      float64   `json:"close"`
	Volume     float64   `json:"volume"`
	TradeCount int       `json:"tradeCount"`
}

type SwapEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
	TokenIn   string    `json:"tokenIn"`
	TokenOut  string    `json:"tokenOut"`
	AmountIn  float64   `json:"amountIn"`
	AmountOut float64   `json:"amountOut"`
}

func main() {
	// Test token: 7UUUHUxGrDunmbL5g3veDMtb5ai7gCexiNLZtBJgpump
	tokenMint := "7UUUHUxGrDunmbL5g3veDMtb5ai7gCexiNLZtBJgpump"

	// Get Helius API key from environment
	heliusAPIKey := os.Getenv("HELIUS_API_KEY")
	if heliusAPIKey == "" {
		log.Fatal("HELIUS_API_KEY environment variable not set")
	}

	fmt.Printf("🔍 Fetching swap events for token: %s\n", tokenMint)
	fmt.Printf("📅 Time range: Last 24 hours\n\n")

	// Fetch recent transactions for the token
	transactions, err := fetchTokenTransactions(heliusAPIKey, tokenMint)
	if err != nil {
		log.Fatalf("Failed to fetch transactions: %v", err)
	}

	fmt.Printf("📊 Found %d transactions\n", len(transactions))

	// Parse swap events from transactions
	swapEvents, err := parseSwapEvents(transactions, tokenMint)
	if err != nil {
		log.Fatalf("Failed to parse swap events: %v", err)
	}

	fmt.Printf("🔄 Found %d swap events\n\n", len(swapEvents))

	// Print first few swap events
	for i, event := range swapEvents {
		if i >= 5 { // Show only first 5 events
			break
		}
		fmt.Printf("Swap %d:\n", i+1)
		fmt.Printf("  Time: %s\n", event.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Price: $%.8f\n", event.Price)
		fmt.Printf("  Volume: %.2f tokens\n", event.Volume)
		fmt.Printf("  Amount In: %.2f\n", event.AmountIn)
		fmt.Printf("  Amount Out: %.2f\n", event.AmountOut)
		fmt.Println()
	}

	// Generate OHLCV candles (1-second intervals)
	candles := generateOHLCVCandles(swapEvents, time.Second)

	fmt.Printf("📈 Generated %d OHLCV candles (1-second intervals)\n\n", len(candles))

	// Print first few candles
	for i, candle := range candles {
		if i >= 3 { // Show only first 3 candles
			break
		}
		fmt.Printf("Candle %d (%s):\n", i+1, candle.Timestamp.Format("15:04:05"))
		fmt.Printf("  O: $%.8f  H: $%.8f  L: $%.8f  C: $%.8f\n",
			candle.Open, candle.High, candle.Low, candle.Close)
		fmt.Printf("  Volume: %.2f  Trades: %d\n", candle.Volume, candle.TradeCount)
		fmt.Println()
	}
}

func fetchTokenTransactions(apiKey, tokenMint string) ([]HeliusTransaction, error) {
	url := fmt.Sprintf("https://mainnet.helius-rpc.com/?api-key=%s", apiKey)
	var allTransactions []HeliusTransaction
	var before string
	maxBatches := 20 // 20 * 50 = 1000 max
	timeLimit := time.Now().Add(-24 * time.Hour).Unix()

	for batch := 0; batch < maxBatches; batch++ {
		params := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "getSignaturesForAddress",
			"params": []interface{}{
				tokenMint,
				map[string]interface{}{
					"limit": 50,
				},
			},
		}
		if before != "" {
			params["params"].([]interface{})[1].(map[string]interface{})["before"] = before
		}

		jsonData, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %v", err)
		}

		// Make HTTP request
		var resp *http.Response
		for attempt := 1; attempt <= 3; attempt++ {
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{Timeout: 30 * time.Second}
			resp, err = client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				break
			}
			if resp != nil {
				resp.Body.Close()
			}
			if resp != nil && resp.StatusCode == 429 {
				fmt.Printf("Rate limited, sleeping before retry...\n")
				time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
				continue
			}
			fmt.Printf("Request failed (attempt %d): %v\n", attempt, err)
			time.Sleep(500 * time.Millisecond)
		}
		if resp == nil || resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API request failed after retries")
		}
		defer resp.Body.Close()

		var rpcResponse struct {
			JSONRPC string `json:"jsonrpc"`
			ID      int    `json:"id"`
			Result  []struct {
				Signature string      `json:"signature"`
				Slot      int64       `json:"slot"`
				Err       interface{} `json:"err"`
				Memo      interface{} `json:"memo"`
				BlockTime int64       `json:"blockTime"`
			} `json:"result"`
			Error *struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error,omitempty"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
			return nil, fmt.Errorf("failed to decode response: %v", err)
		}
		if rpcResponse.Error != nil {
			return nil, fmt.Errorf("RPC error: %d - %s", rpcResponse.Error.Code, rpcResponse.Error.Message)
		}
		if len(rpcResponse.Result) == 0 {
			break // No more signatures
		}

		for _, sig := range rpcResponse.Result {
			if sig.Err != nil {
				continue // Skip failed transactions
			}
			if sig.BlockTime < timeLimit {
				return allTransactions, nil // Stop if older than 24h
			}
			fullTx, err := fetchTransactionDetails(apiKey, sig.Signature)
			if err != nil {
				fmt.Printf("Warning: Failed to fetch transaction %s: %v\n", sig.Signature, err)
				continue
			}
			allTransactions = append(allTransactions, fullTx)
			before = sig.Signature
		}
		// Sleep between batches to avoid rate limiting
		time.Sleep(200 * time.Millisecond)
	}
	return allTransactions, nil
}

func fetchTransactionDetails(apiKey, signature string) (HeliusTransaction, error) {
	// Use Helius RPC endpoint to get full transaction details
	url := fmt.Sprintf("https://mainnet.helius-rpc.com/?api-key=%s", apiKey)

	// Use getTransaction RPC method
	params := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTransaction",
		"params": []interface{}{
			signature,
			map[string]interface{}{
				"encoding":                       "json",
				"maxSupportedTransactionVersion": 0,
			},
		},
	}

	jsonData, err := json.Marshal(params)
	if err != nil {
		return HeliusTransaction{}, fmt.Errorf("failed to marshal params: %v", err)
	}

	// Make HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return HeliusTransaction{}, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return HeliusTransaction{}, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return HeliusTransaction{}, fmt.Errorf("API request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse RPC response
	var rpcResponse struct {
		JSONRPC string             `json:"jsonrpc"`
		ID      int                `json:"id"`
		Result  *HeliusTransaction `json:"result"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
		return HeliusTransaction{}, fmt.Errorf("failed to decode response: %v", err)
	}

	if rpcResponse.Error != nil {
		return HeliusTransaction{}, fmt.Errorf("RPC error: %d - %s", rpcResponse.Error.Code, rpcResponse.Error.Message)
	}

	if rpcResponse.Result == nil {
		return HeliusTransaction{}, fmt.Errorf("no transaction found for signature: %s", signature)
	}

	return *rpcResponse.Result, nil
}

func parseSwapEvents(transactions []HeliusTransaction, tokenMint string) ([]SwapEvent, error) {
	var swapEvents []SwapEvent

	fmt.Printf("🔍 Analyzing %d transactions for swap events...\n", len(transactions))

	for i, tx := range transactions {
		// Skip failed transactions
		if tx.Meta.Err != nil {
			continue
		}

		// Debug: Print first few log messages from each transaction
		if i < 3 { // Only show first 3 transactions for debugging
			shortSig := tx.Signature
			if len(shortSig) > 8 {
				shortSig = shortSig[:8] + "..."
			}
			fmt.Printf("\n📋 Transaction %d (%s):\n", i+1, shortSig)
			fmt.Printf("   Block time: %s\n", time.Unix(tx.BlockTime, 0).Format("2006-01-02 15:04:05"))
			fmt.Printf("   Log messages (%d):\n", len(tx.Meta.LogMessages))
			for j, logMsg := range tx.Meta.LogMessages {
				if j < 5 { // Show first 5 log messages
					fmt.Printf("     %d: %s\n", j+1, logMsg)
				}
			}
		}

		// Look for swap events in log messages
		for _, logMsg := range tx.Meta.LogMessages {
			// Check if this is a swap event with more flexible patterns
			if isSwapLog(logMsg, tokenMint) {
				event, err := parseSwapFromLog(logMsg, tx.BlockTime, tx.Meta)
				if err != nil {
					fmt.Printf("   ⚠️  Failed to parse swap log: %v\n", err)
					continue // Skip if we can't parse this log
				}
				swapEvents = append(swapEvents, event)
				fmt.Printf("   ✅ Found swap event: %.8f tokens at $%.8f\n", event.Volume, event.Price)
			}
		}
	}

	return swapEvents, nil
}

func isSwapLog(logMsg, tokenMint string) bool {
	// More flexible patterns for pump.fun swap detection
	patterns := []string{
		"swap",
		"Swap",
		"SWAP",
		"pump",
		"Pump",
		"PUMP",
		"amm",
		"AMM",
		"trade",
		"Trade",
		"TRADE",
	}

	// Check if log contains any swap-related keywords
	for _, pattern := range patterns {
		if strings.Contains(logMsg, pattern) {
			return true
		}
	}

	// Also check if it contains the token mint address
	if strings.Contains(logMsg, tokenMint) {
		return true
	}

	return false
}

func parseSwapFromLog(logMsg string, blockTime int64, meta Meta) (SwapEvent, error) {
	// Try multiple regex patterns for different log formats
	patterns := []*regexp.Regexp{
		// Pattern 1: "swapped X TOKEN1 for Y TOKEN2"
		regexp.MustCompile(`(?i)swapped\s+([\d,]+\.?\d*)\s+([A-Z0-9]+)\s+for\s+([\d,]+\.?\d*)\s+([A-Z0-9]+)`),
		// Pattern 2: "Swap: X TOKEN1 -> Y TOKEN2"
		regexp.MustCompile(`(?i)swap[:\s]+([\d,]+\.?\d*)\s+([A-Z0-9]+)\s*[->→]\s*([\d,]+\.?\d*)\s+([A-Z0-9]+)`),
		// Pattern 3: "Trade: X TOKEN1 for Y TOKEN2"
		regexp.MustCompile(`(?i)trade[:\s]+([\d,]+\.?\d*)\s+([A-Z0-9]+)\s+for\s+([\d,]+\.?\d*)\s+([A-Z0-9]+)`),
		// Pattern 4: "Amount: X, Token: Y"
		regexp.MustCompile(`(?i)amount[:\s]+([\d,]+\.?\d*).*token[:\s]+([A-Z0-9]+)`),
		// Pattern 5: Generic number patterns
		regexp.MustCompile(`([\d,]+\.?\d*)\s+([A-Z0-9]+)`),
		// Pattern 6: Pump.fun SwapEvent struct
		regexp.MustCompile(`(?i)SwapEvent \{[^}]*amount_in: ([\d,]+)\, amount_out: ([\d,]+)`),
		// Pattern 7: Dex::Pumpfunamm amount_in
		regexp.MustCompile(`(?i)Dex::Pumpfunamm amount_in: ([\d,]+)`),
	}

	for i, pattern := range patterns {
		matches := pattern.FindStringSubmatch(logMsg)
		if len(matches) >= 3 {
			fmt.Printf("   🎯 Pattern %d matched: %s\n", i+1, logMsg)

			var amountIn, amountOut float64
			var tokenIn, tokenOut string

			switch i {
			case 0, 1, 2: // Patterns with 4+ matches
				if len(matches) >= 5 {
					amountIn, _ = strconv.ParseFloat(strings.ReplaceAll(matches[1], ",", ""), 64)
					tokenIn = matches[2]
					amountOut, _ = strconv.ParseFloat(strings.ReplaceAll(matches[3], ",", ""), 64)
					tokenOut = matches[4]
				}
			case 3: // Amount/Token pattern
				if len(matches) >= 3 {
					amountIn, _ = strconv.ParseFloat(strings.ReplaceAll(matches[1], ",", ""), 64)
					tokenIn = matches[2]
					amountOut = amountIn // Assume same for now
					tokenOut = tokenIn
				}
			case 4: // Generic number pattern
				if len(matches) >= 3 {
					amountIn, _ = strconv.ParseFloat(strings.ReplaceAll(matches[1], ",", ""), 64)
					tokenIn = matches[2]
					amountOut = amountIn // Assume same for now
					tokenOut = tokenIn
				}
			case 5: // Pump.fun SwapEvent struct
				if len(matches) >= 3 {
					amountIn, _ = strconv.ParseFloat(strings.ReplaceAll(matches[1], ",", ""), 64)
					amountOut, _ = strconv.ParseFloat(strings.ReplaceAll(matches[2], ",", ""), 64)
					tokenIn = "SOL/USDC" // Unknown, but usually input is SOL or USDC
					tokenOut = "PUMP_TOKEN"
				}
			case 6: // Dex::Pumpfunamm amount_in
				if len(matches) >= 2 {
					amountIn, _ = strconv.ParseFloat(strings.ReplaceAll(matches[1], ",", ""), 64)
					amountOut = 0 // Only amount_in available
					tokenIn = "SOL/USDC"
					tokenOut = "PUMP_TOKEN"
				}
			}

			if amountIn > 0 {
				// Calculate price if possible
				price := 0.0
				if amountOut > 0 {
					price = amountIn / amountOut
				}
				timestamp := time.Unix(blockTime, 0)
				return SwapEvent{
					Timestamp: timestamp,
					Price:     price,
					Volume:    amountOut,
					TokenIn:   tokenIn,
					TokenOut:  tokenOut,
					AmountIn:  amountIn,
					AmountOut: amountOut,
				}, nil
			}
		}
	}

	return SwapEvent{}, fmt.Errorf("no matching pattern found in log: %s", logMsg)
}

func generateOHLCVCandles(swapEvents []SwapEvent, interval time.Duration) []OHLCVCandle {
	if len(swapEvents) == 0 {
		return []OHLCVCandle{}
	}

	// Group events by time interval
	candleMap := make(map[time.Time]*OHLCVCandle)

	for _, event := range swapEvents {
		// Round timestamp to interval
		candleTime := event.Timestamp.Truncate(interval)

		candle, exists := candleMap[candleTime]
		if !exists {
			candle = &OHLCVCandle{
				Timestamp:  candleTime,
				Open:       event.Price,
				High:       event.Price,
				Low:        event.Price,
				Close:      event.Price,
				Volume:     event.Volume,
				TradeCount: 1,
			}
			candleMap[candleTime] = candle
		} else {
			// Update OHLCV
			if event.Price > candle.High {
				candle.High = event.Price
			}
			if event.Price < candle.Low {
				candle.Low = event.Price
			}
			candle.Close = event.Price
			candle.Volume += event.Volume
			candle.TradeCount++
		}
	}

	// Convert map to slice and sort by timestamp
	var candles []OHLCVCandle
	for _, candle := range candleMap {
		candles = append(candles, *candle)
	}

	// Sort by timestamp (you can implement this if needed)

	return candles
}
