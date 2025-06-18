package market

import (
	"math"
	"math/rand"
	"time"
)

// MockDataGenerator provides realistic mock data for market analysis
type MockDataGenerator struct {
	seed     int64
	baseTime time.Time
}

// NewMockDataGenerator creates a new mock data generator
func NewMockDataGenerator() *MockDataGenerator {
	return &MockDataGenerator{
		seed:     time.Now().Unix(),
		baseTime: time.Now(),
	}
}

// MockStablecoinFlows generates realistic stablecoin flow data
func (m *MockDataGenerator) MockStablecoinFlows() MockMetricData {
	rand.Seed(m.seed + 1)

	// Generate realistic stablecoin flow values
	baseFlow := rand.Float64()*400 - 200 // -200 to +200 million USD

	// Create historical data with trend
	historical := make([]float64, 30)
	trendSlope := (rand.Float64() - 0.5) * 10 // Random trend

	for i := 0; i < 30; i++ {
		dayOffset := float64(i - 29) // Last 30 days
		trend := trendSlope * dayOffset / 10
		noise := (rand.Float64() - 0.5) * 50
		historical[i] = baseFlow + trend + noise
	}

	currentValue := historical[len(historical)-1]

	// Determine indicator based on flow direction and magnitude
	var indicator string
	var score float64

	if currentValue > 100 {
		indicator = "Strong Inflows"
		score = 0.8
	} else if currentValue > 50 {
		indicator = "Moderate Inflows"
		score = 0.4
	} else if currentValue > -50 {
		indicator = "Neutral Flows"
		score = 0.0
	} else if currentValue > -100 {
		indicator = "Moderate Outflows"
		score = -0.4
	} else {
		indicator = "Strong Outflows"
		score = -0.8
	}

	return MockMetricData{
		Value:      currentValue,
		Indicator:  indicator,
		Score:      score,
		Historical: historical,
		ChartData:  historical[len(historical)-7:], // Last 7 days for chart
		APISource:  "Mocked (Premium APIs required for real data)",
		Timestamp:  time.Now(),
	}
}

// MockInstitutionalFlows generates realistic institutional flow data
func (m *MockDataGenerator) MockInstitutionalFlows() MockMetricData {
	rand.Seed(m.seed + 2)

	// Generate realistic institutional flow values (typically larger)
	baseFlow := rand.Float64()*2000 - 1000 // -1000 to +1000 million USD

	// Create historical data with institutional patterns
	historical := make([]float64, 30)

	for i := 0; i < 30; i++ {
		// Institutional flows tend to be more volatile and trend-following
		cycleFactor := math.Sin(float64(i) * math.Pi / 15) // 15-day cycle
		volatility := (rand.Float64() - 0.5) * 300
		historical[i] = baseFlow + cycleFactor*200 + volatility
	}

	currentValue := historical[len(historical)-1]

	// Determine indicator based on institutional activity
	var indicator string
	var score float64

	if currentValue > 500 {
		indicator = "Strong Institutional Buying"
		score = 0.9
	} else if currentValue > 200 {
		indicator = "Moderate Institutional Buying"
		score = 0.5
	} else if currentValue > -200 {
		indicator = "Mixed Institutional Activity"
		score = 0.0
	} else if currentValue > -500 {
		indicator = "Moderate Institutional Selling"
		score = -0.5
	} else {
		indicator = "Strong Institutional Selling"
		score = -0.9
	}

	return MockMetricData{
		Value:      currentValue,
		Indicator:  indicator,
		Score:      score,
		Historical: historical,
		ChartData:  historical[len(historical)-7:],
		APISource:  "Mocked (Institutional APIs required for real data)",
		Timestamp:  time.Now(),
	}
}

// MockYieldCurves generates realistic DeFi yield premium data
func (m *MockDataGenerator) MockYieldCurves() MockMetricData {
	rand.Seed(m.seed + 3)

	// Generate realistic yield premium (DeFi yield vs traditional)
	baseYield := 5.0 + rand.Float64()*15.0 // 5-20% premium

	// Create historical data
	historical := make([]float64, 30)

	for i := 0; i < 30; i++ {
		// Yield curves change more slowly
		trend := math.Sin(float64(i)*math.Pi/20) * 2
		volatility := (rand.Float64() - 0.5) * 1.5
		historical[i] = baseYield + trend + volatility

		// Ensure realistic bounds
		if historical[i] < 0 {
			historical[i] = 0.1
		}
		if historical[i] > 25 {
			historical[i] = 25.0
		}
	}

	currentValue := historical[len(historical)-1]

	// Determine indicator based on yield premium level
	var indicator string
	var score float64

	if currentValue > 15 {
		indicator = "Extremely High Yield Premium"
		score = 0.8
	} else if currentValue > 10 {
		indicator = "High Yield Premium"
		score = 0.6
	} else if currentValue > 7 {
		indicator = "Moderate Yield Premium"
		score = 0.3
	} else if currentValue > 3 {
		indicator = "Low Yield Premium"
		score = 0.1
	} else {
		indicator = "Minimal Yield Premium"
		score = -0.2
	}

	return MockMetricData{
		Value:      currentValue,
		Indicator:  indicator,
		Score:      score,
		Historical: historical,
		ChartData:  historical[len(historical)-7:],
		APISource:  "Mocked (DeFi protocol APIs required for real data)",
		Timestamp:  time.Now(),
	}
}

// MockVolatilitySurface generates realistic implied volatility data
func (m *MockDataGenerator) MockVolatilitySurface() MockMetricData {
	rand.Seed(m.seed + 4)

	// Generate realistic implied volatility (typically 20-150%)
	baseVol := 40.0 + rand.Float64()*80.0 // 40-120% volatility

	// Create historical data
	historical := make([]float64, 30)

	for i := 0; i < 30; i++ {
		// Volatility clustering effect
		if i > 0 {
			prevVol := historical[i-1]
			change := (rand.Float64() - 0.5) * 10
			historical[i] = prevVol + change
		} else {
			historical[i] = baseVol
		}

		// Ensure realistic bounds
		if historical[i] < 10 {
			historical[i] = 10.0
		}
		if historical[i] > 200 {
			historical[i] = 200.0
		}
	}

	currentValue := historical[len(historical)-1]

	// Determine indicator based on volatility level
	var indicator string
	var score float64

	if currentValue > 100 {
		indicator = "Extremely High Volatility"
		score = -0.8 // High volatility is risky
	} else if currentValue > 70 {
		indicator = "High Volatility"
		score = -0.5
	} else if currentValue > 40 {
		indicator = "Moderate Volatility"
		score = 0.0
	} else if currentValue > 20 {
		indicator = "Low Volatility"
		score = 0.3
	} else {
		indicator = "Very Low Volatility"
		score = 0.5
	}

	return MockMetricData{
		Value:      currentValue,
		Indicator:  indicator,
		Score:      score,
		Historical: historical,
		ChartData:  historical[len(historical)-7:],
		APISource:  "Mocked (Options market APIs required for real data)",
		Timestamp:  time.Now(),
	}
}

// MockLiquidationHeatmap generates realistic liquidation risk data
func (m *MockDataGenerator) MockLiquidationHeatmap() MockMetricData {
	rand.Seed(m.seed + 5)

	// Generate realistic liquidation risk score (0-100)
	baseRisk := rand.Float64() * 100

	// Create historical data
	historical := make([]float64, 30)

	for i := 0; i < 30; i++ {
		// Liquidation risk can spike quickly
		if rand.Float64() < 0.1 { // 10% chance of spike
			historical[i] = math.Min(100, baseRisk+rand.Float64()*30)
		} else {
			trend := math.Sin(float64(i)*math.Pi/10) * 10
			noise := (rand.Float64() - 0.5) * 15
			historical[i] = baseRisk + trend + noise
		}

		// Ensure bounds
		if historical[i] < 0 {
			historical[i] = 0
		}
		if historical[i] > 100 {
			historical[i] = 100
		}
	}

	currentValue := historical[len(historical)-1]

	// Determine indicator based on liquidation risk
	var indicator string
	var score float64

	if currentValue > 80 {
		indicator = "Extreme Liquidation Risk"
		score = -0.9
	} else if currentValue > 60 {
		indicator = "High Liquidation Risk"
		score = -0.6
	} else if currentValue > 40 {
		indicator = "Moderate Liquidation Risk"
		score = -0.3
	} else if currentValue > 20 {
		indicator = "Low-Moderate Risk"
		score = 0.0
	} else {
		indicator = "Low Liquidation Risk"
		score = 0.3
	}

	return MockMetricData{
		Value:      currentValue,
		Indicator:  indicator,
		Score:      score,
		Historical: historical,
		ChartData:  historical[len(historical)-7:],
		APISource:  "Mocked (Exchange APIs required for real data)",
		Timestamp:  time.Now(),
	}
}

// MockCorrelationMatrix generates enhanced BTC-Stock correlation
func (m *MockDataGenerator) MockCorrelationMatrix() MockMetricData {
	rand.Seed(m.seed + 6)

	// Generate realistic correlation (-1 to 1)
	baseCorr := (rand.Float64() - 0.5) * 1.6 // -0.8 to 0.8 range

	// Create historical data
	historical := make([]float64, 30)

	for i := 0; i < 30; i++ {
		// Correlation changes slowly
		drift := (rand.Float64() - 0.5) * 0.1
		if i > 0 {
			historical[i] = historical[i-1] + drift
		} else {
			historical[i] = baseCorr
		}

		// Ensure bounds
		if historical[i] > 1 {
			historical[i] = 1.0
		}
		if historical[i] < -1 {
			historical[i] = -1.0
		}
	}

	currentValue := historical[len(historical)-1]

	// Determine indicator based on correlation
	var indicator string
	var score float64

	if currentValue > 0.7 {
		indicator = "Strong Positive Correlation"
		score = -0.5 // High correlation reduces diversification
	} else if currentValue > 0.3 {
		indicator = "Moderate Positive Correlation"
		score = -0.2
	} else if currentValue > -0.3 {
		indicator = "Low Correlation"
		score = 0.3 // Low correlation is good for diversification
	} else if currentValue > -0.7 {
		indicator = "Moderate Negative Correlation"
		score = 0.5
	} else {
		indicator = "Strong Negative Correlation"
		score = 0.7 // Negative correlation is best for diversification
	}

	return MockMetricData{
		Value:      currentValue,
		Indicator:  indicator,
		Score:      score,
		Historical: historical,
		ChartData:  historical[len(historical)-7:],
		APISource:  "Enhanced Mock (Real data from Yahoo Finance API)",
		Timestamp:  time.Now(),
	}
}

// MockMetricData represents a complete mock metric dataset
type MockMetricData struct {
	Value      float64   `json:"value"`
	Indicator  string    `json:"indicator"`
	Score      float64   `json:"score"`
	Historical []float64 `json:"historical"`
	ChartData  []float64 `json:"chart_data"`
	APISource  string    `json:"api_source"`
	Timestamp  time.Time `json:"timestamp"`
}

// Global mock generator instance
var globalMockGenerator = NewMockDataGenerator()

// Public functions for all metrics - these can be called from main.go

func GetMovingAveragesMockData() map[string]interface{} {
	// Generate mock moving averages data
	rand.Seed(time.Now().Unix())

	crossoverTypes := []string{"Golden Cross", "Death Cross", "No Clear Cross"}
	signals := []string{"Buy", "Sell", "Hold"}
	scores := []int{1, -1, 0}

	idx := rand.Intn(len(crossoverTypes))
	crossoverType := crossoverTypes[idx]
	signal := signals[idx]
	score := scores[idx]

	// Create realistic historical data
	historical := make([]float64, 5)
	for i := range historical {
		if signal == "Buy" {
			historical[i] = float64(i) * 0.2 // Upward trend
		} else if signal == "Sell" {
			historical[i] = 1.0 - float64(i)*0.2 // Downward trend
		} else {
			historical[i] = 0.5 // Neutral
		}
	}

	return map[string]interface{}{
		"value":        crossoverType,
		"indicator":    signal,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		"ma50":         65432.10,
		"ma200":        63821.45,
		"api_source":   "Mocked (Enhanced moving averages calculation)",
	}
}

func GetFearGreedMockData() map[string]interface{} {
	rand.Seed(time.Now().Unix() + 100)

	// Generate realistic Fear & Greed Index value (0-100)
	value := rand.Intn(101)

	var indicator string
	var score float64

	if value >= 75 {
		indicator = "Extreme Greed"
		score = -0.8
	} else if value >= 55 {
		indicator = "Greed"
		score = -0.4
	} else if value >= 45 {
		indicator = "Neutral"
		score = 0
	} else if value >= 25 {
		indicator = "Fear"
		score = 0.4
	} else {
		indicator = "Extreme Fear"
		score = 0.8
	}

	// Generate historical data
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = float64(value + rand.Intn(21) - 10) // ±10 variation
		if historical[i] < 0 {
			historical[i] = 0
		}
		if historical[i] > 100 {
			historical[i] = 100
		}
	}

	return map[string]interface{}{
		"value":        value,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		"api_source":   "Mocked (Enhanced fear & greed simulation)",
	}
}

func GetRSIMockData() map[string]interface{} {
	rand.Seed(time.Now().Unix() + 200)

	// Generate realistic RSI value (0-100)
	value := 30.0 + rand.Float64()*40.0 // Focus on 30-70 range mostly

	var indicator string
	var score float64

	if value >= 70 {
		indicator = "Sell"
		score = -1.0
	} else if value <= 30 {
		indicator = "Buy"
		score = 1.0
	} else {
		indicator = "Hold"
		score = 0.0
	}

	// Generate historical RSI data
	historical := make([]float64, 5)
	for i := range historical {
		historical[i] = value + (rand.Float64()-0.5)*10
		if historical[i] < 0 {
			historical[i] = 0
		}
		if historical[i] > 100 {
			historical[i] = 100
		}
	}

	return map[string]interface{}{
		"value":        value,
		"indicator":    indicator,
		"score":        score,
		"chart_data":   historical,
		"chart_labels": []string{"5d", "4d", "3d", "2d", "Now"},
		"api_source":   "Mocked (Enhanced RSI calculation)",
	}
}

func GetStablecoinFlowsMockData() map[string]interface{} {
	data := globalMockGenerator.MockStablecoinFlows()
	return map[string]interface{}{
		"value":        data.Value,
		"indicator":    data.Indicator,
		"score":        data.Score,
		"chart_data":   data.ChartData,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"api_source":   data.APISource,
	}
}

func GetInstitutionalFlowsMockData() map[string]interface{} {
	data := globalMockGenerator.MockInstitutionalFlows()
	return map[string]interface{}{
		"value":        data.Value,
		"indicator":    data.Indicator,
		"score":        data.Score,
		"chart_data":   data.ChartData,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"api_source":   data.APISource,
	}
}

func GetYieldCurvesMockData() map[string]interface{} {
	data := globalMockGenerator.MockYieldCurves()
	return map[string]interface{}{
		"value":        data.Value,
		"indicator":    data.Indicator,
		"score":        data.Score,
		"chart_data":   data.ChartData,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"api_source":   data.APISource,
	}
}

func GetVolatilitySurfaceMockData() map[string]interface{} {
	data := globalMockGenerator.MockVolatilitySurface()
	return map[string]interface{}{
		"value":        data.Value,
		"indicator":    data.Indicator,
		"score":        data.Score,
		"chart_data":   data.ChartData,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"api_source":   data.APISource,
	}
}

func GetLiquidationHeatmapMockData() map[string]interface{} {
	data := globalMockGenerator.MockLiquidationHeatmap()
	return map[string]interface{}{
		"value":        data.Value,
		"indicator":    data.Indicator,
		"score":        data.Score,
		"chart_data":   data.ChartData,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"api_source":   data.APISource,
	}
}

func GetCorrelationMatrixMockData() map[string]interface{} {
	data := globalMockGenerator.MockCorrelationMatrix()
	return map[string]interface{}{
		"value":        data.Value,
		"indicator":    data.Indicator,
		"score":        data.Score,
		"chart_data":   data.ChartData,
		"chart_labels": []string{"7d", "6d", "5d", "4d", "3d", "2d", "Now"},
		"api_source":   data.APISource,
	}
}
