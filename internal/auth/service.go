package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/taiwoajasa245/torah_ai_backend/internal/mail"
	apperror "github.com/taiwoajasa245/torah_ai_backend/pkg/app_error"
	"github.com/taiwoajasa245/torah_ai_backend/pkg/util"
)

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (*User, error)
	Login(ctx context.Context, email, password string) (*User, error)
	ForgetPassword(ctx context.Context, email string) (bool, error)
	ResetPassword(ctx context.Context, email, otp, newPassword string) (bool, error)
	VerifyOTP(ctx context.Context, email, otp string) (bool, error)
	GetUser(ctx context.Context, userId string) (*User, error)
}

type authService struct {
	repo Repository
	mail *mail.Mailer
}

func NewauthService(repo Repository, mail *mail.Mailer) AuthService {
	return &authService{
		repo: repo,
		mail: mail,
	}
}

func (h *authService) Register(ctx context.Context, username, email, password string) (*User, error) {

	if err := validateEmail(email); err != nil {
		return nil, apperror.New(http.StatusBadRequest, err.Error())
	}

	if err := validateUserName(username); err != nil {
		return nil, apperror.New(http.StatusBadRequest, err.Error())
	}

	if err := validatePassword(password); err != nil {
		return nil, apperror.New(http.StatusBadRequest, err.Error())
	}

	hashed, err := util.HashPasswordBcrypt(password)
	if err != nil {
		return nil, apperror.New(http.StatusInternalServerError, "could not process password")
	}

	user := User{Email: email, Password: hashed, UserName: username}

	_, err = h.repo.CreateUser(ctx, user)
	if err != nil {
		log.Printf("Service err: %v", err.Error())
		return nil, err
	}

	// return usr, nil

	logInUser, err := h.Login(ctx, email, password)
	if err != nil {
		return nil, err
	}

	// data := map[string]interface{}{
	// 	"Name":         user.Email,
	// 	"DashboardURL": "https://nexus/dashboard",
	// }

	// // Send welcome mail asynchronously
	// go func() {
	// 	if err := h.mail.SendHTML(email, "🎉 Welcome to nexus", "welcome.html", data); err != nil {
	// 		log.Printf("failed to send welcome email: %v", err)
	// 	} else {
	// 		log.Println("Email sent successfully")
	// 	}
	// }()

	return logInUser, nil
}

func (h *authService) Login(ctx context.Context, email, password string) (*User, error) {

	if err := validateEmail(email); err != nil {
		return nil, apperror.New(http.StatusBadRequest, err.Error())
	}

	if err := validatePassword(password); err != nil {
		return nil, apperror.New(http.StatusBadRequest, err.Error())
	}

	user, err := h.repo.GetUserByEmail(ctx, email)
	if err != nil {
		log.Printf("Service err: %v", err.Error())
		return nil, err
	}

	err = util.ComparePasswordBcrypt(user.Password, password)
	if err != nil {
		return nil, apperror.New(http.StatusInternalServerError, "incorrect password")
	}

	token, err := util.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	user.Token = token

	return user, nil

}

func (h *authService) ForgetPassword(ctx context.Context, email string) (bool, error) {
	if err := validateEmail(email); err != nil {
		return false, apperror.New(http.StatusBadRequest, err.Error())
	}

	user, err := h.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	// generate OTP
	// 10 minutes expiration
	otp := util.GenerateOTP()
	expiration := time.Now().Add(10 * time.Minute)

	err = h.repo.SavePasswordReset(ctx, email, otp, expiration)
	if err != nil {
		log.Printf("Service err: %v", err.Error())
		return false, ErrInternalServer
	}

	data := map[string]interface{}{
		"Name": user.Email,
		"OTP":  otp,
	}

	go func() {
		h.mail.SendHTML(email, "Reset Your Password OTP", "reset_otp.html", data)
	}()

	return true, nil
}

func (h *authService) VerifyOTP(ctx context.Context, email, otp string) (bool, error) {
	savedOTP, expiresAt, err := h.repo.GetPasswordReset(ctx, email)
	if err != nil {
		return false, errors.New("OTP not found")
	}

	if time.Now().After(expiresAt) {
		return false, errors.New("OTP expired")
	}

	if otp != savedOTP {
		return false, errors.New("invalid OTP")
	}

	return true, nil
}

func (h *authService) ResetPassword(ctx context.Context, email, otp, newPassword string) (bool, error) {

	ok, err := h.VerifyOTP(ctx, email, otp)
	if !ok || err != nil {
		return false, errors.New("invalid or expired OTP")
	}

	hashed, err := util.HashPasswordBcrypt(newPassword)
	if err != nil {
		return false, err
	}

	// update password
	err = h.repo.UpdateUserPassword(ctx, email, hashed)
	if err != nil {
		return false, err
	}

	// delete OTP in DB
	if err = h.repo.DeletePasswordReset(ctx, email); err != nil {
		log.Printf("failed to delete used OTP: %v", err)
		return false, err
	}

	return true, nil
}

func (h *authService) GetUser(ctx context.Context, userId string) (*User, error) {

	usr, err := h.repo.GetUserByID(ctx, userId)
	if err != nil {
		log.Printf("Service err: %v", err.Error())
		return nil, err
	}

	return usr, nil

}
