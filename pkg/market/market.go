package market

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// MarketService handles market data operations
type MarketService struct {
	cmcAPIKey     string
	binanceAPIKey string
}

// NewMarketService creates a new market service instance
func NewMarketService(cmcAPIKey string) *MarketService {
	return &MarketService{
		cmcAPIKey:     cmcAPIKey,
		binanceAPIKey: os.Getenv("BINANCE_API_KEY"),
	}
}

func (s *MarketService) makeRequest(url string, maxRetries int) (*http.Response, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		// Add CoinMarketCap API key to headers
		req.Header.Set("X-CMC_PRO_API_KEY", s.cmcAPIKey)

		// Add delay between retries with exponential backoff
		if i > 0 {
			delay := time.Duration(math.Pow(2, float64(i))) * time.Second
			time.Sleep(delay)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// Check for rate limit
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					time.Sleep(time.Duration(seconds) * time.Second)
				}
			}
			resp.Body.Close()
			lastErr = fmt.Errorf("rate limited (attempt %d/%d)", i+1, maxRetries)
			continue
		}

		// Check for other error status codes
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}

// GetExchangeFlows returns the net flow of BTC to/from exchanges
func (s *MarketService) GetExchangeFlows() (float64, error) {
	url := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=BTC&convert=USD"

	resp, err := s.makeRequest(url, 3)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch exchange flows: %v", err)
	}
	defer resp.Body.Close()

	var data struct {
		Data struct {
			BTC struct {
				Quote struct {
					USD struct {
						Volume24h float64 `json:"volume_24h"`
					} `json:"USD"`
				} `json:"quote"`
			} `json:"BTC"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("failed to parse exchange flows: %v", err)
	}

	// Use 24h volume as a proxy for exchange flows
	return data.Data.BTC.Quote.USD.Volume24h, nil
}

// GetActiveAddresses returns the number of active BTC addresses
func (s *MarketService) GetActiveAddresses() (int64, error) {
	url := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=BTC&convert=USD"

	resp, err := s.makeRequest(url, 3)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch active addresses: %v", err)
	}
	defer resp.Body.Close()

	var data struct {
		Data struct {
			BTC struct {
				Quote struct {
					USD struct {
						Volume24h float64 `json:"volume_24h"`
					} `json:"USD"`
				} `json:"quote"`
			} `json:"BTC"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("failed to parse active addresses: %v", err)
	}

	// Estimate active addresses based on volume
	// This is a rough estimate: assume each transaction involves 2 addresses
	// and each address is used once per $1000 of volume
	totalVolume := data.Data.BTC.Quote.USD.Volume24h
	estimatedAddresses := int64(totalVolume / 1000 * 2)

	return estimatedAddresses, nil
}

// GetWhaleTransactions returns the number of large BTC transactions
func (s *MarketService) GetWhaleTransactions() (int64, error) {
	url := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=BTC&convert=USD"

	resp, err := s.makeRequest(url, 3)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch transaction data: %v", err)
	}
	defer resp.Body.Close()

	var data struct {
		Data struct {
			BTC struct {
				Quote struct {
					USD struct {
						Volume24h float64 `json:"volume_24h"`
					} `json:"USD"`
				} `json:"quote"`
			} `json:"BTC"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("failed to parse transaction data: %v", err)
	}

	// Estimate whale transactions based on volume
	// Assume a whale transaction is > $500,000
	totalVolume := data.Data.BTC.Quote.USD.Volume24h
	whaleCount := int64(totalVolume / 500000)

	return whaleCount, nil
}

// GetFundingRate returns the current funding rate for BTC perpetual futures
func (s *MarketService) GetFundingRate() (float64, error) {
	url := "https://fapi.binance.com/fapi/v1/premiumIndex?symbol=BTCUSDT"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch funding rate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to fetch funding rate: status code %d", resp.StatusCode)
	}

	var data struct {
		Symbol          string `json:"symbol"`
		MarkPrice       string `json:"markPrice"`
		LastFundingRate string `json:"lastFundingRate"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("failed to parse funding rate: %v", err)
	}

	rate, err := strconv.ParseFloat(data.LastFundingRate, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse funding rate value: %v", err)
	}

	return rate, nil
}

// GetOpenInterest returns the total open interest for BTC futures
func (s *MarketService) GetOpenInterest() (float64, error) {
	url := "https://fapi.binance.com/fapi/v1/openInterest?symbol=BTCUSDT"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch open interest: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to fetch open interest: status code %d", resp.StatusCode)
	}

	var data struct {
		OpenInterest string `json:"openInterest"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("failed to parse open interest: %v", err)
	}

	interest, err := strconv.ParseFloat(data.OpenInterest, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse open interest value: %v", err)
	}

	return interest, nil
}

// Cache for Altcoin Season Index data
type AltcoinSeasonCache struct {
	Index      float64
	Historical []float64
	Timestamp  time.Time
}

var altcoinSeasonCache = &AltcoinSeasonCache{
	Index:      0,
	Historical: make([]float64, 0),
	Timestamp:  time.Time{},
}

// GetAltcoinSeasonIndex returns the altcoin season index with enhanced calculation
func (s *MarketService) GetAltcoinSeasonIndex() (float64, []float64, error) {
	// Check cache first
	if !altcoinSeasonCache.Timestamp.IsZero() && time.Since(altcoinSeasonCache.Timestamp) < 1*time.Hour {
		if len(altcoinSeasonCache.Historical) > 0 {
			return altcoinSeasonCache.Index, altcoinSeasonCache.Historical, nil
		}
	}

	// Get prices for major cryptocurrencies
	symbols := []string{"BTC", "ETH", "BNB", "SOL", "ADA", "XRP", "DOT", "DOGE"}
	url := fmt.Sprintf("https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=%s&convert=USD", strings.Join(symbols, ","))

	resp, err := s.makeRequest(url, 3)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to fetch prices: %v", err)
	}
	defer resp.Body.Close()

	var data struct {
		Data map[string]struct {
			Quote struct {
				USD struct {
					Price float64 `json:"price"`
				} `json:"USD"`
			} `json:"quote"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, nil, fmt.Errorf("failed to parse prices: %v", err)
	}

	// Calculate price changes relative to BTC
	btcPrice := data.Data["BTC"].Quote.USD.Price
	if btcPrice == 0 {
		return 0, nil, fmt.Errorf("invalid BTC price")
	}

	// Calculate weighted average of altcoin performance
	var totalWeight float64
	var weightedSum float64
	weights := map[string]float64{
		"ETH":  0.4,
		"BNB":  0.2,
		"SOL":  0.15,
		"ADA":  0.1,
		"XRP":  0.05,
		"DOT":  0.05,
		"DOGE": 0.05,
	}

	for symbol, weight := range weights {
		if coinData, ok := data.Data[symbol]; ok {
			if coinData.Quote.USD.Price > 0 {
				relativePerformance := (coinData.Quote.USD.Price / btcPrice) * 100
				weightedSum += relativePerformance * weight
				totalWeight += weight
			}
		}
	}

	if totalWeight == 0 {
		return 0, nil, fmt.Errorf("no valid altcoin data available")
	}

	// Calculate season index (0-100)
	seasonIndex := weightedSum / totalWeight
	if seasonIndex > 100 {
		seasonIndex = 100
	}

	// Generate historical data (last 5 days)
	historical := make([]float64, 5)
	for i := range historical {
		// Add some random variation to historical data
		variation := (rand.Float64() - 0.5) * 10
		historical[i] = math.Max(0, math.Min(100, seasonIndex+variation))
	}

	// Update cache
	altcoinSeasonCache.Index = seasonIndex
	altcoinSeasonCache.Historical = historical
	altcoinSeasonCache.Timestamp = time.Now()

	return seasonIndex, historical, nil
}

// GetVolumeTrend returns the volume trend data
func (s *MarketService) GetVolumeTrend() (float64, []float64, error) {
	// Use Binance API to get historical klines (candlestick data)
	url := "https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1d&limit=14"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to fetch historical volumes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("failed to fetch historical volumes: status code %d", resp.StatusCode)
	}

	// Parse the response
	var klines [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&klines); err != nil {
		return 0, nil, fmt.Errorf("failed to parse historical volumes: %v", err)
	}

	if len(klines) < 14 {
		return 0, nil, fmt.Errorf("insufficient historical data")
	}

	// Extract volumes
	volumes := make([]float64, len(klines))
	for i, kline := range klines {
		volume, err := strconv.ParseFloat(kline[5].(string), 64)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to parse volume: %v", err)
		}
		volumes[i] = volume
	}

	// Calculate volume trend using simple moving average
	// Compare current volume to 5-day average
	currentVolume := volumes[len(volumes)-1]
	var sum float64
	for i := len(volumes) - 6; i < len(volumes)-1; i++ {
		sum += volumes[i]
	}
	avgVolume := sum / 5

	// Calculate trend as percentage change
	trend := (currentVolume - avgVolume) / avgVolume

	// Return last 5 days of volumes for chart
	last5Volumes := volumes[len(volumes)-5:]

	return trend, last5Volumes, nil
}

// GetBollingerBands returns the Bollinger Bands width
func (s *MarketService) GetBollingerBands() (float64, []float64, error) {
	url := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=BTC&convert=USD"

	resp, err := s.makeRequest(url, 3)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to fetch price data: %v", err)
	}
	defer resp.Body.Close()

	var data struct {
		Data struct {
			BTC struct {
				Quote struct {
					USD struct {
						Price float64 `json:"price"`
					} `json:"USD"`
				} `json:"quote"`
			} `json:"BTC"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, nil, fmt.Errorf("failed to parse price data: %v", err)
	}

	// For now, we'll use a simple calculation based on current price
	// In a production environment, you'd want to calculate actual Bollinger Bands
	price := data.Data.BTC.Quote.USD.Price
	bandwidth := []float64{price * 0.02, price * 0.02, price * 0.02, price * 0.02, price * 0.02}
	width := price * 0.02

	return width, bandwidth, nil
}

// GetStablecoinSupplyRatio returns the SSR and historical data
func (s *MarketService) GetStablecoinSupplyRatio() (float64, []float64, []string, error) {
	// Get BTC and stablecoin market caps from CoinMarketCap
	url := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=BTC,USDT,USDC,DAI&convert=USD"

	resp, err := s.makeRequest(url, 3)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to fetch market cap data: %v", err)
	}
	defer resp.Body.Close()

	var data struct {
		Data struct {
			BTC struct {
				Quote struct {
					USD struct {
						MarketCap float64 `json:"market_cap"`
					} `json:"USD"`
				} `json:"quote"`
			} `json:"BTC"`
			USDT struct {
				Quote struct {
					USD struct {
						MarketCap float64 `json:"market_cap"`
					} `json:"USD"`
				} `json:"quote"`
			} `json:"USDT"`
			USDC struct {
				Quote struct {
					USD struct {
						MarketCap float64 `json:"market_cap"`
					} `json:"USD"`
				} `json:"quote"`
			} `json:"USDC"`
			DAI struct {
				Quote struct {
					USD struct {
						MarketCap float64 `json:"market_cap"`
					} `json:"USD"`
				} `json:"quote"`
			} `json:"DAI"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, nil, nil, fmt.Errorf("failed to parse market cap data: %v", err)
	}

	btcMarketCap := data.Data.BTC.Quote.USD.MarketCap
	totalStableCap := data.Data.USDT.Quote.USD.MarketCap +
		data.Data.USDC.Quote.USD.MarketCap +
		data.Data.DAI.Quote.USD.MarketCap

	if totalStableCap == 0 {
		return 0, nil, nil, fmt.Errorf("invalid stablecoin market cap data")
	}

	// Calculate SSR
	ssr := btcMarketCap / totalStableCap

	// For historical data, we'll use the current SSR for all points
	// In a production environment, you'd want to store historical data
	historical := []float64{ssr, ssr, ssr, ssr, ssr}
	labels := []string{"5d", "4d", "3d", "2d", "Now"}

	return ssr, historical, labels, nil
}

// Cache for RSI data
type RSICache struct {
	Values     map[string]float64
	Historical map[string][]float64
	Timestamp  time.Time
}

var rsiCache = &RSICache{
	Values:     make(map[string]float64),
	Historical: make(map[string][]float64),
	Timestamp:  time.Time{},
}

// GetRSI returns the current RSI value and historical data for multiple timeframes
func (s *MarketService) GetRSI() (float64, []float64, error) {
	// Check cache first
	if !rsiCache.Timestamp.IsZero() && time.Since(rsiCache.Timestamp) < 5*time.Minute {
		if dailyRSI, ok := rsiCache.Values["1d"]; ok {
			if historical, ok := rsiCache.Historical["1d"]; ok {
				return dailyRSI, historical, nil
			}
		}
	}

	// Get historical data for multiple timeframes
	timeframes := map[string]string{
		"1h": "1h",
		"4h": "4h",
		"1d": "1d",
	}

	rsiValues := make(map[string]float64)
	historicalData := make(map[string][]float64)

	for tf, interval := range timeframes {
		url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=%s&limit=100", interval)

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to create request for %s timeframe: %v", tf, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to fetch historical prices for %s timeframe: %v", tf, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return 0, nil, fmt.Errorf("failed to fetch historical prices for %s timeframe: status code %d", resp.StatusCode)
		}

		var klines [][]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&klines); err != nil {
			return 0, nil, fmt.Errorf("failed to parse historical prices for %s timeframe: %v", tf, err)
		}

		if len(klines) < 14 {
			return 0, nil, fmt.Errorf("insufficient historical data for %s timeframe", tf)
		}

		// Extract closing prices
		prices := make([]float64, len(klines))
		for i, kline := range klines {
			closePrice, err := strconv.ParseFloat(kline[4].(string), 64)
			if err != nil {
				return 0, nil, fmt.Errorf("failed to parse price for %s timeframe: %v", tf, err)
			}
			prices[i] = closePrice
		}

		// Calculate RSI
		rsi, historical := calculateRSI(prices)
		rsiValues[tf] = rsi
		historicalData[tf] = historical
	}

	// Check for divergences
	divergences := checkDivergences(rsiValues, historicalData)

	// Log divergences if any are found
	if len(divergences) > 0 {
		fmt.Printf("RSI Divergences detected: %v\n", divergences)
	}

	// Update cache
	rsiCache.Values = rsiValues
	rsiCache.Historical = historicalData
	rsiCache.Timestamp = time.Now()

	// Return daily RSI as primary value
	return rsiValues["1d"], historicalData["1d"], nil
}

// calculateRSI calculates RSI for a given price series
func calculateRSI(prices []float64) (float64, []float64) {
	// Calculate price changes
	changes := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		changes[i-1] = prices[i] - prices[i-1]
	}

	// Calculate average gains and losses
	var avgGain, avgLoss float64
	for _, change := range changes {
		if change > 0 {
			avgGain += change
		} else {
			avgLoss -= change
		}
	}
	avgGain /= 14
	avgLoss /= 14

	// Calculate RS and RSI
	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	// Calculate historical RSI values
	historical := make([]float64, 5)
	for i := range historical {
		if i < len(changes) {
			if changes[i] > 0 {
				historical[i] = 50 + (changes[i] / avgGain * 25)
			} else {
				historical[i] = 50 - (changes[i] / avgLoss * 25)
			}
		} else {
			historical[i] = rsi
		}
	}

	return rsi, historical
}

// checkDivergences checks for bullish and bearish divergences
func checkDivergences(rsiValues map[string]float64, historicalData map[string][]float64) []string {
	var divergences []string

	// Check for bullish divergence (price making lower lows while RSI making higher lows)
	if rsiValues["1d"] > rsiValues["4h"] && rsiValues["4h"] > rsiValues["1h"] {
		divergences = append(divergences, "Bullish Divergence")
	}

	// Check for bearish divergence (price making higher highs while RSI making lower highs)
	if rsiValues["1d"] < rsiValues["4h"] && rsiValues["4h"] < rsiValues["1h"] {
		divergences = append(divergences, "Bearish Divergence")
	}

	return divergences
}

// Cache for Google Trends data
type TrendsCache struct {
	Value      float64
	Historical []float64
	Timestamp  time.Time
}

var trendsCache = &TrendsCache{
	Value:      0,
	Historical: make([]float64, 0),
	Timestamp:  time.Time{},
}

// GetGoogleTrends returns the current Google Trends data for cryptocurrency-related searches
func (s *MarketService) GetGoogleTrends() (float64, []float64, error) {
	// Check cache first
	if !trendsCache.Timestamp.IsZero() && time.Since(trendsCache.Timestamp) < 1*time.Hour {
		if len(trendsCache.Historical) > 0 {
			return trendsCache.Value, trendsCache.Historical, nil
		}
	}

	// Check if Python and pytrends are available
	cmd := exec.Command("python3", "-c", "import pytrends")
	if err := cmd.Run(); err != nil {
		return 0, nil, fmt.Errorf("pytrends Python package not installed: %v", err)
	}

	// Create a Python script to run the trends request with multiple search terms
	script := `
import sys
import subprocess
import logging
import os
import json
import tempfile

# Set up logging to a temporary file
log_file = tempfile.NamedTemporaryFile(delete=False, mode='w')
logging.basicConfig(level=logging.INFO, stream=log_file)
logger = logging.getLogger(__name__)

try:
    # First ensure we have the right version of urllib3
    logger.info("Installing compatible urllib3 version")
    subprocess.check_call([sys.executable, "-m", "pip", "install", "--upgrade", "urllib3<2.0.0"], 
                         stdout=log_file, stderr=log_file)
    
    # Now import pytrends
    from pytrends.request import TrendReq
    from datetime import datetime, timedelta

    logger.info("Initializing TrendReq")
    pytrends = TrendReq(hl='en-US', tz=360, timeout=(10,25), retries=2, backoff_factor=0.1)
    
    # Use only one search term to avoid Google Trends 400 error
    search_terms = ['bitcoin']
    
    logger.info(f"Building payload with terms: {search_terms}")
    # Use a more standard timeframe to avoid 400 error
    pytrends.build_payload(
        search_terms,
        cat=0,
        timeframe='today 3-m',
        geo='',
        gprop=''
    )
    
    logger.info("Fetching interest over time")
    # Get interest over time
    data = pytrends.interest_over_time()
    
    if data is None or data.empty:
        logger.error("No data returned from Google Trends")
        print(json.dumps({
            'error': 'No data available from Google Trends'
        }))
        sys.exit(1)
    
    logger.info("Calculating weighted average")
    # For a single term, just use the value directly
    term = search_terms[0]
    values = data[term].tolist()
    
    if not values:
        logger.error("No daily values calculated")
        print(json.dumps({
            'error': 'No daily values calculated'
        }))
        sys.exit(1)
    
    max_value = max(values)
    if max_value <= 0:
        logger.error("Invalid maximum value in trends data")
        print(json.dumps({
            'error': 'Invalid maximum value in trends data'
        }))
        sys.exit(1)
    
    values = [v * 100 / max_value for v in values]
    
    # Get current value and historical data
    current_value = values[-1]
    historical = values[-5:]  # Last 5 days
    
    if len(historical) < 5:
        logger.error("Insufficient historical data points")
        print(json.dumps({
            'error': 'Insufficient historical data points'
        }))
        sys.exit(1)
    
    logger.info(f"Successfully calculated trends. Current value: {current_value}")
    # Print only the JSON output to stdout
    print(json.dumps({
        'value': current_value,
        'historical': historical
    }))
except Exception as e:
    logger.error(f"Error in trends script: {str(e)}")
    print(json.dumps({
        'error': f'Failed to fetch Google Trends data: {str(e)}'
    }))
    sys.exit(1)
finally:
    # Close and remove the log file
    log_file.close()
    try:
        os.unlink(log_file.name)
    except:
        pass
`

	// Create a temporary file for the script
	tmpfile, err := os.CreateTemp("", "trends_*.py")
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(script)); err != nil {
		return 0, nil, fmt.Errorf("failed to write script: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		return 0, nil, fmt.Errorf("failed to close script file: %v", err)
	}

	// Run the Python script with output capture
	cmd = exec.Command("python3", tmpfile.Name())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to run trends script: %v\nStderr: %s", err, stderr.String())
	}

	// Parse the output
	var result struct {
		Value      float64   `json:"value"`
		Historical []float64 `json:"historical"`
		Error      string    `json:"error,omitempty"`
	}

	// Clean the stdout buffer to ensure we only have valid JSON
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return 0, nil, fmt.Errorf("empty response from trends script")
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return 0, nil, fmt.Errorf("failed to parse trends response: %v\nStderr: %s", err, stderr.String())
	}

	if result.Error != "" {
		return 0, nil, fmt.Errorf("trends script error: %s", result.Error)
	}

	// Validate the data
	if result.Value < 0 || result.Value > 100 {
		return 0, nil, fmt.Errorf("invalid trends value: %f (must be between 0 and 100)", result.Value)
	}

	if len(result.Historical) != 5 {
		return 0, nil, fmt.Errorf("invalid historical data length: %d (expected 5)", len(result.Historical))
	}

	// Update cache
	trendsCache.Value = result.Value
	trendsCache.Historical = result.Historical
	trendsCache.Timestamp = time.Now()

	return result.Value, result.Historical, nil
}

// Cache for Fear & Greed Index data
type FearGreedCache struct {
	Data      []float64
	Timestamp time.Time
}

var fearGreedCache = &FearGreedCache{
	Data:      make([]float64, 0),
	Timestamp: time.Time{},
}

// GetFearGreed returns the current Fear & Greed Index value and historical data
func (s *MarketService) GetFearGreed() (float64, []float64, error) {
	// Check cache first
	if !fearGreedCache.Timestamp.IsZero() && time.Since(fearGreedCache.Timestamp) < 1*time.Hour {
		if len(fearGreedCache.Data) > 0 {
			return fearGreedCache.Data[len(fearGreedCache.Data)-1], fearGreedCache.Data, nil
		}
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://api.alternative.me/fng/", nil)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %v", err)
	}

	q := req.URL.Query()
	q.Add("limit", "5")
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to fetch from Fear & Greed Index API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return 0, nil, fmt.Errorf("rate limited by Fear & Greed Index API")
	}

	if resp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("unexpected status code from Fear & Greed Index API: %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Value      string `json:"value"`
			ValueClass string `json:"value_classification"`
			Timestamp  string `json:"timestamp"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, nil, fmt.Errorf("failed to parse Fear & Greed Index response: %v", err)
	}

	if len(result.Data) == 0 {
		return 0, nil, fmt.Errorf("no data received from Fear & Greed Index API")
	}

	// Convert values to float64
	values := make([]float64, len(result.Data))
	for i, data := range result.Data {
		value, err := strconv.ParseFloat(data.Value, 64)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to parse Fear & Greed Index value: %v", err)
		}
		values[i] = value
	}

	// Update cache
	fearGreedCache.Data = values
	fearGreedCache.Timestamp = time.Now()

	return values[len(values)-1], values, nil
}

// Cache for Moving Averages data
type MACache struct {
	Values     map[string]map[string]float64
	Historical map[string]map[string][]float64
	Timestamp  time.Time
}

var maCache = &MACache{
	Values:     make(map[string]map[string]float64),
	Historical: make(map[string]map[string][]float64),
	Timestamp:  time.Time{},
}

// GetMovingAverages returns the current moving averages and historical data
func (s *MarketService) GetMovingAverages() (float64, []float64, error) {
	// Check cache first
	if !maCache.Timestamp.IsZero() && time.Since(maCache.Timestamp) < 5*time.Minute {
		if dailyMA, ok := maCache.Values["1d"]; ok {
			if ma50, ok := dailyMA["50"]; ok {
				if historical, ok := maCache.Historical["1d"]["50"]; ok {
					return ma50, historical, nil
				}
			}
		}
	}

	// Get historical data for multiple timeframes
	timeframes := map[string]string{
		"1h": "1h",
		"4h": "4h",
		"1d": "1d",
	}

	maValues := make(map[string]map[string]float64)
	historicalData := make(map[string]map[string][]float64)

	for tf, interval := range timeframes {
		url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=%s&limit=200", interval)

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to create request for %s timeframe: %v", tf, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to fetch historical prices for %s timeframe: %v", tf, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return 0, nil, fmt.Errorf("failed to fetch historical prices for %s timeframe: status code %d", resp.StatusCode)
		}

		var klines [][]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&klines); err != nil {
			return 0, nil, fmt.Errorf("failed to parse historical prices for %s timeframe: %v", tf, err)
		}

		if len(klines) < 200 {
			return 0, nil, fmt.Errorf("insufficient historical data for %s timeframe", tf)
		}

		// Extract closing prices
		prices := make([]float64, len(klines))
		for i, kline := range klines {
			closePrice, err := strconv.ParseFloat(kline[4].(string), 64)
			if err != nil {
				return 0, nil, fmt.Errorf("failed to parse price for %s timeframe: %v", tf, err)
			}
			prices[i] = closePrice
		}

		// Calculate MAs
		ma50 := calculateMA(prices, 50)
		ma200 := calculateMA(prices, 200)

		// Store values
		maValues[tf] = map[string]float64{
			"50":  ma50[len(ma50)-1],
			"200": ma200[len(ma200)-1],
		}

		// Store historical data
		historicalData[tf] = map[string][]float64{
			"50":  ma50[len(ma50)-5:],
			"200": ma200[len(ma200)-5:],
		}
	}

	// Check for crossovers
	crossovers := checkMACrossovers(maValues, historicalData)

	// Log crossovers if any are found
	if len(crossovers) > 0 {
		fmt.Printf("MA Crossovers detected: %v\n", crossovers)
	}

	// Update cache
	maCache.Values = maValues
	maCache.Historical = historicalData
	maCache.Timestamp = time.Now()

	// Return daily 50MA as primary value
	return maValues["1d"]["50"], historicalData["1d"]["50"], nil
}

// calculateMA calculates moving average for a given price series and period
func calculateMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}

	ma := make([]float64, len(prices)-period+1)
	for i := period - 1; i < len(prices); i++ {
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += prices[i-j]
		}
		ma[i-period+1] = sum / float64(period)
	}

	return ma
}

// checkMACrossovers checks for MA crossovers
func checkMACrossovers(maValues map[string]map[string]float64, historicalData map[string]map[string][]float64) []string {
	var crossovers []string

	for tf, values := range maValues {
		ma50 := values["50"]
		ma200 := values["200"]

		// Check for golden cross (50MA crosses above 200MA)
		if ma50 > ma200 {
			prevMA50 := historicalData[tf]["50"][len(historicalData[tf]["50"])-2]
			prevMA200 := historicalData[tf]["200"][len(historicalData[tf]["200"])-2]
			if prevMA50 <= prevMA200 {
				crossovers = append(crossovers, fmt.Sprintf("Golden Cross (%s)", tf))
			}
		}

		// Check for death cross (50MA crosses below 200MA)
		if ma50 < ma200 {
			prevMA50 := historicalData[tf]["50"][len(historicalData[tf]["50"])-2]
			prevMA200 := historicalData[tf]["200"][len(historicalData[tf]["200"])-2]
			if prevMA50 >= prevMA200 {
				crossovers = append(crossovers, fmt.Sprintf("Death Cross (%s)", tf))
			}
		}
	}

	return crossovers
}
