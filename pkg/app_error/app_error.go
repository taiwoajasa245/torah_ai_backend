package apperror

import (
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/taiwoajasa245/torah_ai_backend/pkg/pgerrors"
	"github.com/taiwoajasa245/torah_ai_backend/pkg/response"
)

var (
	appErr *AppError
	pgErr  *pgconn.PgError
)

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Handle(w http.ResponseWriter, err error) {
	log.Printf("[ERROR]: %v", err)

	if errors.As(err, &appErr) {
		response.Error(w, appErr.Code, appErr.Message, nil)
		return
	}

	if errors.As(err, &pgErr) {
		pgerrors.HandlePgError(w, pgErr)
		return
	}

	response.Error(w, http.StatusInternalServerError, "Something went wrong, please try again", nil)
}
