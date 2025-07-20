package telegram

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	telegram := api.Group("/telegram")
	{
		telegram.GET("/auth/callback", handleTelegramAuthCallback)
		telegram.GET("/phone", handleTelegramPhone)
		telegram.POST("/verify-code", handleTelegramVerifyCode)
		telegram.POST("/verify-2fa", handleTelegramVerify2FA)
		telegram.GET("/status", handleTelegramStatus)
		telegram.POST("/logout", handleTelegramLogout)
		telegram.GET("/groups", handleGetGroups)
		telegram.GET("/current-user", handleGetCurrentUser)
		telegram.GET("/signal-channels", handleTelegramSignalChannels)
		telegram.GET("/analyze-token", handleTokenAnalysis)
		telegram.POST("/analyze-group-tokens", handleBulkTokenAnalysis)
	}
}
