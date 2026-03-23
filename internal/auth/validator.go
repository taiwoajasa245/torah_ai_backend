package auth

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode"
)

func validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email is required")
	}

	// net/mail is stdlib — properly validates email format
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email address")
	}

	return nil
}

func validatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("password is required")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsNumber(ch):
			hasNumber = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

func validateUserName(username string) error {
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("username is required")
	}

	if len(strings.TrimSpace(username)) <= 5 {
		return fmt.Errorf("username must be more than 5 characters")
	}

	return nil
}
