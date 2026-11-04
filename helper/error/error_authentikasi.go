package helper

import "errors"

var (
	ErrEmailExist       = errors.New("Email already exist")
	ErrPhoneExist       = errors.New("Phone already exist")
	ErrPasswordRequired = errors.New("password is required")
)
