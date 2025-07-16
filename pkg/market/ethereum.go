package market

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	COINGECKO_BASE_URL = "https://api.coingecko.com/api/v3"
	ETHERSCAN_BASE_URL = "https://api.etherscan.io/api"
	DEXTOOLS_BASE_URL  = "https://www.dextools.io/shared/data/pair"
)

type EthereumService struct {
	client  *http.Client
	headers map[string]string
}

type CoinGeckoTokenData struct {
	ID         string `json:"id"`
	Symbol     string `json:"symbol"`
	Name       string `json:"name"`
	MarketData struct {
		CurrentPrice struct {
			USD float64 `json:"usd"`
		} `json:"current_price"`
		PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
		PriceChangePercentage7d  float64 `json:"price_change_percentage_7d"`
		PriceChangePercentage30d float64 `json:"price_change_percentage_30d"`
		High24h                  struct {
			USD float64 `json:"usd"`
		} `json:"high_24h"`
		Low24h struct {
			USD float64 `json:"usd"`
		} `json:"low_24h"`
		TotalVolume struct {
			USD float64 `json:"usd"`
		} `json:"total_volume"`
		MarketCap struct {
			USD float64 `json:"usd"`
		} `json:"market_cap"`
	} `json:"market_data"`
}

type DexToolsResponse struct {
	Data struct {
		Pair struct {
			Price       float64 `json:"price"`
			PriceChange struct {
				H1  float64 `json:"h1"`
				H6  float64 `json:"h6"`
				H24 float64 `json:"h24"`
			} `json:"priceChange"`
			Volume struct {
				H24 float64 `json:"h24"`
			} `json:"volume"`
		} `json:"pair"`
	} `json:"data"`
}

type EthereumAnalysisResult struct {
	EntryPrice  float64 `json:"entry_price"`
	MaxPrice    float64 `json:"max_price"`
	ROIPercent  float64 `json:"roi_percent"`
	MaxX        float64 `json:"max_x"`
	Error       string  `json:"error,omitempty"`
	IsUseful    bool    `json:"is_useful"`
	OptimalExit string  `json:"optimal_exit"`
}

// No hardcoded token performance data - all data comes from real APIs

func NewEthereumService() *EthereumService {
	return &EthereumService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: map[string]string{
			"Accept": "application/json",
		},
	}
}

func (e *EthereumService) AnalyzeTokenPerformance(tokenAddress string, callTimeISO string) (*EthereumAnalysisResult, error) {
	// Try to get real data from CoinGecko
	result, err := e.getCoinGeckoData(tokenAddress, callTimeISO)
	if err == nil && result.Error == "" {
		return result, nil
	}

	// Try DexTools as fallback
	result, err = e.getDexToolsData(tokenAddress, callTimeISO)
	if err == nil && result.Error == "" {
		return result, nil
	}

	// If all real APIs fail, return error instead of fallback data
	return &EthereumAnalysisResult{
		Error: "Unable to fetch real data for Ethereum token",
	}, fmt.Errorf("all Ethereum token APIs failed for %s", tokenAddress)
}

func (e *EthereumService) getCoinGeckoData(tokenAddress string, callTimeISO string) (*EthereumAnalysisResult, error) {
	// First, try to find the token by contract address
	url := fmt.Sprintf("%s/coins/ethereum/contract/%s", COINGECKO_BASE_URL, tokenAddress)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	for key, value := range e.headers {
		req.Header.Set(key, value)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from CoinGecko: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("CoinGecko API returned status %d", resp.StatusCode)
	}

	var tokenData CoinGeckoTokenData
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		return nil, fmt.Errorf("failed to decode CoinGecko response: %v", err)
	}

	// Calculate performance based on available data
	currentPrice := tokenData.MarketData.CurrentPrice.USD
	high24h := tokenData.MarketData.High24h.USD
	low24h := tokenData.MarketData.Low24h.USD
	priceChange24h := tokenData.MarketData.PriceChangePercentage24h
	priceChange7d := tokenData.MarketData.PriceChangePercentage7d

	// Use the highest available price change as max performance
	maxPerformance := priceChange24h
	if priceChange7d > maxPerformance {
		maxPerformance = priceChange7d
	}

	// Calculate entry price based on signal time
	signalTime, err := time.Parse(time.RFC3339, callTimeISO)
	if err != nil {
		signalTime = time.Now().Add(-24 * time.Hour)
	}

	hoursElapsed := time.Since(signalTime).Hours()
	var entryPrice float64

	if hoursElapsed <= 24 {
		// Use 24h low as entry price approximation
		entryPrice = low24h
	} else {
		// Estimate entry price based on current price and time elapsed
		entryPrice = currentPrice / (1 + maxPerformance/100)
	}

	maxPrice := high24h
	if maxPrice < currentPrice {
		maxPrice = currentPrice
	}

	roi := (maxPrice/entryPrice - 1) * 100
	xGrowth := maxPrice / entryPrice

	isUseful := roi > 15.0
	optimalExit := e.calculateOptimalExit(hoursElapsed, roi)

	log.Printf("CoinGecko data for %s: Entry=%.8f, Max=%.8f, ROI=%.2f%%", tokenAddress, entryPrice, maxPrice, roi)

	return &EthereumAnalysisResult{
		EntryPrice:  entryPrice,
		MaxPrice:    maxPrice,
		ROIPercent:  roi,
		MaxX:        xGrowth,
		IsUseful:    isUseful,
		OptimalExit: optimalExit,
	}, nil
}

func (e *EthereumService) getDexToolsData(tokenAddress string, callTimeISO string) (*EthereumAnalysisResult, error) {
	// DexTools API endpoint (note: this is a simplified example)
	url := fmt.Sprintf("https://api.dextools.io/v1/token?chain=ether&address=%s", tokenAddress)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create DexTools request: %v", err)
	}

	for key, value := range e.headers {
		req.Header.Set(key, value)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from DexTools: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("DexTools API returned status %d", resp.StatusCode)
	}

	var dexData DexToolsResponse
	if err := json.NewDecoder(resp.Body).Decode(&dexData); err != nil {
		return nil, fmt.Errorf("failed to decode DexTools response: %v", err)
	}

	// Calculate performance from DexTools data
	currentPrice := dexData.Data.Pair.Price
	priceChange24h := dexData.Data.Pair.PriceChange.H24

	entryPrice := currentPrice / (1 + priceChange24h/100)
	maxPrice := currentPrice
	if priceChange24h > 0 {
		maxPrice = currentPrice
	} else {
		maxPrice = entryPrice
	}

	roi := (maxPrice/entryPrice - 1) * 100
	xGrowth := maxPrice / entryPrice

	isUseful := roi > 15.0
	optimalExit := "2h 30m after signal" // Default

	log.Printf("DexTools data for %s: Entry=%.8f, Max=%.8f, ROI=%.2f%%", tokenAddress, entryPrice, maxPrice, roi)

	return &EthereumAnalysisResult{
		EntryPrice:  entryPrice,
		MaxPrice:    maxPrice,
		ROIPercent:  roi,
		MaxX:        xGrowth,
		IsUseful:    isUseful,
		OptimalExit: optimalExit,
	}, nil
}

func (e *EthereumService) calculateOptimalExit(hoursElapsed, roi float64) string {
	if roi > 300 {
		return "2h 30m after signal"
	} else if roi > 200 {
		return "3h 15m after signal"
	} else if roi > 100 {
		return "4h 45m after signal"
	} else if roi > 50 {
		return "6h 30m after signal"
	} else {
		return "12h after signal"
	}
}

// GetTokenPerformance is a wrapper that matches the interface expected by the main application
func (e *EthereumService) GetTokenPerformance(contractAddr, signalTimeStr string) (interface{}, interface{}, string, bool) {
	result, err := e.AnalyzeTokenPerformance(contractAddr, signalTimeStr)
	if err != nil {
		log.Printf("Ethereum analysis failed for %s: %v", contractAddr, err)
		return "N/A", "N/A", "N/A", false
	}

	if result.Error != "" {
		log.Printf("Ethereum analysis error for %s: %s", contractAddr, result.Error)
		return "N/A", "N/A", "N/A", false
	}

	// Return performance as percentage string for consistency
	perfStr := fmt.Sprintf("%.2f%%", result.ROIPercent)
	maxPerfStr := fmt.Sprintf("%.2f%%", result.ROIPercent)

	log.Printf("Ethereum performance for %s: %s (Max: %s)", contractAddr, perfStr, maxPerfStr)

	return perfStr, maxPerfStr, result.OptimalExit, result.IsUseful
}
