package chat

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	apperror "github.com/taiwoajasa245/torah_ai_backend/pkg/app_error"
	"google.golang.org/genai"
)

type ChatService interface {
	StreamChat(ctx context.Context, userID, message string, chatID string, isNewChat bool, onChatCreated func(chatID string) error, onToken func(token string) error) error
	GetAllChats(ctx context.Context, userID string) ([]Chat, error)
	GetChatByID(ctx context.Context, chatID, userID string) (*ChatWithMessages, error)
	UpdateChatTitle(ctx context.Context, chatID, userID, title string) error
	DeleteChat(ctx context.Context, chatID, userID string) error
}

type chatService struct {
	client *genai.Client
	model  string
	repo   Repository
}

func NewChatService(apiKey, model string, repo Repository) (ChatService, error) {
	if model == "" {
		model = "gemini-2.5-flash-lite"
	}

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}
	return &chatService{client: client, model: model, repo: repo}, nil
}

func (s *chatService) buildContents(ctx context.Context, chatID string) ([]*genai.Content, error) {
	messages, err := s.repo.GetChatMessages(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get chat messages: %w", err)
	}

	contents := make([]*genai.Content, 0, len(messages))
	for _, m := range messages {
		role := "user"
		if m.Role == "model" {
			role = "model"
		}
		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{{Text: m.Content}},
		})
	}
	return contents, nil
}

func truncateTitle(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func (s *chatService) StreamChat(ctx context.Context, userID, message string, chatID string, isNewChat bool, onChatCreated func(chatID string) error, onToken func(token string) error) error {
	var activeChatID string

	if isNewChat || chatID == "" {
		// Create new chat
		title := truncateTitle(message, 50)
		if title == "" {
			title = "New Chat"
		}
		chat, err := s.repo.CreateChat(ctx, userID, title)
		if err != nil {
			return fmt.Errorf("create chat: %w", err)
		}
		activeChatID = chat.ID
		if onChatCreated != nil {
			if err := onChatCreated(chat.ID); err != nil {
				return err
			}
		}
	} else {
		// Continue existing chat - verify user owns it
		chat, err := s.repo.GetChatByID(ctx, chatID, userID)
		if err != nil {
			return fmt.Errorf("chat not found or access denied: %w", err)
		}
		activeChatID = chat.ID
	}

	if err := s.repo.CreateChatMessage(ctx, activeChatID, "user", message); err != nil {
		return fmt.Errorf("save user message: %w", err)
	}

	var fullResponse strings.Builder
	systemPrompt := `You are TorahAI, a wise and knowledgeable Bible scholar. 
	Always respond using Markdown formatting:
	- Use **bold** for key terms and important concepts
	- Use ## for section headings
	- Use bullet points for lists
	- Use > for quoting scripture verses
	- Use *italic* for Hebrew words
	- Always end responses with a cited verse like: 📖 *Genesis 1:1*
	Keep responses clear, structured, and spiritually insightful.`

	// Build contents: includes full conversation history (user message already saved above)
	contents, err := s.buildContents(ctx, activeChatID)
	if err != nil {
		return err
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
		Tools: []*genai.Tool{
			{GoogleSearch: &genai.GoogleSearch{}},
		},
	}

	for resp, err := range s.client.Models.GenerateContentStream(ctx, s.model, contents, config) {
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}
		if resp != nil && resp.Text() != "" {
			fullResponse.WriteString(resp.Text())
			if onToken != nil {
				if err := onToken(resp.Text()); err != nil {
					return err
				}
			}
		}
	}

	if err := s.repo.CreateChatMessage(ctx, activeChatID, "model", fullResponse.String()); err != nil {
		return fmt.Errorf("save model message: %w", err)
	}

	return nil
}

func (s *chatService) GetAllChats(ctx context.Context, userID string) ([]Chat, error) {

	allchats, err := s.repo.GetAllChatsByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.New(http.StatusBadRequest, err.Error())
	}

	return allchats, nil
}

func (s *chatService) GetChatByID(ctx context.Context, chatID, userID string) (*ChatWithMessages, error) {
	chat, err := s.repo.GetChatByID(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	messages, err := s.repo.GetChatMessages(ctx, chatID)
	if err != nil {
		return nil, err
	}
	return &ChatWithMessages{Chat: *chat, Messages: messages}, nil
}

func (s *chatService) UpdateChatTitle(ctx context.Context, chatID, userID, title string) error {
	if _, err := s.repo.GetChatByID(ctx, chatID, userID); err != nil {
		return err
	}
	return s.repo.UpdateChatTitle(ctx, chatID, userID, title)
}

func (s *chatService) DeleteChat(ctx context.Context, chatID, userID string) error {
	if _, err := s.repo.GetChatByID(ctx, chatID, userID); err != nil {
		return err
	}
	return s.repo.DeleteChat(ctx, chatID, userID)
}
