package controllers

import (
	"encoding/json"
	"net/http"

	"go-vue/internal/models"
	"go-vue/pkg/telegram"
)

type TelegramController struct {
	telegramService *telegram.TelegramService
}

func NewTelegramController(telegramService *telegram.TelegramService) *TelegramController {
	return &TelegramController{
		telegramService: telegramService,
	}
}

func (c *TelegramController) GetAuthLink(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	authLink := c.telegramService.GenerateAuthLink()
	response := models.TelegramAuthResponse{
		AuthLink: authLink,
	}

	json.NewEncoder(w).Encode(response)
}

func (c *TelegramController) GetUserGroups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// In a real application, you would get the user ID from the session/token
	userID := int64(123456) // Placeholder user ID

	groups, err := c.telegramService.GetUserGroups(userID)
	if err != nil {
		response := models.TelegramGroupsResponse{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert Telegram API groups to our model
	var telegramGroups []models.TelegramGroup
	for _, group := range groups {
		telegramGroups = append(telegramGroups, models.TelegramGroup{
			ID:          group.ID,
			Title:       group.Title,
			Username:    group.UserName,
			Type:        group.Type,
			Description: group.Description,
		})
	}

	response := models.TelegramGroupsResponse{
		Groups: telegramGroups,
	}

	json.NewEncoder(w).Encode(response)
}
