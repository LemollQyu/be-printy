package repository

import (
	"context"
	"errors"
	"time"

	models "github.com/LemollQyu/be-printy/models/authentikasi"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// repository insert otp
func (r *PrintyRepository) InsertOtp(ctx context.Context, otp models.OTP) error {
	if err := r.Database.WithContext(ctx).Omit("User").Create(&otp).Error; err != nil {
		return err
	}

	return nil
}

// tandai semua OTP milik user (dengan type tertentu) yang belum dipakai sebagai terpakai
func (r *PrintyRepository) InvalidateOtps(ctx context.Context, userID uuid.UUID, otpType string) error {
	return r.Database.WithContext(ctx).
		Model(&models.OTP{}).
		Where("user_id = ? AND type = ? AND is_used = ?", userID, otpType, false).
		Update("is_used", true).Error
}

// ambil OTP terbaru milik user berdasarkan type
// return (nil, nil) kalau belum ada
func (r *PrintyRepository) GetLatestOtp(ctx context.Context, userID uuid.UUID, otpType string) (*models.OTP, error) {
	var otp models.OTP

	err := r.Database.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, otpType).
		Order("created_at DESC").
		First(&otp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &otp, nil
}

// tambah counter percobaan salah (atomik di level DB)
func (r *PrintyRepository) IncrementOtpAttempts(ctx context.Context, otpID uuid.UUID) error {
	return r.Database.WithContext(ctx).
		Model(&models.OTP{}).
		Where("id = ?", otpID).
		UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error
}

// tandai OTP sudah dipakai secara atomik
// return true kalau berhasil di-update, false kalau sudah keduluan dipakai request lain
func (r *PrintyRepository) MarkOtpUsed(ctx context.Context, otpID uuid.UUID) (bool, error) {
	res := r.Database.WithContext(ctx).
		Model(&models.OTP{}).
		Where("id = ? AND is_used = ?", otpID, false).
		Update("is_used", true)
	if res.Error != nil {
		return false, res.Error
	}

	return res.RowsAffected > 0, nil
}

// update password user
// nama tabel/kolom disesuaikan dengan model User-mu
func (r *PrintyRepository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	return r.Database.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("password_hash", hashedPassword).Error
}

// hitung jumlah OTP yang dibuat sejak waktu tertentu
func (r *PrintyRepository) CountOtpsSince(ctx context.Context, userID uuid.UUID, otpType string, since time.Time) (int64, error) {
	var count int64

	err := r.Database.WithContext(ctx).
		Model(&models.OTP{}).
		Where("user_id = ? AND type = ? AND created_at >= ?", userID, otpType, since).
		Count(&count).Error

	return count, err
}
