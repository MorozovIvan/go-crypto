package market

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type BinanceService struct {
	apiKey    string
	apiSecret string
	baseURL   string
}

func NewBinanceService() *BinanceService {
	// For testing, use the real API key directly
	apiKey := os.Getenv("BINANCE_API_KEY")

	apiSecret := os.Getenv("BINANCE_API_SECRET")
	// Note: API secret is required for private endpoints like account info

	return &BinanceService{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   "https://api.binance.com",
	}
}

type Asset struct {
	Symbol       string       `json:"symbol"`
	Amount       float64      `json:"amount"`
	Value        float64      `json:"value"`
	Change       float64      `json:"change"`
	High24h      float64      `json:"high24h"`
	Low24h       float64      `json:"low24h"`
	Volume24h    float64      `json:"volume24h"`
	MarketCap    float64      `json:"marketCap"`
	PriceHistory []PricePoint `json:"priceHistory"`
}

type PricePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Price     float64   `json:"price"`
}

type PortfolioResponse struct {
	Assets          []Asset `json:"assets"`
	TotalValue      float64 `json:"totalValue"`
	PortfolioChange float64 `json:"portfolioChange"`
	Volume24h       float64 `json:"volume24h"`
}

func (s *BinanceService) GetPortfolio() (*PortfolioResponse, error) {
	// Check if API key is available
	if s.apiKey == "" {
		return s.getMockPortfolio(), nil
	}

	// Check if API secret is available (required for private endpoints)
	if s.apiSecret == "" {
		fmt.Printf("Warning: BINANCE_API_SECRET not set. Private endpoints require both API key and secret. Returning mock data.\n")
		return s.getMockPortfolio(), nil
	}

	// Get account information
	accountInfo, err := s.getAccountInfo()
	if err != nil {
		// Return mock data on error
		fmt.Printf("Warning: Failed to get real Binance data, returning mock data: %v\n", err)
		return s.getMockPortfolio(), nil
	}

	// Get prices for all assets
	prices, err := s.getAllPrices()
	if err != nil {
		// Return mock data on error
		fmt.Printf("Warning: Failed to get real Binance prices, returning mock data: %v\n", err)
		return s.getMockPortfolio(), nil
	}

	// Get 24h statistics for all assets
	stats24h, err := s.get24hStats()
	if err != nil {
		// Return mock data on error
		fmt.Printf("Warning: Failed to get real Binance stats, returning mock data: %v\n", err)
		return s.getMockPortfolio(), nil
	}

	// Process assets
	var assets []Asset
	var totalValue float64
	var previousTotalValue float64

	for _, balance := range accountInfo.Balances {
		if balance.Free == 0 && balance.Locked == 0 {
			continue
		}

		amount := balance.Free + balance.Locked
		symbol := balance.Asset

		// Skip assets that don't have a USDT pair
		price, ok := prices[symbol+"USDT"]
		if !ok {
			continue
		}

		value := amount * price
		totalValue += value

		// Get 24h statistics
		stat24h := stats24h[symbol+"USDT"]
		change := 0.0
		if stat24h != nil {
			change = stat24h.PriceChangePercent
		}

		// Get historical prices
		priceHistory, err := s.getHistoricalPrices(symbol + "USDT")
		if err != nil {
			fmt.Printf("Warning: failed to get historical prices for %s: %v\n", symbol, err)
		}

		asset := Asset{
			Symbol:       symbol,
			Amount:       amount,
			Value:        value,
			Change:       change,
			High24h:      stat24h.HighPrice,
			Low24h:       stat24h.LowPrice,
			Volume24h:    stat24h.Volume * price,
			MarketCap:    stat24h.Volume * price * 24, // Rough estimate
			PriceHistory: priceHistory,
		}

		assets = append(assets, asset)
	}

	// Calculate portfolio change
	portfolioChange := 0.0
	if previousTotalValue > 0 {
		portfolioChange = ((totalValue - previousTotalValue) / previousTotalValue) * 100
	}

	return &PortfolioResponse{
		Assets:          assets,
		TotalValue:      totalValue,
		PortfolioChange: portfolioChange,
		Volume24h:       totalValue * 0.1, // Rough estimate
	}, nil
}

// getMockPortfolio returns mock portfolio data for demonstration
func (s *BinanceService) getMockPortfolio() *PortfolioResponse {
	now := time.Now()

	// Create mock price history
	createPriceHistory := func(basePrice float64, volatility float64) []PricePoint {
		history := make([]PricePoint, 7)
		for i := 0; i < 7; i++ {
			timestamp := now.AddDate(0, 0, -6+i)
			// Simulate price movement
			priceChange := (volatility * 2 * (0.5 - float64(i%3)/3.0))
			price := basePrice * (1 + priceChange/100)
			history[i] = PricePoint{
				Timestamp: timestamp,
				Price:     price,
			}
		}
		return history
	}

	assets := []Asset{
		{
			Symbol:       "BTC",
			Amount:       0.25,
			Value:        15750.00,
			Change:       2.45,
			High24h:      64200.00,
			Low24h:       62800.00,
			Volume24h:    1250000.00,
			MarketCap:    1200000000000.00,
			PriceHistory: createPriceHistory(63000.00, 3.2),
		},
		{
			Symbol:       "ETH",
			Amount:       8.5,
			Value:        27200.00,
			Change:       -1.23,
			High24h:      3250.00,
			Low24h:       3180.00,
			Volume24h:    850000.00,
			MarketCap:    380000000000.00,
			PriceHistory: createPriceHistory(3200.00, 4.1),
		},
		{
			Symbol:       "BNB",
			Amount:       45.2,
			Value:        18080.00,
			Change:       0.87,
			High24h:      410.00,
			Low24h:       395.00,
			Volume24h:    320000.00,
			MarketCap:    65000000000.00,
			PriceHistory: createPriceHistory(400.00, 2.8),
		},
		{
			Symbol:       "ADA",
			Amount:       12500.0,
			Value:        5000.00,
			Change:       3.21,
			High24h:      0.42,
			Low24h:       0.38,
			Volume24h:    180000.00,
			MarketCap:    14000000000.00,
			PriceHistory: createPriceHistory(0.40, 5.5),
		},
		{
			Symbol:       "DOT",
			Amount:       850.0,
			Value:        5950.00,
			Change:       -2.15,
			High24h:      7.20,
			Low24h:       6.95,
			Volume24h:    95000.00,
			MarketCap:    8500000000.00,
			PriceHistory: createPriceHistory(7.00, 3.7),
		},
		{
			Symbol:       "LINK",
			Amount:       420.0,
			Value:        6720.00,
			Change:       1.98,
			High24h:      16.50,
			Low24h:       15.80,
			Volume24h:    125000.00,
			MarketCap:    9200000000.00,
			PriceHistory: createPriceHistory(16.00, 4.3),
		},
	}

	totalValue := 0.0
	totalVolume := 0.0
	for _, asset := range assets {
		totalValue += asset.Value
		totalVolume += asset.Volume24h
	}

	return &PortfolioResponse{
		Assets:          assets,
		TotalValue:      totalValue,
		PortfolioChange: 1.45, // Mock portfolio change
		Volume24h:       totalVolume,
	}
}

type AccountInfo struct {
	Balances []Balance `json:"balances"`
}

type Balance struct {
	Asset  string  `json:"asset"`
	Free   float64 `json:"free,string"`
	Locked float64 `json:"locked,string"`
}

func (s *BinanceService) getAccountInfo() (*AccountInfo, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create query parameters
	params := url.Values{}
	params.Set("timestamp", fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond)))

	// Create signature
	signature := s.createSignature(params.Encode())
	params.Set("signature", signature)

	req, err := http.NewRequest("GET", s.baseURL+"/api/v3/account?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", s.apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var accountInfo AccountInfo
	if err := json.NewDecoder(resp.Body).Decode(&accountInfo); err != nil {
		return nil, err
	}

	return &accountInfo, nil
}

// createSignature creates HMAC-SHA256 signature for Binance API
func (s *BinanceService) createSignature(queryString string) string {
	if s.apiSecret == "" {
		return ""
	}

	h := hmac.New(sha256.New, []byte(s.apiSecret))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}

type PriceResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func (s *BinanceService) getAllPrices() (map[string]float64, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", s.baseURL+"/api/v3/ticker/price", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var prices []PriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&prices); err != nil {
		return nil, err
	}

	priceMap := make(map[string]float64)
	for _, price := range prices {
		value, err := strconv.ParseFloat(price.Price, 64)
		if err != nil {
			continue
		}
		priceMap[price.Symbol] = value
	}

	return priceMap, nil
}

type Stats24h struct {
	Symbol             string  `json:"symbol"`
	PriceChangePercent float64 `json:"priceChangePercent,string"`
	HighPrice          float64 `json:"highPrice,string"`
	LowPrice           float64 `json:"lowPrice,string"`
	Volume             float64 `json:"volume,string"`
}

func (s *BinanceService) get24hStats() (map[string]*Stats24h, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", s.baseURL+"/api/v3/ticker/24hr", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stats []Stats24h
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	statsMap := make(map[string]*Stats24h)
	for i := range stats {
		statsMap[stats[i].Symbol] = &stats[i]
	}

	return statsMap, nil
}

type Kline struct {
	OpenTime            int64  `json:"openTime"`
	Open                string `json:"open"`
	High                string `json:"high"`
	Low                 string `json:"low"`
	Close               string `json:"close"`
	Volume              string `json:"volume"`
	CloseTime           int64  `json:"closeTime"`
	QuoteVolume         string `json:"quoteVolume"`
	Trades              int    `json:"trades"`
	TakerBuyVolume      string `json:"takerBuyVolume"`
	TakerBuyQuoteVolume string `json:"takerBuyQuoteVolume"`
}

func (s *BinanceService) getHistoricalPrices(symbol string) ([]PricePoint, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Get last 30 days of daily candles
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v3/klines?symbol=%s&interval=1d&limit=30", s.baseURL, symbol), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var klines [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&klines); err != nil {
		return nil, err
	}

	var priceHistory []PricePoint
	for _, kline := range klines {
		closePrice, err := strconv.ParseFloat(kline[4].(string), 64)
		if err != nil {
			continue
		}

		timestamp := time.Unix(int64(kline[0].(float64))/1000, 0)
		priceHistory = append(priceHistory, PricePoint{
			Timestamp: timestamp,
			Price:     closePrice,
		})
	}

	return priceHistory, nil
}
