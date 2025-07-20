package system

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	system := api.Group("/system")
	{
		system.GET("/health", handleHealth)
		system.GET("/metrics", handleSystemMetrics)
		system.GET("/cache-stats", handleCacheStats)
		system.GET("/circuit-breaker-stats", handleCircuitBreakerStats)
	}
}
