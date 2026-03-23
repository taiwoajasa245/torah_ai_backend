package chat

import "time"

type ChatRequest struct {
	Message   string `json:"message"`
	ChatID    string `json:"chat_id,omitempty"`
	IsNewChat bool   `json:"is_new_chat"`
}

type UpdateChatRequest struct {
	Title string `json:"title"`
}

type Chat struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatMessage struct {
	ID        string    `json:"id"`
	ChatID    string    `json:"chat_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatWithMessages struct {
	Chat     Chat          `json:"chat"`
	Messages []ChatMessage `json:"messages"`
}
