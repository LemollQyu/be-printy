package helper

import "errors"

var (
	ErrEmailExist       = errors.New("Email already exist")
	ErrPhoneExist       = errors.New("Phone already exist")
	ErrPasswordRequired = errors.New("Password is required")
	ErrTooManyRequests  = errors.New("Is too many request")
	ErrUserNotFound     = errors.New("User not found")
	ErrUserSigninWrong  = errors.New("Email or password is wrong")
)

var (
	ErrAccountNotVerified = errors.New("User not verified")
	ErrUnauthorized       = errors.New("Unauthorized")
	ErrInvalidToken       = errors.New("Token verify invalid")
	ErrTokenUsed          = errors.New("Token is used")
	ErrTokenExpired       = errors.New("Token is expired")
	ErrInvalidOTP         = errors.New("invalid or expired otp")
	ErrTooManyAttempts    = errors.New("too many otp attempts")
	ErrAccountIsVerified  = errors.New("User is verified")
)
