package system

import (
	"net/http"
	"time"

	"go-vue/internal/services"
	"go-vue/pkg/database"
	"go-vue/pkg/market"

	"github.com/gin-gonic/gin"
)

func handleHealth(c *gin.Context) {
	// Check actual telegram service status
	telegramStatus := "not_initialized"
	if services.GlobalRegistry != nil && services.GlobalRegistry.Telegram != nil {
		status := services.GlobalRegistry.Telegram.GetStatus()
		if authenticated, ok := status["authenticated"].(bool); ok && authenticated {
			telegramStatus = "authenticated"
		} else if errorMsg, ok := status["error"].(string); ok && errorMsg != "Telegram service not initialized" {
			telegramStatus = "initialized"
		}
	}

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"services": map[string]interface{}{
			"database": map[string]interface{}{
				"status": "connected",
			},
			"websocket": map[string]interface{}{
				"status": "running",
			},
			"telegram": map[string]interface{}{
				"status": telegramStatus,
			},
		},
	}

	c.JSON(http.StatusOK, health)
}

func handleSystemMetrics(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	// Get API performance stats for major APIs
	apis := []string{"coinmarketcap", "binance", "fear-greed", "alternative.me"}
	apiStats := make(map[string]interface{})

	for _, api := range apis {
		stats, err := db.GetAPIPerformanceStats(api, 24) // Last 24 hours
		if err != nil {
			apiStats[api] = map[string]interface{}{"error": err.Error()}
		} else {
			apiStats[api] = stats
		}
	}

	metrics := map[string]interface{}{
		"timestamp":       time.Now(),
		"api_performance": apiStats,
		"websocket": map[string]interface{}{
			"connected_clients": 0, // Will need to get from hub
			"total_broadcasts":  "N/A",
		},
		"system": map[string]interface{}{
			"uptime": time.Since(time.Now().Add(-time.Hour)), // Placeholder
			"memory": "N/A",
		},
	}

	c.JSON(http.StatusOK, metrics)
}

func handleCacheStats(c *gin.Context) {
	stats := market.GetCacheStats()
	c.JSON(http.StatusOK, stats)
}

func handleCircuitBreakerStats(c *gin.Context) {
	stats := market.GetCircuitBreakerStats()
	c.JSON(http.StatusOK, stats)
}
