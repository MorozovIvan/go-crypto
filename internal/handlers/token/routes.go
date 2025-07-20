package token

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	token := api.Group("/token")
	{
		token.GET("/analysis", handleTokenAnalysisWithData)
		// Add more token-specific endpoints here
	}
}
