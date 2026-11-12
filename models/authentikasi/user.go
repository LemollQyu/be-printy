package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleCustomer UserRole = "customer"
)

type User struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FullName        string    `json:"full_name" gorm:"column:full_name;type:varchar(100);not null"`
	Email           string    `json:"email" gorm:"column:email;type:varchar(100);uniqueIndex;not null"`
	Phone           *string   `json:"phone" gorm:"column:phone;type:varchar(100)"`
	AvatarURL       *string   `json:"avatar_url" gorm:"column:avatar_url;type:text"`
	PasswordHash    *string   `json:"-" gorm:"column:password_hash;type:varchar(255)"`
	Role            UserRole  `json:"role" gorm:"column:role;type:user_role;not null;default:customer"`
	IsEmailVerified bool      `json:"is_email_verified" gorm:"column:is_email_verified;not null;default:false"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type EmailVerificationToken struct {
	ID        uuid.UUID  `json:"id" gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID  `json:"user_id" gorm:"column:user_id;type:uuid;not null;index:idx_email_verification_tokens_user_id"`
	TokenHash string     `json:"-" gorm:"column:token_hash;type:varchar(255);unique;not null"`
	UsedAt    *time.Time `json:"used_at" gorm:"column:used_at;type:timestamptz"`
	ExpiredAt time.Time  `json:"expired_at" gorm:"column:expired_at;type:timestamptz;not null"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;type:timestamptz;autoCreateTime"`

	User User `json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

func (EmailVerificationToken) TableName() string {
	return "email_verification_tokens"
}

func (User) TableName() string {
	return "users"
}

type ResendVerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type RegisterParameter struct {
	FullName        string `json:"full_name" binding:"required,min=3,max=100"`
	Email           string `json:"email" binding:"required,email,max=100"`
	Phone           string `json:"phone" binding:"required,numeric,min=9,max=15"`
	Password        string `json:"password" binding:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}
