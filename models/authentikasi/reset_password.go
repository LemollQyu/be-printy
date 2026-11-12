package models

import (
	"time"

	"github.com/google/uuid"
)

type ForgotPasswordParam struct {
	Email string `json:"email" binding:"required,email,max=100"`
}

type ResendOTPParam struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOTPParam struct {
	Email string `json:"email" binding:"required,email,max=100"`
	OTP   string `json:"otp" binding:"required,numeric,len=5"`
}

type ResetPasswordParam struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required,len=5,numeric"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

type OTP struct {
	ID          uuid.UUID `json:"id" gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID `json:"user_id" gorm:"column:user_id;type:uuid;not null;index:idx_otps_user_id_type,priority:1"`
	Type        string    `json:"type" gorm:"column:type;type:varchar(50);not null;index:idx_otps_user_id_type,priority:2"`
	OTPCodeHash string    `json:"-" gorm:"column:otp_code_hash;type:varchar(255);not null"`
	IsUsed      bool      `json:"is_used" gorm:"column:is_used;not null;default:false"`
	Attempts    int       `json:"attempts" gorm:"column:attempts;not null;default:0"`
	ExpiredAt   time.Time `json:"expired_at" gorm:"column:expired_at;type:timestamptz;not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;type:timestamptz;autoCreateTime"`

	User User `json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

func (OTP) TableName() string {
	return "otps"
}
