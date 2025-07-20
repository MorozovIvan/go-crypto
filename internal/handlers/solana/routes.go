package solana

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	solana := api.Group("/solana")
	{
		solana.GET("/balance/:address", getSolanaBalance)
		// Add more Solana-specific endpoints here
	}
}
