package token

import (
	"fmt"
	"go-vue/pkg/market"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// TODO: Move actual implementation from main.go

func handleTokenAnalysisWithData(c *gin.Context) {
	tokenAddress := c.Query("address")
	period := c.Query("period")
	resolveOnly := c.Query("resolve_only") == "true"

	if period == "" {
		period = "24h"
	}

	if tokenAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing token address",
			"success": false,
		})
		return
	}

	log.Printf("Starting real token analysis for: %s (period: %s, resolve_only: %v)", tokenAddress, period, resolveOnly)

	// Detect token type
	addressType := detectTokenType(tokenAddress)

	// For resolve_only requests, return basic address info
	if resolveOnly {
		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"mint_address": tokenAddress,
			"address_type": addressType,
		})
		return
	}

	// Fetch real OHLCV data using the smart combiner
	timeframe := "5m" // 5-minute candles for detailed analysis
	log.Printf("Fetching real OHLCV data for %s with period %s and timeframe %s", tokenAddress, period, timeframe)

	ohlcvData, err := market.GetSmartOHLCVData(tokenAddress, period, timeframe)
	if err != nil {
		log.Printf("Failed to fetch OHLCV data for %s: %v", tokenAddress, err)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   fmt.Sprintf("Failed to fetch token data: %v", err),
			"success": false,
		})
		return
	}

	if len(ohlcvData) == 0 {
		log.Printf("No OHLCV data available for %s", tokenAddress)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "No price data available for this token",
			"success": false,
		})
		return
	}

	log.Printf("Successfully fetched %d OHLCV candles for %s", len(ohlcvData), tokenAddress)

	// Convert OHLCV data to chart format
	prices := make([]float64, len(ohlcvData))
	timestamps := make([]string, len(ohlcvData))
	volumes := make([]float64, len(ohlcvData))
	opens := make([]float64, len(ohlcvData))
	highs := make([]float64, len(ohlcvData))
	lows := make([]float64, len(ohlcvData))
	closes := make([]float64, len(ohlcvData))

	for i, candle := range ohlcvData {
		prices[i] = candle.Close
		closes[i] = candle.Close
		opens[i] = candle.Open
		highs[i] = candle.High
		lows[i] = candle.Low
		volumes[i] = candle.Volume

		// Parse timestamp and format for display
		if t, err := time.Parse(time.RFC3339, candle.Timestamp); err == nil {
			timestamps[i] = t.Format("15:04")
		} else {
			timestamps[i] = fmt.Sprintf("T%d", i)
		}
	}

	// Calculate analysis metrics from real data
	entryPrice := opens[0]
	currentPrice := closes[len(closes)-1]

	maxPrice := highs[0]
	minPrice := lows[0]
	totalVolume := 0.0

	for i, high := range highs {
		if high > maxPrice {
			maxPrice = high
		}
		if lows[i] < minPrice {
			minPrice = lows[i]
		}
		totalVolume += volumes[i]
	}

	// Calculate ROI and other metrics
	roiPercent := 0.0
	if entryPrice > 0 {
		roiPercent = ((currentPrice - entryPrice) / entryPrice) * 100
	}

	// Determine if analysis is useful based on real price action
	priceRange := 0.0
	if minPrice > 0 {
		priceRange = (maxPrice - minPrice) / minPrice
	}
	isUseful := priceRange > 0.02 && totalVolume > 100 // 2% price movement and some volume

	// Try to get token metadata
	moralisService := market.NewMoralisService()
	tokenMetadata, err := moralisService.GetTokenMetadata(tokenAddress)

	var tokenName, tokenSymbol string
	var estimatedSupply float64 = 1000000000.0 // Default 1B tokens

	if err == nil && tokenMetadata != nil {
		tokenName = tokenMetadata.Name
		tokenSymbol = tokenMetadata.Symbol
		if tokenMetadata.TotalSupply > 0 {
			estimatedSupply = tokenMetadata.TotalSupply
		}
	} else {
		log.Printf("Could not fetch token metadata for %s: %v", tokenAddress, err)
		tokenName = "Unknown Token"
		tokenSymbol = "UNK"
	}

	// Calculate market cap estimates
	entryMarketCap := entryPrice * estimatedSupply
	maxMarketCap := maxPrice * estimatedSupply
	currentMarketCap := currentPrice * estimatedSupply

	// Determine optimal exit point (90% of max price)
	optimalExit := maxPrice * 0.9

	// Calculate additional metrics
	avgVolume := totalVolume / float64(len(volumes))
	priceVolatility := priceRange * 100 // As percentage

	result := map[string]interface{}{
		"success":          true,
		"address_type":     addressType,
		"original_address": tokenAddress,
		"mint_address":     tokenAddress,
		"analysis_period":  period,
		"analyzed_at":      time.Now().Format(time.RFC3339),
		"status":           "analysis_complete",
		"message":          fmt.Sprintf("Real token analysis completed with %d data points", len(ohlcvData)),
		"data_points":      len(ohlcvData),
		"chart_data": map[string]interface{}{
			"prices":     prices,
			"timestamps": timestamps,
			"volumes":    volumes,
			"opens":      opens,
			"highs":      highs,
			"lows":       lows,
			"closes":     closes,
		},
		"analysis": map[string]interface{}{
			"entry_price":        entryPrice,
			"current_price":      currentPrice,
			"max_price":          maxPrice,
			"min_price":          minPrice,
			"entry_market_cap":   entryMarketCap,
			"current_market_cap": currentMarketCap,
			"max_market_cap":     maxMarketCap,
			"roi_percent":        roiPercent,
			"is_useful":          isUseful,
			"optimal_exit":       optimalExit,
			"total_volume":       totalVolume,
			"avg_volume":         avgVolume,
			"price_volatility":   priceVolatility,
			"price_range_pct":    priceRange * 100,
		},
		"metadata": map[string]interface{}{
			"name":             tokenName,
			"symbol":           tokenSymbol,
			"description":      fmt.Sprintf("Real-time analysis of %s token", tokenSymbol),
			"estimated_supply": estimatedSupply,
		},
	}

	log.Printf("Completed real analysis for %s (%s): ROI=%.2f%%, Volatility=%.2f%%, Data Points=%d",
		tokenAddress, tokenSymbol, roiPercent, priceVolatility, len(ohlcvData))
	c.JSON(http.StatusOK, result)
}

// detectTokenType determines the type of token address
func detectTokenType(address string) string {
	// Basic heuristics for token type detection
	if len(address) == 44 && !strings.Contains(address, "0x") {
		// Likely Solana address (base58, ~44 chars)
		if strings.Contains(strings.ToLower(address), "pump") {
			return "pumpfun"
		}
		return "solana"
	} else if strings.HasPrefix(address, "0x") && len(address) == 42 {
		// Ethereum address
		return "ethereum"
	}

	// Default to solana for unknown formats
	return "solana"
}
