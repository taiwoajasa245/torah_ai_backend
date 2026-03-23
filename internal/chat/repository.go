package chat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	db "github.com/taiwoajasa245/torah_ai_backend/db/sqlc"
)

var (
	ErrChatNotFound = errors.New("chat not found")
)

type Repository interface {
	CreateChat(ctx context.Context, userID, title string) (*Chat, error)
	GetChatByID(ctx context.Context, chatID, userID string) (*Chat, error)
	GetAllChatsByUserID(ctx context.Context, userID string) ([]Chat, error)
	UpdateChatTitle(ctx context.Context, chatID, userID, title string) error
	DeleteChat(ctx context.Context, chatID, userID string) error
	CreateChatMessage(ctx context.Context, chatID, role, content string) error
	GetChatMessages(ctx context.Context, chatID string) ([]ChatMessage, error)
}

type repository struct {
	queries *db.Queries
}

func NewRepository(database *sql.DB) Repository {
	return &repository{queries: db.New(database)}
}

func toChat(c db.Chat) Chat {
	return Chat{
		ID:        c.ID,
		UserID:    c.UserID,
		Title:     c.Title,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func toChatMessage(m db.ChatMessage) ChatMessage {
	return ChatMessage{
		ID:        m.ID,
		ChatID:    m.ChatID,
		Role:      m.Role,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

func (r *repository) CreateChat(ctx context.Context, userID, title string) (*Chat, error) {
	if title == "" {
		title = "New Chat"
	}
	c, err := r.queries.CreateChat(ctx, db.CreateChatParams{UserID: userID, Title: title})
	if err != nil {
		return nil, fmt.Errorf("CreateChat: %w", err)
	}
	chat := toChat(c)
	return &chat, nil
}

func (r *repository) GetChatByID(ctx context.Context, chatID, userID string) (*Chat, error) {
	c, err := r.queries.GetChatByID(ctx, db.GetChatByIDParams{ID: chatID, UserID: userID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChatNotFound
		}
		return nil, fmt.Errorf("GetChatByID: %w", err)
	}
	chat := toChat(c)
	return &chat, nil
}

func (r *repository) GetAllChatsByUserID(ctx context.Context, userID string) ([]Chat, error) {
	rows, err := r.queries.GetAllChatsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetAllChatsByUserID: %w", err)
	}
	chats := make([]Chat, len(rows))
	for i, c := range rows {
		chats[i] = toChat(c)
	}
	return chats, nil
}

func (r *repository) UpdateChatTitle(ctx context.Context, chatID, userID, title string) error {
	err := r.queries.UpdateChatTitle(ctx, db.UpdateChatTitleParams{
		Title: title, ID: chatID, UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("UpdateChatTitle: %w", err)
	}
	return nil
}

func (r *repository) DeleteChat(ctx context.Context, chatID, userID string) error {
	err := r.queries.DeleteChat(ctx, db.DeleteChatParams{ID: chatID, UserID: userID})
	if err != nil {
		return fmt.Errorf("DeleteChat: %w", err)
	}
	return nil
}

func (r *repository) CreateChatMessage(ctx context.Context, chatID, role, content string) error {
	_, err := r.queries.CreateChatMessage(ctx, db.CreateChatMessageParams{
		ChatID: chatID, Role: role, Content: content,
	})
	return err
}

func (r *repository) GetChatMessages(ctx context.Context, chatID string) ([]ChatMessage, error) {
	rows, err := r.queries.GetChatMessages(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("GetChatMessages: %w", err)
	}
	msgs := make([]ChatMessage, len(rows))
	for i, m := range rows {
		msgs[i] = toChatMessage(m)
	}
	return msgs, nil
}
