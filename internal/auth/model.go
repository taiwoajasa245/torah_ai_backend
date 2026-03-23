// User model definition
package auth

import "time"

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	UserName string `json:"username"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgetPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

type User struct {
	ID                 string    `json:"id"`
	UserName           string    `json:"username,omitempty"`
	Email              string    `json:"email"`
	Password           string    `json:"-"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	Token              string    `json:"token,omitempty"`
	IsProfileCompleted bool      `json:"is_profile_completed,omitempty"`
}
