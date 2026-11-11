package helper

import "errors"

var (
	ErrEmailExist       = errors.New("Email already exist")
	ErrPhoneExist       = errors.New("Phone already exist")
	ErrPasswordRequired = errors.New("Password is required")
	ErrTooManyRequests  = errors.New("Is too many request")
	ErrUserNotFound     = errors.New("User not found")
)

var (
	ErrInvalidToken      = errors.New("Token verify invalid")
	ErrTokenUsed         = errors.New("Token is used")
	ErrTokenExpired      = errors.New("Token is expired")
	ErrInvalidOTP        = errors.New("invalid or expired otp")
	ErrTooManyAttempts   = errors.New("too many otp attempts")
	ErrAccountIsVerified = errors.New("User is verified")
)
