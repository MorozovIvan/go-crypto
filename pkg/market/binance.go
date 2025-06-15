package market

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	return &BinanceService{
		apiKey:    os.Getenv("BINANCE_API_KEY"),
		apiSecret: os.Getenv("BINANCE_API_SECRET"),
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
	// Get account information
	accountInfo, err := s.getAccountInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %v", err)
	}

	// Get prices for all assets
	prices, err := s.getAllPrices()
	if err != nil {
		return nil, fmt.Errorf("failed to get prices: %v", err)
	}

	// Get 24h statistics for all assets
	stats24h, err := s.get24hStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get 24h stats: %v", err)
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

	req, err := http.NewRequest("GET", s.baseURL+"/api/v3/account", nil)
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
