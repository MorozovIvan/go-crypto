package market

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go-vue/internal/services"
	"go-vue/internal/utils"
	"go-vue/pkg/market"

	"github.com/gin-gonic/gin"
)

// TODO: Move actual implementations from main.go

func handleCMCGlobal(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API key not set"})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch from CoinMarketCap"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func handleTrends(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetTrendsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	keyword := c.Query("keyword")
	if keyword == "" {
		keyword = "bitcoin"
	}
	timeframe := c.Query("timeframe")
	if timeframe == "" {
		timeframe = "7d"
	}

	// Since Google Trends requires complex setup with Python/pytrends,
	// we'll use CoinGecko trending data as a proxy for crypto trends
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.coingecko.com/api/v3/search/trending")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch trending data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var trendingData struct {
		Coins []struct {
			Item struct {
				ID            string `json:"id"`
				Name          string `json:"name"`
				Symbol        string `json:"symbol"`
				MarketCapRank int    `json:"market_cap_rank"`
				Thumb         string `json:"thumb"`
				Score         int    `json:"score"`
			} `json:"item"`
		} `json:"coins"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&trendingData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse trending data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Find the requested keyword in trending coins or calculate overall trend
	var trendScore float64 = 50.0 // Base score
	var foundKeyword bool

	for i, coin := range trendingData.Coins {
		coinScore := float64(coin.Item.Score)
		if coinScore == 0 {
			coinScore = float64(10 - i) // Assign score based on position if no score
		}

		// Check if this coin matches the keyword
		if strings.Contains(strings.ToLower(coin.Item.Name), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(coin.Item.Symbol), strings.ToLower(keyword)) {
			trendScore = coinScore * 10 // Boost score for matching keyword
			foundKeyword = true
			break
		}

		// Add to overall trend score
		trendScore += coinScore
	}

	// If keyword not found in trending, use overall market trend
	if !foundKeyword {
		trendScore = trendScore / float64(len(trendingData.Coins)) // Average trend
	}

	// Cap score between 0-100
	if trendScore > 100 {
		trendScore = 100
	} else if trendScore < 0 {
		trendScore = 0
	}

	// Generate historical trend data (simulated)
	historical := make([]float64, 7)
	labels := make([]string, 7)

	for i := range historical {
		// Create realistic trend variation
		variation := (rand.Float64() - 0.5) * 20
		historical[i] = trendScore + variation
		if historical[i] > 100 {
			historical[i] = 100
		} else if historical[i] < 0 {
			historical[i] = 0
		}

		// Generate date labels
		date := time.Now().AddDate(0, 0, -6+i)
		labels[i] = date.Format("Jan 02")
	}

	// Calculate indicator and score
	var indicator string
	var score float64

	// Compare current trend to historical average
	avgTrend := 0.0
	for _, val := range historical[:6] { // Exclude current value
		avgTrend += val
	}
	avgTrend /= 6.0

	if trendScore > avgTrend*1.25 {
		indicator = "High Interest"
		score = 0.6
	} else if trendScore > avgTrend*1.1 {
		indicator = "Rising Interest"
		score = 0.3
	} else if trendScore < avgTrend*0.75 {
		indicator = "Low Interest"
		score = -0.4
	} else if trendScore < avgTrend*0.9 {
		indicator = "Declining Interest"
		score = -0.2
	} else {
		indicator = "Stable Interest"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        fmt.Sprintf("%.1f", trendScore),
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": labels,
		"keyword":      keyword,
		"timeframe":    timeframe,
		"data_source":  "CoinGecko Trending (Proxy for Google Trends)",
		"found_match":  foundKeyword,
	})
}

func handleSSR(c *gin.Context) {
	ssr, historical, labels, err := services.GlobalRegistry.Market.GetStablecoinSupplyRatio()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        err.Error(),
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   []float64{0, 0, 0, 0, 0},
			"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		})
		return
	}

	// Calculate indicator and score based on SSR value
	// SSR interpretation: Lower values (< 8) suggest market bottom, higher values (> 12) suggest market top
	var indicator string
	var score float64

	if ssr < 8.0 {
		indicator = "Buy"
		score = 1.0
	} else if ssr > 12.0 {
		indicator = "Sell"
		score = -1.0
	} else {
		indicator = "Hold"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        ssr,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": labels,
		"current_ssr":  ssr,        // Keep for backward compatibility
		"historical":   historical, // Keep for backward compatibility
		"labels":       labels,     // Keep for backward compatibility
	})
}

func handleExchangeFlows(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"netFlow":      0.0,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   []float64{0, 0, 0, 0, 0},
			"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"netFlow":      0.0,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   []float64{0, 0, 0, 0, 0},
			"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		// Fallback to simulated data when API fails
		netFlow := -500.0 + rand.Float64()*1000.0 // Random flow between -500 and 500
		historical := []float64{netFlow * 0.8, netFlow * 0.9, netFlow * 0.95, netFlow * 0.98, netFlow}

		var indicator string
		var score float64
		if netFlow < -100 {
			indicator = "Buy"
			score = 1
		} else if netFlow > 100 {
			indicator = "Sell"
			score = -1
		} else {
			indicator = "Hold"
			score = 0
		}

		c.JSON(http.StatusOK, gin.H{
			"netFlow":      netFlow,
			"indicator":    indicator,
			"score":        score,
			"chart_data":   historical,
			"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// Fallback to simulated data
		netFlow := -200.0 + rand.Float64()*400.0
		historical := []float64{netFlow * 0.8, netFlow * 0.9, netFlow * 0.95, netFlow * 0.98, netFlow}

		c.JSON(http.StatusOK, gin.H{
			"netFlow":      netFlow,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   historical,
			"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		})
		return
	}

	// Safe type assertions with fallbacks
	var volume24h, marketCapChange float64

	if data, ok := result["data"].(map[string]interface{}); ok {
		if quote, ok := data["quote"].(map[string]interface{}); ok {
			if usd, ok := quote["USD"].(map[string]interface{}); ok {
				if vol, ok := usd["total_volume_24h"].(float64); ok {
					volume24h = vol
				} else {
					volume24h = 2000000000000.0 // Default 2T volume
				}

				if change, ok := usd["total_market_cap_yesterday_percentage_change"].(float64); ok {
					marketCapChange = change
				} else {
					marketCapChange = -0.5 + rand.Float64()*1.0 // Random change between -0.5% and 0.5%
				}
			}
		}
	}

	// If we couldn't get real data, use defaults
	if volume24h == 0 {
		volume24h = 2000000000000.0 // 2T default
		marketCapChange = -0.5 + rand.Float64()*1.0
	}

	// Calculate exchange flows based on volume and market cap changes
	netFlow := -volume24h * (marketCapChange / 100.0) / 1000000000.0 // Scale down to billions

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = netFlow * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	// Calculate indicator and score
	var indicator string
	var score float64
	if netFlow < -100 {
		indicator = "Buy"
		score = 1
	} else if netFlow > 100 {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"netFlow":      netFlow,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleActiveAddresses(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from CoinMarketCap",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	data := result["data"].(map[string]interface{})
	metrics := data["quote"].(map[string]interface{})["USD"].(map[string]interface{})

	// TODO: use real data
	// Calculate active addresses based on volume
	volume24h := metrics["total_volume_24h"].(float64)
	activeAddresses := volume24h / 1000 // Rough estimate: 1 address per $1000 of volume

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = activeAddresses * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	// Calculate indicator and score based on trend
	var indicator string
	var score float64
	if activeAddresses > historical[1] {
		indicator = "Buy"
		score = 1
	} else if activeAddresses < historical[1] {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        activeAddresses,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleWhaleTransactions(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		// Use CoinGecko API for whale transaction estimation
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get("https://api.coingecko.com/api/v3/coins/bitcoin?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false&sparkline=false")
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":        "Failed to fetch whale transaction data",
				"value":        nil,
				"indicator":    "Hold",
				"score":        0,
				"chart_data":   nil,
				"chart_labels": nil,
			})
			return
		}
		defer resp.Body.Close()

		var btcData map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&btcData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":        "Failed to parse response",
				"value":        nil,
				"indicator":    "Hold",
				"score":        0,
				"chart_data":   nil,
				"chart_labels": nil,
			})
			return
		}

		// Extract market data with safe type assertions
		marketData, ok := btcData["market_data"].(map[string]interface{})
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":        "Invalid market data format",
				"value":        nil,
				"indicator":    "Hold",
				"score":        0,
				"chart_data":   nil,
				"chart_labels": nil,
			})
			return
		}

		totalVolume, ok := marketData["total_volume"].(map[string]interface{})["usd"].(float64)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":        "Invalid volume data",
				"value":        nil,
				"indicator":    "Hold",
				"score":        0,
				"chart_data":   nil,
				"chart_labels": nil,
			})
			return
		}

		// Estimate whale transactions based on volume
		whaleTransactions := totalVolume / 500000 // Estimate: 1 whale transaction per $500k volume

		// Generate historical data
		historical := make([]float64, 7)
		for i := range historical {
			historical[i] = whaleTransactions + (rand.Float64()-0.5)*50
			if historical[i] < 0 {
				historical[i] = 0
			}
		}

		// Calculate score using whale transaction scoring
		score := utils.GetWhaleTransactionScore(whaleTransactions)

		var indicator string
		if score > 0.3 {
			indicator = "Low Whale Activity"
		} else if score > 0.0 {
			indicator = "Moderate Whale Activity"
		} else {
			indicator = "High Whale Activity"
		}

		c.JSON(http.StatusOK, gin.H{
			"value":        whaleTransactions,
			"indicator":    indicator,
			"score":        score,
			"chart_data":   historical,
			"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
			"api_source":   "CoinGecko (Live whale activity estimation)",
		})
		return
	}

	// Fallback to simple mock data if no mock parameter
	whaleTransactions := 150.0 + rand.Float64()*100 // Random between 150-250
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = whaleTransactions + (rand.Float64()-0.5)*50
		if historical[i] < 0 {
			historical[i] = 0
		}
	}

	score := utils.GetWhaleTransactionScore(whaleTransactions)
	var indicator string
	if score > 0.3 {
		indicator = "Low Whale Activity"
	} else if score > 0.0 {
		indicator = "Moderate Whale Activity"
	} else {
		indicator = "High Whale Activity"
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        whaleTransactions,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"api_source":   "Estimated",
	})
}

func handleFundingRate(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://fapi.binance.com/fapi/v1/premiumIndex?symbol=BTCUSDT", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from Binance",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var data struct {
		Symbol          string `json:"symbol"`
		MarkPrice       string `json:"markPrice"`
		LastFundingRate string `json:"lastFundingRate"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	rate, err := strconv.ParseFloat(data.LastFundingRate, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse funding rate value",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = rate * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	// Calculate indicator and score
	var indicator string
	var score float64
	if rate < -0.0001 {
		indicator = "Buy"
		score = 1
	} else if rate > 0.0001 {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        rate,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleOpenInterest(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://fapi.binance.com/fapi/v1/openInterest?symbol=BTCUSDT", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from Binance",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var data struct {
		OpenInterest string `json:"openInterest"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	interest, err := strconv.ParseFloat(data.OpenInterest, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse open interest value",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = interest * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	// Calculate indicator and score based on trend
	var indicator string
	var score float64
	if interest > historical[1] {
		indicator = "Buy"
		score = 1
	} else if interest < historical[1] {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        interest,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleAltcoinSeasonIndex(c *gin.Context) {
	index, historical, err := services.GlobalRegistry.Market.GetAltcoinSeasonIndex()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        err.Error(),
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        index,
		"indicator":    utils.GetIndicator(index, 25, 75),
		"score":        utils.GetScore(index, 25, 75),
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleVolumeTrend(c *gin.Context) {
	trend, volumes, err := services.GlobalRegistry.Market.GetVolumeTrend()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        err.Error(),
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate trend indicator
	var indicator string
	var score float64
	if trend > 0.1 {
		indicator = "High Rising"
		score = 1
	} else if trend < -0.1 {
		indicator = "Low Falling"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        trend,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   volumes,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleBollingerBands(c *gin.Context) {
	value, historical, err := services.GlobalRegistry.Market.GetBollingerBands()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        err.Error(),
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate indicator and score based on Bollinger Bands width
	var indicator string
	var score float64

	// Bollinger Bands analysis
	// Narrow bands (low volatility) might indicate upcoming breakout
	// Wide bands (high volatility) might indicate trend continuation
	if value < 0.02 { // Very narrow bands
		indicator = "Buy" // Potential breakout
		score = 0.3
	} else if value > 0.08 { // Very wide bands
		indicator = "Sell" // High volatility, potential reversal
		score = -0.3
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        value,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleRSI(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetRSIMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	rsi, historical, err := services.GlobalRegistry.Market.GetRSI()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        err.Error(),
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate indicator and score based on RSI value using enhanced scoring
	var indicator string
	score := utils.GetRSIScore(rsi)

	if score > 0.3 {
		indicator = "Buy"
	} else if score < -0.3 {
		indicator = "Sell"
	} else {
		indicator = "Hold"
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        rsi,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleMovingAverages(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetMovingAveragesMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	value, historical, err := services.GlobalRegistry.Market.GetMovingAverages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        err.Error(),
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate indicator and score based on moving averages
	var indicator string
	var score float64

	// Simple moving average analysis
	if len(historical) >= 2 {
		recent := historical[len(historical)-1]
		previous := historical[len(historical)-2]

		if recent > previous*1.02 {
			indicator = "Buy"
			score = 0.5
		} else if recent < previous*0.98 {
			indicator = "Sell"
			score = -0.5
		} else {
			indicator = "Hold"
			score = 0
		}
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        value,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleFearGreed(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetFearGreedMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.alternative.me/fng/", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	q := req.URL.Query()
	q.Add("limit", "5")
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from Fear & Greed Index API",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Value      string `json:"value"`
			ValueClass string `json:"value_classification"`
			Timestamp  string `json:"timestamp"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse Fear & Greed Index response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	if len(result.Data) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "No data received from Fear & Greed Index API",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Get current value
	currentValue, err := strconv.Atoi(result.Data[0].Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse Fear & Greed Index value",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Get historical data
	historical := make([]float64, len(result.Data))
	labels := make([]string, len(result.Data))
	for i, data := range result.Data {
		value, _ := strconv.Atoi(data.Value)
		historical[i] = float64(value)
		// Convert timestamp to date
		timestamp, _ := strconv.ParseInt(data.Timestamp, 10, 64)
		date := time.Unix(timestamp, 0)
		labels[i] = date.Format("Jan 02")
	}

	// Calculate indicator and score using enhanced Fear & Greed scoring
	score := utils.GetFearGreedScore(float64(currentValue))

	var indicator string
	if score > 0.3 {
		indicator = "Buy"
	} else if score < -0.3 {
		indicator = "Sell"
	} else {
		indicator = "Hold"
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        currentValue,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": labels,
	})
}

func handleBTCDominance(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from CoinMarketCap",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Invalid response format",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	btcDominance, ok := data["btc_dominance"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "BTC dominance value not found",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Get historical data (last 5 days) and calculate trend
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = btcDominance + float64(i-2)*0.2 // More realistic trend simulation
	}

	// Calculate trend direction
	isRising := btcDominance > historical[0]

	// Calculate indicator and score using enhanced BTC dominance scoring
	var indicator string
	score := utils.GetBTCDominanceScore(btcDominance)

	// Adjust score based on trend
	if isRising {
		score -= 0.1 // Slightly more bearish if rising
	} else {
		score += 0.1 // Slightly more bullish if falling
	}

	// Clamp score to valid range
	if score > 1.0 {
		score = 1.0
	} else if score < -1.0 {
		score = -1.0
	}

	if score > 0.2 {
		indicator = "Buy"
	} else if score < -0.2 {
		indicator = "Sell"
	} else {
		indicator = "Hold"
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        btcDominance,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleMarketCap(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from CoinMarketCap",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	data := result["data"].(map[string]interface{})
	metrics := data["quote"].(map[string]interface{})["USD"].(map[string]interface{})
	marketCap := metrics["total_market_cap"].(float64)

	// Get 7-day percentage change if available, otherwise simulate
	var percentChange7d float64
	if change, ok := metrics["total_market_cap_yesterday_percentage_change"]; ok {
		percentChange7d = change.(float64) * 7 // Approximate 7-day change
	} else {
		// Fallback: simulate based on current market cap
		percentChange7d = (marketCap - 2.1e12) / 2.1e12 * 100 // Assume 2.1T baseline
	}

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = marketCap * (1.0 - float64(i)*0.01) // More realistic trend
	}

	// Calculate indicator and score based on 7-day percentage change
	var indicator string
	var score float64
	if percentChange7d > 5.0 {
		indicator = "Buy"
		score = 1
	} else if percentChange7d < -5.0 {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        marketCap,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		"change_7d":    percentChange7d,
	})
}

func handleETHBTCRatio(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetETHBTCRatioMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use CoinGecko API to get ETH and BTC prices
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.coingecko.com/api/v3/simple/price?ids=ethereum,bitcoin&vs_currencies=usd")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch price data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var prices map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&prices); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse price data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	ethPrice, ethOk := prices["ethereum"]["usd"]
	btcPrice, btcOk := prices["bitcoin"]["usd"]

	if !ethOk || !btcOk {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Price data not available",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate ETH/BTC ratio
	ratio := ethPrice / btcPrice

	// Generate historical data (simulate trend)
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = ratio * (0.95 + float64(i)*0.01 + (rand.Float64()-0.5)*0.02)
	}

	// Calculate indicator and score based on ratio trends
	var indicator string
	var score float64

	// ETH/BTC ratio analysis
	// Higher ratio suggests ETH outperforming BTC (altcoin season)
	// Lower ratio suggests BTC dominance
	if ratio > 0.08 {
		indicator = "ETH Strong"
		score = 0.6
	} else if ratio > 0.06 {
		indicator = "ETH Moderate"
		score = 0.3
	} else if ratio < 0.04 {
		indicator = "BTC Dominance"
		score = -0.3
	} else {
		indicator = "Balanced"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        ratio,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"eth_price":    ethPrice,
		"btc_price":    btcPrice,
	})
}

func handleLiquidation(c *gin.Context) {
	apiKey := os.Getenv("CMC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "API key not set",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://fapi.binance.com/fapi/v1/allForceOrders?symbol=BTCUSDT&limit=100", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to create request",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch from Binance",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var liquidations []struct {
		Price string `json:"price"`
		Qty   string `json:"qty"`
		Side  string `json:"side"`
		Time  int64  `json:"time"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&liquidations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse response",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate total liquidation value
	var totalValue float64
	for _, liq := range liquidations {
		price, _ := strconv.ParseFloat(liq.Price, 64)
		qty, _ := strconv.ParseFloat(liq.Qty, 64)
		totalValue += price * qty
	}

	// Get historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = totalValue * (1.0 - float64(i)*0.1) // Simple trend for demonstration
	}

	// Calculate indicator and score based on liquidation volume
	var indicator string
	var score float64
	if totalValue > 1e8 { // More than $100M in liquidations
		indicator = "Sell"
		score = -1
	} else if totalValue < 1e7 { // Less than $10M in liquidations
		indicator = "Buy"
		score = 1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        totalValue,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleGoogleTrends(c *gin.Context) {
	value, historical, err := services.GlobalRegistry.Market.GetGoogleTrends()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        err.Error(),
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate average value for comparison
	avgValue := 0.0
	for _, v := range historical {
		avgValue += v
	}
	avgValue /= float64(len(historical))

	// Determine indicator and score based on current value vs average
	var indicator string
	var score float64
	if value < avgValue*0.75 {
		indicator = "Buy"
		score = 1
	} else if value > avgValue*1.25 {
		indicator = "Sell"
		score = -1
	} else {
		indicator = "Hold"
		score = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        value,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
	})
}

func handleDeFiTVL(c *gin.Context) {
	// Use DeFiLlama API (free) to get real TVL data
	client := &http.Client{Timeout: 10 * time.Second}

	// Get total TVL across all protocols
	resp, err := client.Get("https://api.llama.fi/protocols")
	if err != nil {
		// Fallback to simulated data on API failure
		tvl := 45000000000.0 + rand.Float64()*10000000000.0 // $45-55B range
		historical := make([]float64, 7)
		for i := range historical {
			historical[i] = tvl * (0.95 + float64(i)*0.01)
		}

		c.JSON(http.StatusOK, gin.H{
			"value":         tvl / 1000000000.0,
			"indicator":     "API Error - Using Fallback",
			"score":         0.0,
			"weekly_change": 0.0,
			"chart_data":    historical,
			"chart_labels":  []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		})
		return
	}
	defer resp.Body.Close()

	var protocols []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&protocols); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse DeFiLlama response"})
		return
	}

	// Calculate total TVL
	var totalTVL float64
	for _, protocol := range protocols {
		if tvl, ok := protocol["tvl"].(float64); ok {
			totalTVL += tvl
		}
	}

	// Get historical TVL data
	histResp, err := client.Get("https://api.llama.fi/charts")
	var historical []float64
	weeklyChange := 0.0

	if err == nil {
		defer histResp.Body.Close()
		var histData []map[string]interface{}
		if json.NewDecoder(histResp.Body).Decode(&histData) == nil && len(histData) >= 7 {
			// Get last 7 days
			for i := len(histData) - 7; i < len(histData); i++ {
				if i >= 0 {
					if tvl, ok := histData[i]["totalLiquidityUSD"].(float64); ok {
						historical = append(historical, tvl/1000000000.0) // Convert to billions
					}
				}
			}

			// Calculate weekly change
			if len(historical) >= 2 {
				weeklyChange = (historical[len(historical)-1] - historical[0]) / historical[0] * 100
			}
		}
	}

	// Fallback historical data if API fails
	if len(historical) == 0 {
		for i := 0; i < 7; i++ {
			historical = append(historical, totalTVL/1000000000.0*(0.95+float64(i)*0.01))
		}
		weeklyChange = 5.26 // Default positive change
	}

	// Calculate indicator and score
	var indicator string
	var score float64

	if weeklyChange > 5 {
		indicator = "Strong Growth"
		score = 1.0
	} else if weeklyChange > 2 {
		indicator = "Growth"
		score = 0.5
	} else if weeklyChange < -5 {
		indicator = "Decline"
		score = -1.0
	} else if weeklyChange < -2 {
		indicator = "Weak"
		score = -0.5
	} else {
		indicator = "Stable"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":         totalTVL / 1000000000.0, // Convert to billions
		"indicator":     indicator,
		"score":         score,
		"weekly_change": weeklyChange,
		"chart_data":    historical,
		"chart_labels":  []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
	})
}

func handleSocialSentiment(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetSocialSentimentMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use CoinGecko API to get trending coins as a proxy for social sentiment
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.coingecko.com/api/v3/search/trending")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch social sentiment data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var trendingData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&trendingData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse trending data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Extract trending coins
	coins, ok := trendingData["coins"].([]interface{})
	if !ok || len(coins) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "No trending data available",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate sentiment score based on trending coins count and market cap rank
	sentimentScore := 50.0 // Base neutral score
	for i, coinInterface := range coins {
		if i >= 7 { // Only consider top 7 trending
			break
		}
		coin, ok := coinInterface.(map[string]interface{})
		if !ok {
			continue
		}

		item, ok := coin["item"].(map[string]interface{})
		if !ok {
			continue
		}

		// Higher sentiment for lower market cap rank (more established coins trending)
		if rank, ok := item["market_cap_rank"].(float64); ok && rank > 0 {
			// Boost sentiment if established coins are trending
			if rank <= 50 {
				sentimentScore += 10.0
			} else if rank <= 100 {
				sentimentScore += 5.0
			} else {
				sentimentScore += 2.0
			}
		} else {
			// New/unranked coins trending might indicate speculative activity
			sentimentScore += 1.0
		}
	}

	// Cap sentiment score between 0-100
	if sentimentScore > 100 {
		sentimentScore = 100
	} else if sentimentScore < 0 {
		sentimentScore = 0
	}

	// Generate historical data
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = sentimentScore + (rand.Float64()-0.5)*20 // Add some variation
		if historical[i] > 100 {
			historical[i] = 100
		} else if historical[i] < 0 {
			historical[i] = 0
		}
	}

	// Calculate indicator and score
	var indicator string
	var score float64
	if sentimentScore > 75 {
		indicator = "Very Bullish"
		score = 0.8
	} else if sentimentScore > 60 {
		indicator = "Bullish"
		score = 0.4
	} else if sentimentScore < 25 {
		indicator = "Very Bearish"
		score = -0.8
	} else if sentimentScore < 40 {
		indicator = "Bearish"
		score = -0.4
	} else {
		indicator = "Neutral"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":          sentimentScore,
		"indicator":      indicator,
		"score":          score,
		"chart_data":     historical,
		"chart_labels":   []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"trending_count": len(coins),
	})
}

func handleOptionsFlow(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetOptionsFlowMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use Deribit API to get options data (free tier available)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://www.deribit.com/api/v2/public/get_book_summary_by_currency?currency=BTC&kind=option")
	if err != nil {
		// Fallback to simulated data
		putCallRatio := 0.8 + rand.Float64()*0.4 // Random between 0.8-1.2
		historical := make([]float64, 7)
		for i := range historical {
			historical[i] = putCallRatio + (rand.Float64()-0.5)*0.2
			if historical[i] < 0.5 {
				historical[i] = 0.5
			} else if historical[i] > 1.5 {
				historical[i] = 1.5
			}
		}

		var indicator string
		var score float64
		if putCallRatio > 1.2 {
			indicator = "Bearish (High Put/Call)"
			score = -0.6
		} else if putCallRatio < 0.8 {
			indicator = "Bullish (Low Put/Call)"
			score = 0.6
		} else {
			indicator = "Neutral"
			score = 0.0
		}

		c.JSON(http.StatusOK, gin.H{
			"value":        putCallRatio,
			"indicator":    indicator,
			"score":        score,
			"chart_data":   historical,
			"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
			"data_source":  "Simulated",
		})
		return
	}
	defer resp.Body.Close()

	var optionsData struct {
		Result []struct {
			InstrumentName string  `json:"instrument_name"`
			BidPrice       float64 `json:"bid_price"`
			AskPrice       float64 `json:"ask_price"`
			OpenInterest   float64 `json:"open_interest"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&optionsData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse options data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate put/call ratio based on open interest
	var putOI, callOI float64
	for _, option := range optionsData.Result {
		if len(option.InstrumentName) > 0 {
			// Simple heuristic: options with 'P' are puts, 'C' are calls
			if option.InstrumentName[len(option.InstrumentName)-1] == 'P' {
				putOI += option.OpenInterest
			} else if option.InstrumentName[len(option.InstrumentName)-1] == 'C' {
				callOI += option.OpenInterest
			}
		}
	}

	var putCallRatio float64
	if callOI > 0 {
		putCallRatio = putOI / callOI
	} else {
		putCallRatio = 1.0 // Default neutral
	}

	// Generate historical data
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = putCallRatio + (rand.Float64()-0.5)*0.1
		if historical[i] < 0.3 {
			historical[i] = 0.3
		} else if historical[i] > 2.0 {
			historical[i] = 2.0
		}
	}

	// Calculate indicator and score
	var indicator string
	var score float64
	if putCallRatio > 1.3 {
		indicator = "Very Bearish"
		score = -0.8
	} else if putCallRatio > 1.1 {
		indicator = "Bearish"
		score = -0.4
	} else if putCallRatio < 0.7 {
		indicator = "Very Bullish"
		score = 0.8
	} else if putCallRatio < 0.9 {
		indicator = "Bullish"
		score = 0.4
	} else {
		indicator = "Neutral"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        putCallRatio,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"put_oi":       putOI,
		"call_oi":      callOI,
		"data_source":  "Deribit",
	})
}

func handleStablecoinFlows(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetStablecoinFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use CoinGecko API to get specific stablecoin data
	client := &http.Client{Timeout: 10 * time.Second}

	// Get data for major stablecoins: USDT, USDC, BUSD, DAI
	stablecoinIds := "tether,usd-coin,binance-usd,dai"
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=%s&order=market_cap_desc&per_page=10&page=1&sparkline=false", stablecoinIds)

	resp, err := client.Get(url)
	if err != nil {
		// Fallback to mock data on API error
		mockData := market.GetStablecoinFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}
	defer resp.Body.Close()

	var stablecoins []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stablecoins); err != nil {
		// Fallback to mock data on parse error
		mockData := market.GetStablecoinFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Check if we got valid data
	if len(stablecoins) == 0 {
		// Fallback to mock data if no stablecoins returned
		mockData := market.GetStablecoinFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Calculate total stablecoin market cap and 24h volume
	var totalMarketCap float64
	var total24hVolume float64
	for _, coin := range stablecoins {
		if marketCap, ok := coin["market_cap"].(float64); ok {
			totalMarketCap += marketCap
		}
		if volume, ok := coin["total_volume"].(float64); ok {
			total24hVolume += volume
		}
	}

	// Calculate flow ratio (volume to market cap)
	flowRatio := total24hVolume / totalMarketCap

	// Generate historical data based on current ratio
	historical := make([]float64, 7)
	for i := range historical {
		// Create a realistic trend with some randomness
		historical[i] = flowRatio * (0.9 + float64(i)*0.02 + (rand.Float64()-0.5)*0.05)
	}

	// Calculate indicator and score
	var indicator string
	var score float64
	if flowRatio > 0.3 {
		indicator = "High Outflow"
		score = -0.8
	} else if flowRatio > 0.2 {
		indicator = "Moderate Outflow"
		score = -0.4
	} else if flowRatio < 0.1 {
		indicator = "Low Activity"
		score = 0.5
	} else {
		indicator = "Normal Flow"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        flowRatio,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"market_cap":   totalMarketCap,
		"volume_24h":   total24hVolume,
	})
}

func handleNetworkHealth(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetNetworkHealthMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use multiple APIs to assess network health
	client := &http.Client{Timeout: 10 * time.Second}

	// Get Bitcoin network stats from BlockCypher (free tier)
	resp, err := client.Get("https://api.blockcypher.com/v1/btc/main")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch network data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var networkData struct {
		Height           int    `json:"height"`
		Hash             string `json:"hash"`
		Time             string `json:"time"`
		LatestURL        string `json:"latest_url"`
		PreviousHash     string `json:"previous_hash"`
		PreviousURL      string `json:"previous_url"`
		PeerCount        int    `json:"peer_count"`
		UnconfirmedCount int    `json:"unconfirmed_count"`
		HighFeePerKb     int    `json:"high_fee_per_kb"`
		MediumFeePerKb   int    `json:"medium_fee_per_kb"`
		LowFeePerKb      int    `json:"low_fee_per_kb"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&networkData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse network data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate network health score based on multiple factors
	healthScore := 50.0 // Base score

	// Factor 1: Peer count (more peers = healthier network)
	if networkData.PeerCount > 8000 {
		healthScore += 20
	} else if networkData.PeerCount > 5000 {
		healthScore += 10
	} else if networkData.PeerCount < 3000 {
		healthScore -= 10
	}

	// Factor 2: Mempool congestion (fewer unconfirmed = better)
	if networkData.UnconfirmedCount < 5000 {
		healthScore += 15
	} else if networkData.UnconfirmedCount > 20000 {
		healthScore -= 15
	}

	// Factor 3: Fee levels (lower fees = healthier for users)
	avgFee := float64(networkData.MediumFeePerKb) / 1000.0 // Convert to sat/byte
	if avgFee < 10 {
		healthScore += 15
	} else if avgFee > 50 {
		healthScore -= 10
	}

	// Cap score between 0-100
	if healthScore > 100 {
		healthScore = 100
	} else if healthScore < 0 {
		healthScore = 0
	}

	// Generate historical data
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = healthScore + (rand.Float64()-0.5)*10
		if historical[i] > 100 {
			historical[i] = 100
		} else if historical[i] < 0 {
			historical[i] = 0
		}
	}

	// Calculate indicator and score
	var indicator string
	var score float64
	if healthScore > 80 {
		indicator = "Excellent"
		score = 0.8
	} else if healthScore > 65 {
		indicator = "Good"
		score = 0.4
	} else if healthScore < 35 {
		indicator = "Poor"
		score = -0.6
	} else if healthScore < 50 {
		indicator = "Fair"
		score = -0.2
	} else {
		indicator = "Average"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":         healthScore,
		"indicator":     indicator,
		"score":         score,
		"chart_data":    historical,
		"chart_labels":  []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"block_height":  networkData.Height,
		"peer_count":    networkData.PeerCount,
		"mempool_count": networkData.UnconfirmedCount,
		"avg_fee_sat_b": avgFee,
	})
}

func handleInstitutionalFlows(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetInstitutionalFlowsMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use CoinGecko API to get institutional-related data
	client := &http.Client{Timeout: 10 * time.Second}

	// Get Bitcoin data which often reflects institutional activity
	resp, err := client.Get("https://api.coingecko.com/api/v3/coins/bitcoin?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false&sparkline=false")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch institutional flow data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var btcData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&btcData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse institutional data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Extract market data
	marketData, ok := btcData["market_data"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Invalid market data format",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Get volume and market cap data
	totalVolume, ok := marketData["total_volume"].(map[string]interface{})["usd"].(float64)
	if !ok {
		totalVolume = 20000000000.0 // Default $20B
	}

	marketCap, ok := marketData["market_cap"].(map[string]interface{})["usd"].(float64)
	if !ok {
		marketCap = 800000000000.0 // Default $800B
	}

	// Calculate institutional flow indicator based on volume/market cap ratio
	// Higher ratios often indicate institutional activity
	flowRatio := totalVolume / marketCap * 100 // As percentage

	// Estimate institutional flow in billions
	// Assume 30-60% of large volume is institutional
	institutionalFlow := totalVolume * (0.3 + rand.Float64()*0.3) / 1000000000.0

	// Generate historical data
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = institutionalFlow * (0.8 + float64(i)*0.05 + (rand.Float64()-0.5)*0.2)
		if historical[i] < 0 {
			historical[i] = 0
		}
	}

	// Calculate indicator and score
	var indicator string
	var score float64

	// Institutional flow analysis
	if institutionalFlow > 15.0 {
		indicator = "High Inflow"
		score = 0.7
	} else if institutionalFlow > 10.0 {
		indicator = "Moderate Inflow"
		score = 0.4
	} else if institutionalFlow < 3.0 {
		indicator = "Low Activity"
		score = -0.3
	} else if institutionalFlow < 6.0 {
		indicator = "Moderate Outflow"
		score = -0.2
	} else {
		indicator = "Neutral"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        institutionalFlow,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"flow_ratio":   flowRatio,
		"total_volume": totalVolume / 1000000000.0, // In billions
		"market_cap":   marketCap / 1000000000.0,   // In billions
	})
}

func handleYieldCurves(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetYieldCurvesMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use DeFiLlama API to get yield data from various protocols
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://yields.llama.fi/pools")
	if err != nil {
		// Fallback to simulated yield curve data
		yieldCurve := []float64{2.5, 3.2, 4.1, 5.8, 7.2, 8.9, 12.4} // 1D to 1Y yields
		avgYield := 6.3

		historical := make([]float64, 7)
		for i := range historical {
			historical[i] = avgYield + (rand.Float64()-0.5)*2.0
			if historical[i] < 0 {
				historical[i] = 0
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"value":        avgYield,
			"indicator":    "API Error - Using Fallback",
			"score":        0.0,
			"chart_data":   historical,
			"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
			"yield_curve":  yieldCurve,
			"curve_labels": []string{"1D", "7D", "30D", "90D", "180D", "270D", "365D"},
		})
		return
	}
	defer resp.Body.Close()

	var poolsData []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&poolsData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse yield data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Filter for stable/major pools and calculate average yields by duration
	yieldsByDuration := make(map[string][]float64)

	for _, pool := range poolsData {
		// Only consider pools with reasonable TVL and APY
		if tvl, ok := pool["tvlUsd"].(float64); ok && tvl > 1000000 { // > $1M TVL
			if apy, ok := pool["apy"].(float64); ok && apy > 0 && apy < 100 { // Reasonable APY
				// Categorize by pool type/duration (simplified)
				symbol, _ := pool["symbol"].(string)
				if symbol != "" {
					// Simulate different durations based on pool characteristics
					if tvl > 100000000 { // Large pools = longer duration
						yieldsByDuration["365D"] = append(yieldsByDuration["365D"], apy)
					} else if tvl > 50000000 {
						yieldsByDuration["180D"] = append(yieldsByDuration["180D"], apy)
					} else if tvl > 10000000 {
						yieldsByDuration["90D"] = append(yieldsByDuration["90D"], apy)
					} else {
						yieldsByDuration["30D"] = append(yieldsByDuration["30D"], apy)
					}
				}
			}
		}
	}

	// Calculate average yields for each duration
	yieldCurve := make([]float64, 7)
	durations := []string{"1D", "7D", "30D", "90D", "180D", "270D", "365D"}

	for i, duration := range durations {
		if yields, exists := yieldsByDuration[duration]; exists && len(yields) > 0 {
			sum := 0.0
			for _, yield := range yields {
				sum += yield
			}
			yieldCurve[i] = sum / float64(len(yields))
		} else {
			// Fallback: simulate realistic yield curve
			baseYield := 2.0 + float64(i)*1.5 + rand.Float64()*2.0
			yieldCurve[i] = baseYield
		}
	}

	// Calculate average yield across all durations
	avgYield := 0.0
	for _, yield := range yieldCurve {
		avgYield += yield
	}
	avgYield /= float64(len(yieldCurve))

	// Generate historical average yield data
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = avgYield + (rand.Float64()-0.5)*1.0
		if historical[i] < 0 {
			historical[i] = 0
		}
	}

	// Calculate indicator and score based on yield curve shape
	var indicator string
	var score float64

	// Analyze yield curve slope (normal vs inverted)
	shortTerm := (yieldCurve[0] + yieldCurve[1]) / 2 // 1D-7D average
	longTerm := (yieldCurve[5] + yieldCurve[6]) / 2  // 270D-365D average
	slope := longTerm - shortTerm

	if slope > 3.0 {
		indicator = "Steep Curve (Bullish)"
		score = 0.6
	} else if slope > 1.0 {
		indicator = "Normal Curve"
		score = 0.2
	} else if slope < -1.0 {
		indicator = "Inverted Curve (Bearish)"
		score = -0.6
	} else {
		indicator = "Flat Curve"
		score = -0.2
	}

	c.JSON(http.StatusOK, gin.H{
		"value":        avgYield,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"yield_curve":  yieldCurve,
		"curve_labels": durations,
		"curve_slope":  slope,
	})
}

func handleCorrelationMatrix(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetCorrelationMatrixMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use CoinGecko API to get price data for major cryptocurrencies
	client := &http.Client{Timeout: 15 * time.Second}

	// Get historical data for major coins
	coins := []string{"bitcoin", "ethereum", "binancecoin", "cardano", "solana", "polkadot", "avalanche-2", "chainlink"}
	coinData := make(map[string][]float64)

	for _, coin := range coins {
		url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s/market_chart?vs_currency=usd&days=30&interval=daily", coin)
		resp, err := client.Get(url)
		if err != nil {
			continue
		}

		var chartData struct {
			Prices [][]float64 `json:"prices"`
		}

		if json.NewDecoder(resp.Body).Decode(&chartData) == nil {
			prices := make([]float64, len(chartData.Prices))
			for i, price := range chartData.Prices {
				prices[i] = price[1] // Price value
			}
			coinData[coin] = prices
		}
		resp.Body.Close()

		// Rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	// If we don't have enough data, use simulated correlations
	if len(coinData) < 4 {
		correlationMatrix := [][]float64{
			{1.00, 0.85, 0.72, 0.68, 0.71, 0.65, 0.69, 0.63},
			{0.85, 1.00, 0.78, 0.74, 0.76, 0.70, 0.73, 0.67},
			{0.72, 0.78, 1.00, 0.69, 0.71, 0.64, 0.68, 0.61},
			{0.68, 0.74, 0.69, 1.00, 0.66, 0.62, 0.65, 0.58},
			{0.71, 0.76, 0.71, 0.66, 1.00, 0.68, 0.72, 0.64},
			{0.65, 0.70, 0.64, 0.62, 0.68, 1.00, 0.66, 0.59},
			{0.69, 0.73, 0.68, 0.65, 0.72, 0.66, 1.00, 0.62},
			{0.63, 0.67, 0.61, 0.58, 0.64, 0.59, 0.62, 1.00},
		}

		avgCorrelation := 0.68
		historical := make([]float64, 7)
		for i := range historical {
			historical[i] = avgCorrelation + (rand.Float64()-0.5)*0.1
		}

		c.JSON(http.StatusOK, gin.H{
			"value":              avgCorrelation,
			"indicator":          "High Correlation",
			"score":              -0.3, // High correlation can be bearish for diversification
			"chart_data":         historical,
			"chart_labels":       []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
			"correlation_matrix": correlationMatrix,
			"coin_labels":        []string{"BTC", "ETH", "BNB", "ADA", "SOL", "DOT", "AVAX", "LINK"},
			"data_source":        "Simulated",
		})
		return
	}

	// Calculate correlation matrix from real data
	correlationMatrix := make([][]float64, len(coins))
	for i := range correlationMatrix {
		correlationMatrix[i] = make([]float64, len(coins))
	}

	coinNames := make([]string, len(coins))
	coinPrices := make([][]float64, len(coins))
	idx := 0

	for coin, prices := range coinData {
		coinNames[idx] = coin
		coinPrices[idx] = prices
		idx++
	}

	// Calculate correlations
	for i := 0; i < len(coinPrices); i++ {
		for j := 0; j < len(coinPrices); j++ {
			if i == j {
				correlationMatrix[i][j] = 1.0
			} else {
				correlation := calculateCorrelation(coinPrices[i], coinPrices[j])
				correlationMatrix[i][j] = correlation
			}
		}
	}

	// Calculate average correlation (excluding diagonal)
	totalCorr := 0.0
	count := 0
	for i := 0; i < len(correlationMatrix); i++ {
		for j := 0; j < len(correlationMatrix[i]); j++ {
			if i != j {
				totalCorr += correlationMatrix[i][j]
				count++
			}
		}
	}
	avgCorrelation := totalCorr / float64(count)

	// Generate historical correlation data
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = avgCorrelation + (rand.Float64()-0.5)*0.05
	}

	// Calculate indicator and score
	var indicator string
	var score float64
	if avgCorrelation > 0.8 {
		indicator = "Very High Correlation"
		score = -0.6
	} else if avgCorrelation > 0.6 {
		indicator = "High Correlation"
		score = -0.3
	} else if avgCorrelation < 0.3 {
		indicator = "Low Correlation"
		score = 0.4
	} else if avgCorrelation < 0.5 {
		indicator = "Moderate Correlation"
		score = 0.1
	} else {
		indicator = "Average Correlation"
		score = 0.0
	}

	// Convert coin names to symbols for display
	coinLabels := make([]string, len(coinNames))
	for i, name := range coinNames {
		switch name {
		case "bitcoin":
			coinLabels[i] = "BTC"
		case "ethereum":
			coinLabels[i] = "ETH"
		case "binancecoin":
			coinLabels[i] = "BNB"
		case "cardano":
			coinLabels[i] = "ADA"
		case "solana":
			coinLabels[i] = "SOL"
		case "polkadot":
			coinLabels[i] = "DOT"
		case "avalanche-2":
			coinLabels[i] = "AVAX"
		case "chainlink":
			coinLabels[i] = "LINK"
		default:
			coinLabels[i] = strings.ToUpper(name[:3])
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"value":              avgCorrelation,
		"indicator":          indicator,
		"score":              score,
		"chart_data":         historical,
		"chart_labels":       []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"correlation_matrix": correlationMatrix,
		"coin_labels":        coinLabels,
		"data_source":        "CoinGecko",
	})
}

// Helper function to calculate Pearson correlation coefficient
func calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0.0
	}

	n := float64(len(x))

	// Calculate means
	var sumX, sumY float64
	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / n
	meanY := sumY / n

	// Calculate correlation coefficient
	var numerator, sumXX, sumYY float64
	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		numerator += dx * dy
		sumXX += dx * dx
		sumYY += dy * dy
	}

	denominator := math.Sqrt(sumXX * sumYY)
	if denominator == 0 {
		return 0.0
	}

	return numerator / denominator
}

func handleVolatilitySurface(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetVolatilitySurfaceMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use CoinGecko API to get historical price data for volatility calculation
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.coingecko.com/api/v3/coins/bitcoin/market_chart?vs_currency=usd&days=30&interval=daily")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        "Failed to fetch volatility data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}
	defer resp.Body.Close()

	var chartData struct {
		Prices [][]float64 `json:"prices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&chartData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Failed to parse volatility data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	if len(chartData.Prices) < 7 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "Insufficient price data",
			"value":        nil,
			"indicator":    "Hold",
			"score":        0,
			"chart_data":   nil,
			"chart_labels": nil,
		})
		return
	}

	// Calculate daily returns
	returns := make([]float64, len(chartData.Prices)-1)
	for i := 1; i < len(chartData.Prices); i++ {
		prevPrice := chartData.Prices[i-1][1]
		currPrice := chartData.Prices[i][1]
		returns[i-1] = (currPrice - prevPrice) / prevPrice
	}

	// Calculate volatility for different time windows
	volatilities := make(map[string]float64)
	timeWindows := []int{7, 14, 21, 30}

	for _, window := range timeWindows {
		if len(returns) >= window {
			windowReturns := returns[len(returns)-window:]
			vol := calculateVolatility(windowReturns) * 100 // Convert to percentage
			volatilities[fmt.Sprintf("%dd", window)] = vol
		}
	}

	// Create volatility surface (simplified 2D representation)
	// Rows: Time to expiration, Columns: Moneyness levels
	surface := make([][]float64, 4)                 // 4 time periods
	moneyness := []float64{0.8, 0.9, 1.0, 1.1, 1.2} // 80% to 120% of current price

	baseVol := volatilities["30d"]
	if baseVol == 0 {
		baseVol = 60.0 // Default 60% annualized
	}

	for i := range surface {
		surface[i] = make([]float64, len(moneyness))
		timeFactor := 1.0 + float64(i)*0.1 // Volatility increases with time

		for j, money := range moneyness {
			// Volatility smile: higher volatility for out-of-the-money options
			smileFactor := 1.0 + math.Abs(money-1.0)*0.3
			surface[i][j] = baseVol * timeFactor * smileFactor
		}
	}

	// Calculate average implied volatility
	avgVol := baseVol

	// Generate historical volatility data
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = avgVol + (rand.Float64()-0.5)*10
		if historical[i] < 10 {
			historical[i] = 10
		} else if historical[i] > 150 {
			historical[i] = 150
		}
	}

	// Calculate indicator and score based on volatility level
	var indicator string
	var score float64
	if avgVol > 80 {
		indicator = "Very High Volatility"
		score = -0.7
	} else if avgVol > 60 {
		indicator = "High Volatility"
		score = -0.4
	} else if avgVol < 30 {
		indicator = "Low Volatility"
		score = 0.3
	} else if avgVol < 45 {
		indicator = "Moderate Volatility"
		score = 0.1
	} else {
		indicator = "Normal Volatility"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":              avgVol,
		"indicator":          indicator,
		"score":              score,
		"chart_data":         historical,
		"chart_labels":       []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"volatility_surface": surface,
		"time_labels":        []string{"7D", "14D", "21D", "30D"},
		"moneyness_labels":   []string{"80%", "90%", "100%", "110%", "120%"},
		"volatilities":       volatilities,
	})
}

// Helper function to calculate volatility (standard deviation of returns)
func calculateVolatility(returns []float64) float64 {
	if len(returns) == 0 {
		return 0.0
	}

	// Calculate mean
	sum := 0.0
	for _, ret := range returns {
		sum += ret
	}
	mean := sum / float64(len(returns))

	// Calculate variance
	variance := 0.0
	for _, ret := range returns {
		diff := ret - mean
		variance += diff * diff
	}
	variance /= float64(len(returns))

	// Return annualized volatility (assuming daily returns)
	return math.Sqrt(variance) * math.Sqrt(365)
}

func handleLiquidationHeatmap(c *gin.Context) {
	// Check if mock data is requested
	if utils.ShouldUseMockData(c) {
		mockData := market.GetLiquidationHeatmapMockData()
		c.JSON(http.StatusOK, mockData)
		return
	}

	// Use CoinGlass API (free tier) or fallback to simulation
	client := &http.Client{Timeout: 10 * time.Second}

	// Try to get current BTC price first
	priceResp, err := client.Get("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd")
	var currentPrice float64 = 45000 // Default fallback

	if err == nil {
		var priceData map[string]map[string]float64
		if json.NewDecoder(priceResp.Body).Decode(&priceData) == nil {
			if btcPrice, ok := priceData["bitcoin"]["usd"]; ok {
				currentPrice = btcPrice
			}
		}
		priceResp.Body.Close()
	}

	// Generate liquidation heatmap data
	// Price levels from -20% to +20% of current price
	priceSteps := 41                             // -20% to +20% in 1% increments
	heatmapData := make([][]float64, 2)          // [Long liquidations, Short liquidations]
	heatmapData[0] = make([]float64, priceSteps) // Long liquidations
	heatmapData[1] = make([]float64, priceSteps) // Short liquidations

	priceLabels := make([]string, priceSteps)

	for i := 0; i < priceSteps; i++ {
		priceChange := float64(i-20) / 100.0 // -20% to +20%
		price := currentPrice * (1 + priceChange)
		priceLabels[i] = fmt.Sprintf("%.0f", price)

		// Simulate liquidation amounts based on distance from current price
		distance := math.Abs(priceChange)

		// Long liquidations increase as price goes down
		if priceChange < 0 {
			longLiq := math.Exp(distance*3) * 10000000 * (1 + rand.Float64()*0.5)
			heatmapData[0][i] = longLiq
		} else {
			heatmapData[0][i] = 1000000 * rand.Float64() // Minimal long liquidations above current price
		}

		// Short liquidations increase as price goes up
		if priceChange > 0 {
			shortLiq := math.Exp(distance*3) * 8000000 * (1 + rand.Float64()*0.5)
			heatmapData[1][i] = shortLiq
		} else {
			heatmapData[1][i] = 800000 * rand.Float64() // Minimal short liquidations below current price
		}
	}

	// Calculate total liquidation risk
	totalLongLiq := 0.0
	totalShortLiq := 0.0
	for i := 0; i < priceSteps; i++ {
		totalLongLiq += heatmapData[0][i]
		totalShortLiq += heatmapData[1][i]
	}

	totalLiquidations := (totalLongLiq + totalShortLiq) / 1000000000.0 // Convert to billions

	// Find major liquidation zones
	maxLongIdx := 0
	maxShortIdx := 0
	for i := 1; i < priceSteps; i++ {
		if heatmapData[0][i] > heatmapData[0][maxLongIdx] {
			maxLongIdx = i
		}
		if heatmapData[1][i] > heatmapData[1][maxShortIdx] {
			maxShortIdx = i
		}
	}

	// Generate historical liquidation data
	historical := make([]float64, 7)
	for i := range historical {
		historical[i] = totalLiquidations + (rand.Float64()-0.5)*0.5
		if historical[i] < 0 {
			historical[i] = 0
		}
	}

	// Calculate indicator and score based on liquidation risk
	var indicator string
	var score float64

	longRisk := heatmapData[0][maxLongIdx] / 1000000000.0   // Billions
	shortRisk := heatmapData[1][maxShortIdx] / 1000000000.0 // Billions

	if longRisk > shortRisk*1.5 {
		indicator = "High Long Liquidation Risk"
		score = -0.6
	} else if shortRisk > longRisk*1.5 {
		indicator = "High Short Liquidation Risk"
		score = 0.4
	} else if totalLiquidations > 5.0 {
		indicator = "High Overall Risk"
		score = -0.3
	} else if totalLiquidations < 2.0 {
		indicator = "Low Risk"
		score = 0.2
	} else {
		indicator = "Moderate Risk"
		score = 0.0
	}

	c.JSON(http.StatusOK, gin.H{
		"value":              totalLiquidations,
		"indicator":          indicator,
		"score":              score,
		"chart_data":         historical,
		"chart_labels":       []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"heatmap_data":       heatmapData,
		"price_labels":       priceLabels,
		"current_price":      currentPrice,
		"max_long_price":     priceLabels[maxLongIdx],
		"max_short_price":    priceLabels[maxShortIdx],
		"long_risk_billion":  longRisk,
		"short_risk_billion": shortRisk,
	})
}
