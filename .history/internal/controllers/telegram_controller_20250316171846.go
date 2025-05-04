package controllers

import (
	"encoding/json"
	"net/http"

	"go-vue/pkg/telegram"
)

type TelegramController struct {
	telegramService *telegram.TelegramService
}

func NewTelegramController(service *telegram.TelegramService) *TelegramController {
	return &TelegramController{
		telegramService: service,
	}
}

func (c *TelegramController) GetAuthLink(w http.ResponseWriter, r *http.Request) {
	authLink := c.telegramService.GenerateAuthLink()
	response := map[string]string{
		"auth_link": authLink,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *TelegramController) GetUserGroups(w http.ResponseWriter, r *http.Request) {
	// Get user ID from session or database
	// For now, we'll use a mock user ID
	userID := int64(123456789)

	groups, err := c.telegramService.GetUserGroups(userID)
	if err != nil {
		http.Error(w, "Failed to get user groups", http.StatusInternalServerError)
		return
	}

	// Convert map[string]interface{} to proper response format
	type GroupResponse struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Username    string `json:"username"`
		Type        string `json:"type"`
		Description string `json:"description"`
	}

	response := make([]GroupResponse, len(groups))
	for i, group := range groups {
		response[i] = GroupResponse{
			ID:          group["id"].(string),
			Title:       group["title"].(string),
			Username:    group["username"].(string),
			Type:        group["type"].(string),
			Description: group["description"].(string),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
