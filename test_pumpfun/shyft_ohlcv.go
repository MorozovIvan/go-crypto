package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"
)

const (
	shyftAPIKey   = "CEhOUGrWGGh7zX7i"
	shyftEndpoint = "https://rpc.shyft.to/?api_key=" + shyftAPIKey
)

type ShyftSignatureResp struct {
	Result []struct {
		Signature string `json:"signature"`
		BlockTime int64  `json:"blockTime"`
	} `json:"result"`
}

type ShyftTransactionResp struct {
	Result struct {
		Meta struct {
			LogMessages []string    `json:"logMessages"`
			Err         interface{} `json:"err"`
		} `json:"meta"`
		BlockTime int64 `json:"blockTime"`
	} `json:"result"`
}

type SwapEvent struct {
	Timestamp int64
	Price     float64
	Volume    float64
	AmountIn  float64
	AmountOut float64
}

type OHLCVCandle struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

func fetchSignatures(tokenMint string, max int) ([]string, error) {
	var allSigs []string
	var before string
	batchSize := 100
	for len(allSigs) < max {
		params := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "getSignaturesForAddress",
			"params": []interface{}{
				tokenMint,
				map[string]interface{}{
					"limit": batchSize,
				},
			},
		}
		if before != "" {
			params["params"].([]interface{})[1].(map[string]interface{})["before"] = before
		}
		b, _ := json.Marshal(params)
		resp, err := http.Post(shyftEndpoint, "application/json", bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
		}
		data, _ := ioutil.ReadAll(resp.Body)
		var result ShyftSignatureResp
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("JSON unmarshal error: %v, data: %s", err, string(data))
		}
		if len(result.Result) == 0 {
			break
		}
		for _, r := range result.Result {
			allSigs = append(allSigs, r.Signature)
		}
		before = result.Result[len(result.Result)-1].Signature
		if len(result.Result) < batchSize {
			break
		}
		// Avoid rate limiting
		time.Sleep(200 * time.Millisecond)
	}
	if len(allSigs) > max {
		allSigs = allSigs[:max]
	}
	return allSigs, nil
}

func fetchTransaction(signature string) (*ShyftTransactionResp, error) {
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTransaction",
		"params": []interface{}{signature, map[string]interface{}{
			"encoding":                       "json",
			"maxSupportedTransactionVersion": 0,
		}},
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(shyftEndpoint, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result ShyftTransactionResp
	data, _ := ioutil.ReadAll(resp.Body)

	// Debug: Print raw response for first transaction
	if signature == "5Aurhvhc9dkQ1oz9G844tDN2Mf7WyiJC2Pq6t3GpabKqdMXgtkoE2zz1A87Dn5MLu1FfNCVBb4Ex1ED1ympk9iue" { // First transaction signature
		fmt.Printf("Raw transaction response for %s: %s\n", signature, string(data))
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func parseSwapFromLog(logMsg string, blockTime int64) (SwapEvent, bool) {
	// Try multiple regex patterns for different log formats
	patterns := []*regexp.Regexp{
		// Pattern 1: "SwapEvent { dex: PumpfunammSell, amount_in: 4932559599662, amount_out: 120357777 }"
		regexp.MustCompile(`SwapEvent\s*\{\s*dex:\s*(\w+),\s*amount_in:\s*(\d+),\s*amount_out:\s*(\d+)\s*\}`),
		// Pattern 2: "Dex::Pumpfunamm amount_in: 4932559599662, offset: 0"
		regexp.MustCompile(`Dex::Pumpfunamm\s+amount_in:\s*(\d+),\s*offset:\s*(\d+)`),
		// Pattern 3: "swapped X TOKEN1 for Y TOKEN2"
		regexp.MustCompile(`(?i)swapped\s+([\d,]+\\.?\\d*)\s+([A-Z0-9]+)\s+for\s+([\d,]+\\.?\\d*)\s+([A-Z0-9]+)`),
		// Pattern 4: "Swap: X TOKEN1 -> Y TOKEN2"
		regexp.MustCompile(`(?i)swap[:\\s]+([\d,]+\\.?\\d*)\s+([A-Z0-9]+)\s*[->→]\\s*([\d,]+\\.?\\d*)\s+([A-Z0-9]+)`),
	}

	for i, pattern := range patterns {
		matches := pattern.FindStringSubmatch(logMsg)
		if len(matches) > 0 {
			fmt.Printf("✅ Pattern %d matched: %s\n", i+1, logMsg)

			// Handle different pattern formats
			switch i {
			case 0: // SwapEvent pattern
				if len(matches) >= 4 {
					amountIn := parseFloat(matches[2])
					amountOut := parseFloat(matches[3])
					if amountIn > 0 && amountOut > 0 {
						price := amountIn / amountOut
						volume := amountIn
						return SwapEvent{
							Timestamp: blockTime,
							Price:     price,
							Volume:    volume,
							AmountIn:  amountIn,
							AmountOut: amountOut,
						}, true
					}
				}
			case 1: // Dex::Pumpfunamm pattern
				if len(matches) >= 3 {
					amountIn := parseFloat(matches[1])
					if amountIn > 0 {
						// For this pattern, we only have amount_in, so we'll use a placeholder price
						price := 1.0 // Placeholder
						volume := amountIn
						return SwapEvent{
							Timestamp: blockTime,
							Price:     price,
							Volume:    volume,
							AmountIn:  amountIn,
							AmountOut: 0, // Unknown
						}, true
					}
				}
			case 2, 3: // Traditional swap patterns
				if len(matches) >= 5 {
					amount1 := parseFloat(matches[1])
					amount2 := parseFloat(matches[3])
					if amount1 > 0 && amount2 > 0 {
						price := amount1 / amount2
						volume := amount1
						return SwapEvent{
							Timestamp: blockTime,
							Price:     price,
							Volume:    volume,
							AmountIn:  amount1,
							AmountOut: amount2,
						}, true
					}
				}
			}
		}
	}
	return SwapEvent{}, false
}

func parseFloat(s string) float64 {
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
}

func generateOHLCVCandles(swapEvents []SwapEvent, interval time.Duration) []OHLCVCandle {
	if len(swapEvents) == 0 {
		return []OHLCVCandle{}
	}

	// Sort events by timestamp
	// (assuming they're already sorted from API)

	// Group events by time interval
	candleMap := make(map[int64]*OHLCVCandle)

	for _, event := range swapEvents {
		// Round timestamp to interval
		intervalStart := (event.Timestamp / int64(interval.Seconds())) * int64(interval.Seconds())

		candle, exists := candleMap[intervalStart]
		if !exists {
			candle = &OHLCVCandle{
				Timestamp: intervalStart,
				Open:      event.Price,
				High:      event.Price,
				Low:       event.Price,
				Close:     event.Price,
				Volume:    event.Volume,
			}
			candleMap[intervalStart] = candle
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
		}
	}

	// Convert map to slice and sort by timestamp
	var candles []OHLCVCandle
	for _, candle := range candleMap {
		candles = append(candles, *candle)
	}

	// Sort by timestamp (ascending)
	for i := 0; i < len(candles)-1; i++ {
		for j := i + 1; j < len(candles); j++ {
			if candles[i].Timestamp > candles[j].Timestamp {
				candles[i], candles[j] = candles[j], candles[i]
			}
		}
	}

	return candles
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run shyft_ohlcv.go <token_mint_address>")
		os.Exit(1)
	}

	tokenMint := os.Args[1]
	fmt.Printf("🔍 Fetching transactions for token: %s\n", tokenMint)

	// Fetch up to 1000 signatures
	signatures, err := fetchSignatures(tokenMint, 1000)
	if err != nil {
		fmt.Printf("❌ Error fetching signatures: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📋 Found %d signatures\n", len(signatures))

	// Fetch transaction details
	var swapEvents []SwapEvent
	for i, sig := range signatures {
		tx, err := fetchTransaction(sig)
		if err != nil {
			fmt.Printf("⚠️  Error fetching transaction %s: %v\n", sig, err)
			continue
		}
		if tx.Result.Meta.Err != nil {
			continue
		}
		if i < 3 {
			fmt.Printf("\n📋 Transaction %d (%s):\n", i+1, sig[:8]+"...")
			fmt.Printf("   Block time: %s\n", time.Unix(tx.Result.BlockTime, 0).Format("2006-01-02 15:04:05"))
			fmt.Printf("   Log messages (%d):\n", len(tx.Result.Meta.LogMessages))
			for j, logMsg := range tx.Result.Meta.LogMessages {
				if j < 5 {
					fmt.Printf("     %d: %s\n", j+1, logMsg)
				}
			}
		}
		for _, logMsg := range tx.Result.Meta.LogMessages {
			if swapEvent, found := parseSwapFromLog(logMsg, tx.Result.BlockTime); found {
				swapEvents = append(swapEvents, swapEvent)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\n🔄 Found %d swap events\n", len(swapEvents))

	for i, event := range swapEvents {
		fmt.Printf("  %d. %s | Price: %.6f | Volume: %.2f | In: %.2f | Out: %.2f\n",
			i+1,
			time.Unix(event.Timestamp, 0).Format("15:04:05"),
			event.Price,
			event.Volume,
			event.AmountIn,
			event.AmountOut,
		)
	}

	// Generate OHLCV candles (1-second intervals)
	candles := generateOHLCVCandles(swapEvents, time.Second)

	fmt.Printf("\n📈 Generated %d OHLCV candles (1-second intervals)\n\n", len(candles))

	for i, candle := range candles {
		fmt.Printf("  %d. %s | O: %.6f | H: %.6f | L: %.6f | C: %.6f | V: %.2f\n",
			i+1,
			time.Unix(candle.Timestamp, 0).Format("15:04:05"),
			candle.Open,
			candle.High,
			candle.Low,
			candle.Close,
			candle.Volume,
		)
	}

	if len(candles) > 0 {
		fmt.Printf("\n📊 Summary:\n")
		fmt.Printf("  Time range: %s to %s\n",
			time.Unix(candles[0].Timestamp, 0).Format("2006-01-02 15:04:05"),
			time.Unix(candles[len(candles)-1].Timestamp, 0).Format("2006-01-02 15:04:05"),
		)
		fmt.Printf("  Total candles: %d\n", len(candles))
		fmt.Printf("  Total volume: %.2f\n", func() float64 {
			total := 0.0
			for _, c := range candles {
				total += c.Volume
			}
			return total
		}())
	}
}
