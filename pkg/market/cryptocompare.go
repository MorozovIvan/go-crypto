package market

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CryptoCompareService struct {
	apiKey  string
	baseURL string
}

func NewCryptoCompareService() (*CryptoCompareService, error) {
	apiKey := "cc50bf26846d67683869a720ed1ffe3861535d540bb7ebcb370fa3aefb19b8dd"
	if apiKey == "" {
		return nil, fmt.Errorf("CRYPTOCOMPARE_API_KEY not set")
	}
	return &CryptoCompareService{
		apiKey:  apiKey,
		baseURL: "https://min-api.cryptocompare.com/data",
	}, nil
}

// OHLCV represents Open, High, Low, Close, Volume data with additional peak tracking
type OHLCV struct {
	Time         time.Time `json:"time"`
	Open         float64   `json:"open"`
	High         float64   `json:"high"`
	Low          float64   `json:"low"`
	Close        float64   `json:"close"`
	Volume       float64   `json:"volume"`
	VolumeTO     float64   `json:"volumeto"`
	IsLocalPeak  bool      `json:"isLocalPeak"`  // Indicates if this point is a local peak
	IsSignalPeak bool      `json:"isSignalPeak"` // Indicates if this point is a significant peak for signals
}

// CryptoCompareOHLCVResponse represents the API response structure
type CryptoCompareOHLCVResponse struct {
	Response string `json:"Response"`
	Message  string `json:"Message"`
	Data     struct {
		Data []struct {
			Time       int64   `json:"time"`
			High       float64 `json:"high"`
			Low        float64 `json:"low"`
			Open       float64 `json:"open"`
			Close      float64 `json:"close"`
			VolumeFrom float64 `json:"volumefrom"`
			VolumeTo   float64 `json:"volumeto"`
		} `json:"Data"`
	} `json:"Data"`
}

// GetHistoricalOHLCV fetches historical OHLCV data for a given symbol pair
func (s *CryptoCompareService) GetHistoricalOHLCV(fsym, tsym string, limit int) ([]OHLCV, error) {
	// For Solana contract addresses, use a different endpoint
	if len(fsym) > 30 { // Solana addresses are typically longer than 30 characters
		url := fmt.Sprintf("%s/v2/histohour?fsym=SOL&tsym=%s&limit=%d&api_key=%s&contract=%s",
			s.baseURL, tsym, limit, s.apiKey, fsym)
		return s.fetchOHLCV(url)
	}

	// For regular tokens
	url := fmt.Sprintf("%s/v2/histohour?fsym=%s&tsym=%s&limit=%d&api_key=%s",
		s.baseURL, fsym, tsym, limit, s.apiKey)
	return s.fetchOHLCV(url)
}

func (s *CryptoCompareService) fetchOHLCV(url string) ([]OHLCV, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	var response struct {
		Response string `json:"Response"`
		Message  string `json:"Message"`
		Data     struct {
			Data []struct {
				Time   int64   `json:"time"`
				Open   float64 `json:"open"`
				High   float64 `json:"high"`
				Low    float64 `json:"low"`
				Close  float64 `json:"close"`
				Volume float64 `json:"volumefrom"`
			} `json:"Data"`
		} `json:"Data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Response == "Error" {
		return nil, fmt.Errorf("API error: %s", response.Message)
	}

	var result []OHLCV
	for _, d := range response.Data.Data {
		result = append(result, OHLCV{
			Time:   time.Unix(d.Time, 0),
			Open:   d.Open,
			High:   d.High,
			Low:    d.Low,
			Close:  d.Close,
			Volume: d.Volume,
		})
	}

	return result, nil
}

// GetSolanaPeaks fetches historical data specifically for Solana and identifies peaks
func (s *CryptoCompareService) GetSolanaPeaks(days int) ([]OHLCV, error) {
	// Get more data than requested to properly identify peaks at the edges
	data, err := s.GetHistoricalOHLCV("SOL", "USD", days+2)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Solana data: %v", err)
	}

	// Filter to only include signal peaks and the requested number of days
	var peaks []OHLCV
	for _, d := range data {
		if d.IsSignalPeak {
			peaks = append(peaks, d)
		}
	}

	// If we have more data than requested, trim to the requested period
	if len(peaks) > days {
		peaks = peaks[len(peaks)-days:]
	}

	return peaks, nil
}

// CryptoComparePriceResponse represents the current price API response structure
type CryptoComparePriceResponse struct {
	USD float64 `json:"USD"`
}

// GetCurrentPrice returns the current price for a given symbol pair
func (s *CryptoCompareService) GetCurrentPrice(symbol, quoteCurrency string) (float64, error) {
	url := fmt.Sprintf("%s/price?fsym=%s&tsyms=%s&api_key=%s", s.baseURL, symbol, quoteCurrency, s.apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	var response CryptoComparePriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %v", err)
	}

	return response.USD, nil
}

// TopPairResponse represents the top pairs API response structure
type TopPairResponse struct {
	Data []struct {
		CoinInfo struct {
			Name string `json:"Name"`
		} `json:"CoinInfo"`
		RAW struct {
			USD struct {
				PRICE     float64 `json:"PRICE"`
				VOLUME24H float64 `json:"VOLUME24HOUR"`
			} `json:"USD"`
		} `json:"RAW"`
	} `json:"Data"`
}

// GetTopTradingPairs returns the top trading pairs by volume
func (s *CryptoCompareService) GetTopTradingPairs(limit int) ([]struct {
	FromSymbol string  `json:"fromSymbol"`
	ToSymbol   string  `json:"toSymbol"`
	Volume24h  float64 `json:"volume24h"`
	Price      float64 `json:"price"`
}, error) {
	url := fmt.Sprintf("%s/top/totalvolfull?limit=%d&tsym=USD&api_key=%s", s.baseURL, limit, s.apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	var response TopPairResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	result := make([]struct {
		FromSymbol string  `json:"fromSymbol"`
		ToSymbol   string  `json:"toSymbol"`
		Volume24h  float64 `json:"volume24h"`
		Price      float64 `json:"price"`
	}, len(response.Data))

	for i, pair := range response.Data {
		result[i] = struct {
			FromSymbol string  `json:"fromSymbol"`
			ToSymbol   string  `json:"toSymbol"`
			Volume24h  float64 `json:"volume24h"`
			Price      float64 `json:"price"`
		}{
			FromSymbol: pair.CoinInfo.Name,
			ToSymbol:   "USD",
			Volume24h:  pair.RAW.USD.VOLUME24H,
			Price:      pair.RAW.USD.PRICE,
		}
	}

	return result, nil
}
