package market

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	// Market data endpoints
	api.GET("/cmc/global", handleCMCGlobal)
	api.GET("/trends", handleTrends)
	api.GET("/ssr", handleSSR)
	api.GET("/exchange-flows", handleExchangeFlows)
	api.GET("/active-addresses", handleActiveAddresses)
	api.GET("/whale-transactions", handleWhaleTransactions)
	api.GET("/funding-rate", handleFundingRate)
	api.GET("/open-interest", handleOpenInterest)

	// Market metrics endpoints
	api.GET("/altcoin-season", handleAltcoinSeasonIndex)
	api.GET("/volume-trend", handleVolumeTrend)
	api.GET("/bollinger-bands", handleBollingerBands)
	api.GET("/rsi", handleRSI)
	api.GET("/moving-averages", handleMovingAverages)

	// Additional market endpoints
	api.GET("/fear-greed", handleFearGreed)
	api.GET("/btc-dominance", handleBTCDominance)
	api.GET("/market-cap", handleMarketCap)
	api.GET("/eth-btc-ratio", handleETHBTCRatio)
	api.GET("/liquidation", handleLiquidation)
	api.GET("/google-trends", handleGoogleTrends)

	// Advanced DeFi & Crypto Metrics
	api.GET("/defi-tvl", handleDeFiTVL)
	api.GET("/social-sentiment", handleSocialSentiment)
	api.GET("/options-flow", handleOptionsFlow)
	api.GET("/stablecoin-flows", handleStablecoinFlows)
	api.GET("/network-health", handleNetworkHealth)
	api.GET("/institutional-flows", handleInstitutionalFlows)
	api.GET("/yield-curves", handleYieldCurves)
	api.GET("/correlation-matrix", handleCorrelationMatrix)
	api.GET("/volatility-surface", handleVolatilitySurface)
	api.GET("/liquidation-heatmap", handleLiquidationHeatmap)
}
