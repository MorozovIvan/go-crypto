package models

type TelegramUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TelegramGroup struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Username    string `json:"username"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type TelegramAuthResponse struct {
	AuthLink string `json:"auth_link"`
	Error    string `json:"error,omitempty"`
}

type TelegramGroupsResponse struct {
	Groups []TelegramGroup `json:"groups"`
	Error  string          `json:"error,omitempty"`
}
