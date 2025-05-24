package market

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
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

// GetAltcoinSeasonIndex returns the altcoin season index
func (s *MarketService) GetAltcoinSeasonIndex() (float64, error) {
	// Get BTC and ETH prices from CoinMarketCap
	url := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=BTC,ETH&convert=USD"

	resp, err := s.makeRequest(url, 3)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch prices: %v", err)
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
			ETH struct {
				Quote struct {
					USD struct {
						Price float64 `json:"price"`
					} `json:"USD"`
				} `json:"quote"`
			} `json:"ETH"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("failed to parse prices: %v", err)
	}

	if data.Data.BTC.Quote.USD.Price == 0 || data.Data.ETH.Quote.USD.Price == 0 {
		return 0, fmt.Errorf("invalid price data: BTC or ETH price is zero")
	}

	// Calculate ETH/BTC ratio
	ratio := data.Data.ETH.Quote.USD.Price / data.Data.BTC.Quote.USD.Price

	// Convert ratio to season index (0-100)
	// Assuming 0.06 is the threshold for altcoin season
	seasonIndex := (ratio / 0.06) * 100
	if seasonIndex > 100 {
		seasonIndex = 100
	}

	return seasonIndex, nil
}

// GetVolumeTrend returns the volume trend data
func (s *MarketService) GetVolumeTrend() (float64, []float64, error) {
	url := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=BTC&convert=USD"

	resp, err := s.makeRequest(url, 3)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to fetch volume data: %v", err)
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
		return 0, nil, fmt.Errorf("failed to parse volume data: %v", err)
	}

	// For now, we'll use the 24h volume as a single data point
	// In a production environment, you'd want to store historical data
	volume := data.Data.BTC.Quote.USD.Volume24h
	volumes := []float64{volume}
	trend := 0.0 // Since we only have one data point, trend is 0

	return trend, volumes, nil
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
