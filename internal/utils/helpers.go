package utils

import (
	"github.com/gin-gonic/gin"
)

// shouldUseMockData checks if mock data is requested
func ShouldUseMockData(c *gin.Context) bool {
	return c.Query("mock") == "true"
}

// getRSIScore calculates RSI score based on value
func GetRSIScore(value float64) float64 {
	// RSI: 0-100 scale
	// 0-30: Oversold (bullish signal)
	// 30-45: Bearish momentum weakening
	// 45-55: Neutral
	// 55-70: Bullish momentum
	// 70-100: Overbought (bearish signal)

	if value <= 30 {
		// Oversold - bullish signal
		return 0.4 + (30-value)/30*0.6 // 0.4 to 1.0
	} else if value <= 45 {
		// Bearish momentum weakening
		return (45 - value) / 15 * 0.4 // 0 to 0.4
	} else if value <= 55 {
		// Neutral zone
		return (value - 50) / 5 * 0.2 // -0.2 to 0.2
	} else if value <= 70 {
		// Bullish momentum
		return 0.2 + (value-55)/15*0.3 // 0.2 to 0.5
	} else {
		// Overbought - bearish signal
		return -0.3 - (value-70)/30*0.7 // -0.3 to -1.0
	}
}

// getFearGreedScore calculates Fear & Greed Index score based on value
func GetFearGreedScore(value float64) float64 {
	// Fear & Greed Index: 0-100 scale
	// 0-25: Extreme Fear (bullish contrarian signal)
	// 25-45: Fear (moderately bullish)
	// 45-55: Neutral
	// 55-75: Greed (moderately bearish)
	// 75-100: Extreme Greed (bearish contrarian signal)

	if value <= 25 {
		// Extreme fear - strong buy signal
		return 0.6 + (25-value)/25*0.4 // 0.6 to 1.0
	} else if value <= 45 {
		// Fear - moderate buy signal
		return 0.2 + (45-value)/20*0.4 // 0.2 to 0.6
	} else if value <= 55 {
		// Neutral zone
		return (50 - value) / 5 * 0.2 // -0.2 to 0.2
	} else if value <= 75 {
		// Greed - moderate sell signal
		return -0.2 - (value-55)/20*0.4 // -0.2 to -0.6
	} else {
		// Extreme greed - strong sell signal
		return -0.6 - (value-75)/25*0.4 // -0.6 to -1.0
	}
}

// getBTCDominanceScore calculates BTC Dominance score based on value
func GetBTCDominanceScore(value float64) float64 {
	// BTC Dominance optimal range is typically 40-60%
	// Below 40%: Altcoin season (negative for BTC, positive for alts)
	// Above 60%: BTC dominance (positive for BTC, negative for alts)
	// We'll score from altcoin perspective (lower dominance = positive)

	if value <= 40 {
		// Strong altcoin season
		return 0.5 + (40-value)/40*0.5 // 0.5 to 1.0
	} else if value <= 50 {
		// Moderate altcoin favorability
		return (50 - value) / 10 * 0.5 // 0 to 0.5
	} else if value <= 60 {
		// BTC starting to dominate
		return -(value - 50) / 10 * 0.5 // 0 to -0.5
	} else {
		// Strong BTC dominance
		return -0.5 - (value-60)/40*0.5 // -0.5 to -1.0
	}
}

// getIndicator returns indicator based on value and bounds
func GetIndicator(value float64, lowerBound, upperBound float64) string {
	if value < lowerBound {
		return "Sell"
	} else if value > upperBound {
		return "Buy"
	}
	return "Hold"
}

// getScore returns score based on value and bounds
func GetScore(value float64, lowerBound, upperBound float64) float64 {
	if value < lowerBound {
		return -1
	} else if value > upperBound {
		return 1
	}
	return 0
}

// getWhaleTransactionScore calculates whale transaction score based on count
func GetWhaleTransactionScore(count float64) float64 {
	// Whale transaction count analysis
	// Very high activity: Uncertainty/volatility (slightly bearish)
	// Moderate activity: Normal market (neutral)
	// Low activity: Accumulation phase (bullish)
	// Baseline: ~100-300 transactions per day

	if count <= 100 {
		// Low whale activity - potential accumulation
		return 0.2 + (100-count)/100*0.3 // 0.2 to 0.5
	} else if count <= 200 {
		// Normal whale activity
		return 0.2 - (count-100)/100*0.4 // 0.2 to -0.2
	} else if count <= 400 {
		// High whale activity - market uncertainty
		return -0.2 - (count-200)/200*0.3 // -0.2 to -0.5
	} else {
		// Very high whale activity - potential distribution
		return -0.5 - ((count - 400) / 400 * 0.5) // -0.5 to -1.0 (simplified math.Min)
	}
}
