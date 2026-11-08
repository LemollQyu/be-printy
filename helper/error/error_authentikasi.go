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
	ErrInvalidToken       = errors.New("Token verify invalid")
	ErrTokenUsed          = errors.New("Token is used")
	ErrTokenExpired       = errors.New("Token is expired")
	ErrAccountIsVerified  = errors.New("User is verified")
	ErrAccountNotVerified = errors.New("User not verified")
	ErrUnauthorized       = errors.New("Unauthorized")
)
