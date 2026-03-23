package auth

import (
	"net/http"

	"github.com/taiwoajasa245/torah_ai_backend/internal/middleware"
	apperror "github.com/taiwoajasa245/torah_ai_backend/pkg/app_error"
	"github.com/taiwoajasa245/torah_ai_backend/pkg/request"
	"github.com/taiwoajasa245/torah_ai_backend/pkg/response"
)

type AuthHandler struct {
	service AuthService
}

func NewHandler(service AuthService) AuthHandler {
	return AuthHandler{service: service}
}

// RegisterHandler godoc
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param   request body RegisterRequest true "Register user request"
// @Success 201 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/register-with-email [post]
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if !request.DecodeJson(w, r, &req) {
		return
	}

	user := User{Email: req.Email, Password: req.Password, UserName: req.UserName}

	usr, err := h.service.Register(r.Context(), user.UserName, user.Email, user.Password)
	if err != nil {
		apperror.Handle(w, err)
		return
	}

	response.Success(w, usr, "User registered successfully")
}

// LoginHandler godoc
// @Summary Login a user
// @Description Authenticate a user and return a JWT token
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param   request body LoginRequest true "Login user request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !request.DecodeJson(w, r, &req) {
		return
	}

	user := &User{Email: req.Email, Password: req.Password}

	user, err := h.service.Login(r.Context(), user.Email, user.Password)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	response.Success(w, &user, "Ok")
}

// ForgetPasswordHandler godoc
// @Summary Initiate password reset
// @Description Send OTP to user's email for password reset
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param   request body ForgetPasswordRequest true "Forget password request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /auth/forget-password [post]
func (h *AuthHandler) ForgetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req ForgetPasswordRequest
	if !request.DecodeJson(w, r, &req) {
		return
	}

	if req.Email == "" {
		response.Error(w, http.StatusBadRequest, "Missing required fields", map[string]string{
			"email": "Email is required",
		})
		return
	}

	success, err := h.service.ForgetPassword(r.Context(), req.Email)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to process request", err.Error())
		return
	}

	response.Success(w, success, "OTP sent to email successfully")

}

// ResetPasswordHandler godoc
// @Summary Reset user password
// @Description Reset password using OTP sent to email
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param   request body ResetPasswordRequest true "Reset password request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if !request.DecodeJson(w, r, &req) {
		return
	}

	if req.NewPassword == "" || req.OTP == "" || req.Email == "" {
		response.Error(w, http.StatusBadRequest, "Missing required fields", map[string]string{
			"new_password": "New Password is required",
			"otp":          "OTP is required",
			"email":        "Email is required",
		})
		return
	}

	success, err := h.service.ResetPassword(r.Context(), req.Email, req.OTP, req.NewPassword)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to reset password", err.Error())
		return
	}

	response.Success(w, success, "Password reset successfully")
}

// GetUserDetailsHandler godoc
// @Summary Get user details
// @Description Retrieve detailed information about the authenticated user
// @Tags Auth
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) GetUserDetailsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		apperror.Handle(w, apperror.New(http.StatusUnauthorized, "user not found"))
		return
	}

	userDetails, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		apperror.Handle(w, err)
		return
	}

	if token, ok := middleware.GetUserToken(r); ok {
		userDetails.Token = token
	}

	response.Success(w, userDetails, "User Profile Retrieved Successfully")

}
