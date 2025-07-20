package portfolio

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	api.GET("/portfolio", handlePortfolio)
	api.GET("/balance", handleBalance)
	// Add more portfolio-specific endpoints here
}
