package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/taiwoajasa245/torah_ai_backend/internal/middleware"
	apperror "github.com/taiwoajasa245/torah_ai_backend/pkg/app_error"
	"github.com/taiwoajasa245/torah_ai_backend/pkg/request"
	"github.com/taiwoajasa245/torah_ai_backend/pkg/response"
)

type ChatHandler struct {
	service ChatService
}

func NewHandler(service ChatService) *ChatHandler {
	return &ChatHandler{service: service}
}

// ChatHandler godoc
// @Summary Chat with AI
// @Description Start a chat with AI and save the conversation. Streams response via SSE.
// @Tags Chat
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   request body ChatRequest true "Chat request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /chat [post]
func (h *ChatHandler) ChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		apperror.Handle(w, apperror.New(http.StatusUnauthorized, "user not logged in"))
		return
	}

	var req ChatRequest
	if !request.DecodeJson(w, r, &req) {
		return
	}

	if req.Message == "" {
		response.Error(w, http.StatusBadRequest, "message is required", nil)
		return
	}

	if !req.IsNewChat && req.ChatID == "" {
		response.Error(w, http.StatusBadRequest, "chat_id is required when is_new_chat is false", nil)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "Streaming not supported", nil)
		return
	}

	err := h.service.StreamChat(r.Context(), userID, req.Message, req.ChatID, req.IsNewChat,
		func(chatID string) error {
			chatIDEvent, _ := json.Marshal(map[string]string{"chat_id": chatID})
			fmt.Fprintf(w, "data: %s\n\n", chatIDEvent)
			flusher.Flush()
			return nil
		},
		func(token string) error {
			for _, line := range strings.Split(token, "\n") {
				fmt.Fprintf(w, "data: %s\n", line)
			}
			fmt.Fprint(w, "\n")
			flusher.Flush()
			return nil
		},
	)

	if err != nil {
		fmt.Fprintf(w, "data: {\"error\":%q}\n\n", err.Error())
		flusher.Flush()
		return
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// GetAllChatsHandler godoc
// @Summary Get all chats
// @Description Retrieve all chats for the authenticated user
// @Tags Chat
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /chat [get]
func (h *ChatHandler) GetAllChatsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		apperror.Handle(w, apperror.New(http.StatusUnauthorized, "user not logged in"))
		return
	}

	chats, err := h.service.GetAllChats(r.Context(), userID)
	if err != nil {
		apperror.Handle(w, apperror.New(http.StatusInternalServerError, err.Error()))
		return
	}

	response.Success(w, chats, "Chats retrieved successfully")
}

// GetChatByIDHandler godoc
// @Summary Get chat by ID
// @Description Retrieve a specific chat with its messages
// @Tags Chat
// @Produce  json
// @Security BearerAuth
// @Param   id   path string true "Chat ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /chat/{id} [get]
func (h *ChatHandler) GetChatByIDHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		apperror.Handle(w, apperror.New(http.StatusUnauthorized, "user not logged in"))
		return
	}

	chatID := chi.URLParam(r, "id")
	if chatID == "" {
		response.Error(w, http.StatusBadRequest, "chat id is required", nil)
		return
	}

	chatWithMessages, err := h.service.GetChatByID(r.Context(), chatID, userID)
	if err != nil {
		if errors.Is(err, ErrChatNotFound) {
			apperror.Handle(w, apperror.New(http.StatusNotFound, "chat not found"))
			return
		}
		apperror.Handle(w, apperror.New(http.StatusInternalServerError, err.Error()))
		return
	}

	response.Success(w, chatWithMessages, "Chat retrieved successfully")
}

// UpdateChatHandler godoc
// @Summary Rename chat
// @Description Update the title of a chat
// @Tags Chat
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   id   path string true "Chat ID"
// @Param   request body UpdateChatRequest true "Update chat request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /chat/{id} [patch]
func (h *ChatHandler) UpdateChatHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		apperror.Handle(w, apperror.New(http.StatusUnauthorized, "user not logged in"))
		return
	}

	chatID := chi.URLParam(r, "id")
	if chatID == "" {
		response.Error(w, http.StatusBadRequest, "chat id is required", nil)
		return
	}

	var req UpdateChatRequest
	if !request.DecodeJson(w, r, &req) {
		return
	}

	if req.Title == "" {
		response.Error(w, http.StatusBadRequest, "title is required", nil)
		return
	}

	err := h.service.UpdateChatTitle(r.Context(), chatID, userID, req.Title)
	if err != nil {
		if errors.Is(err, ErrChatNotFound) {
			apperror.Handle(w, apperror.New(http.StatusNotFound, "chat not found"))
			return
		}
		apperror.Handle(w, apperror.New(http.StatusInternalServerError, err.Error()))
		return
	}

	response.Success(w, map[string]string{"message": "Chat renamed successfully"}, "Chat updated successfully")
}

// DeleteChatHandler godoc
// @Summary Delete chat
// @Description Delete a chat and all its messages
// @Tags Chat
// @Produce  json
// @Security BearerAuth
// @Param   id   path string true "Chat ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /chat/{id} [delete]
func (h *ChatHandler) DeleteChatHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		apperror.Handle(w, apperror.New(http.StatusUnauthorized, "user not logged in"))
		return
	}

	chatID := chi.URLParam(r, "id")
	if chatID == "" {
		response.Error(w, http.StatusBadRequest, "chat id is required", nil)
		return
	}

	err := h.service.DeleteChat(r.Context(), chatID, userID)
	if err != nil {
		if errors.Is(err, ErrChatNotFound) {
			apperror.Handle(w, apperror.New(http.StatusNotFound, "chat not found"))
			return
		}
		apperror.Handle(w, apperror.New(http.StatusInternalServerError, err.Error()))
		return
	}

	response.Success(w, map[string]string{"message": "Chat deleted successfully"}, "Chat deleted successfully")
}
