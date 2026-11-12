package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/LemollQyu/be-printy/helper"
	herr "github.com/LemollQyu/be-printy/helper/error"
	"github.com/LemollQyu/be-printy/infrastructure/log"
	models "github.com/LemollQyu/be-printy/models/authentikasi"
	"github.com/LemollQyu/be-printy/utils"
	"github.com/sirupsen/logrus"
)

const (
	forgotPasswordOTPTTL  = 5 * time.Minute
	forgotPasswordOTPType = "forgot_password"
	otpResendCooldown     = 60 * time.Second
	maxOTPPerHour         = 5
)

// logika bersama untuk ForgotPassword dan ResendOTP
func (uc *PrintyUsecase) sendForgotPasswordOTP(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	// 1. cari user
	user, err := uc.PrintyService.GetUserByEmail(ctx, email)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": email,
		}).Errorf("uc.PrintyService.GetUserByEmail got error %v", err)
		return err
	}

	// user tidak ada -> tetap balas sukses supaya email tidak bisa di-enumerate
	if user == nil {
		return nil
	}

	// 2. cooldown: OTP terakhir dibuat < 60 detik lalu -> abaikan (response tetap sukses)
	latest, err := uc.PrintyService.GetLatestOtp(ctx, user.ID, forgotPasswordOTPType)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"UserID": user.ID,
		}).Errorf("uc.PrintyService.GetLatestOtp %v", err)
		return err
	}
	if latest != nil && time.Since(latest.CreatedAt) < otpResendCooldown {
		return nil
	}

	// 3. batas per jam: cegah spam email dan reset counter attempts berulang-ulang
	total, err := uc.PrintyService.CountOtpsSince(ctx, user.ID, forgotPasswordOTPType, time.Now().Add(-time.Hour))
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"UserID": user.ID,
		}).Errorf("uc.PrintyService.CountOtpsSince %v", err)
		return err
	}
	if total >= maxOTPPerHour {
		return nil
	}

	// 4. matikan OTP lama yang belum dipakai (supaya hanya 1 OTP aktif)
	if err := uc.PrintyService.InvalidateOtps(ctx, user.ID, forgotPasswordOTPType); err != nil {
		log.Logger.WithFields(logrus.Fields{
			"ID":   user.ID,
			"type": forgotPasswordOTPType,
		}).Errorf("uc.PrintyService.InvalidateOtps %v", err)
		return err
	}

	// 5. generate OTP (secret dari config, harus sama dengan yang dipakai CompareOTP)
	rawOTP, hashOTP, err := helper.GenerateOTP("forgot_password")
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"type": forgotPasswordOTPType,
		}).Errorf("helper.GenerateOTP %v", err)
		return err
	}

	expiredAt := time.Now().Add(forgotPasswordOTPTTL)

	// 6. simpan hash-nya
	err = uc.PrintyService.InsertOtp(ctx, models.OTP{
		UserID:      user.ID,
		Type:        forgotPasswordOTPType,
		OTPCodeHash: hashOTP,
		ExpiredAt:   expiredAt,
	})
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"type":   forgotPasswordOTPType,
			"UserID": user.ID,
		}).Errorf("uc.PrintyService.InsertOtp %v", err)
		return err
	}

	// 7. kirim raw OTP ke email
	content, err := utils.BuildOTPEmail(user.FullName, uc.ConfigApp.BaseUrl, rawOTP, expiredAt)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"type":   forgotPasswordOTPType,
			"UserID": user.ID,
		}).Errorf("utils.BuildOTPEmail %v", err)
		return err
	}

	if err := uc.PrintyService.SendEmail(ctx, models.EmailParameter{
		To:      user.Email,
		Subject: content.Subject,
		Body:    content.Body,
	}); err != nil {
		log.Logger.WithFields(logrus.Fields{
			"type":   forgotPasswordOTPType,
			"UserID": user.ID,
		}).Errorf("uc.PrintyService.SendEmail %v", err)
		return err
	}

	return nil
}

// usecase Forgot password
func (uc *PrintyUsecase) ForgotPassword(ctx context.Context, email string) error {
	return uc.sendForgotPasswordOTP(ctx, email)
}

// usecase Resend OTP
func (uc *PrintyUsecase) ResendOTP(ctx context.Context, email string) error {
	return uc.sendForgotPasswordOTP(ctx, email)
}

const maxOTPAttempts = 5

// validasi OTP forgot_password, TIDAK mengubah status OTP (is_used).
// kalau kode salah, attempts bertambah.
func (uc *PrintyUsecase) verifyForgotPasswordOTP(ctx context.Context, email, code string) (*models.User, *models.OTP, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	otpType := "forgot_password"

	// 1. cari user
	user, err := uc.PrintyService.GetUserByEmail(ctx, email)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": email,
		}).Errorf("uc.PrintyService.GetUserByEmail got error %v", err)
		return nil, nil, err
	}

	// user tidak ada -> error generik, supaya email tidak bisa di-enumerate
	if user == nil {
		return nil, nil, herr.ErrInvalidOTP
	}

	// 2. ambil OTP terakhir milik user
	otp, err := uc.PrintyService.GetLatestOtp(ctx, user.ID, otpType)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"UserID": user.ID,
			"type":   otpType,
		}).Errorf("uc.PrintyService.GetLatestOtp %v", err)
		return nil, nil, err
	}

	// 3. tidak ada / sudah dipakai / kedaluwarsa -> error generik
	if otp == nil || otp.IsUsed || time.Now().After(otp.ExpiredAt) {
		return nil, nil, herr.ErrInvalidOTP
	}

	// 4. batas percobaan salah (cegah brute force)
	if otp.Attempts >= maxOTPAttempts {
		return nil, nil, herr.ErrTooManyAttempts
	}

	// 5. bandingkan dengan hash (constant-time)
	if !helper.CompareOTP(code, "forgot_password", otp.OTPCodeHash) {
		if err := uc.PrintyService.IncrementOtpAttempts(ctx, otp.ID); err != nil {
			log.Logger.WithFields(logrus.Fields{
				"OtpID": otp.ID,
			}).Errorf("uc.PrintyService.IncrementOtpAttempts %v", err)
			return nil, nil, err
		}
		return nil, nil, herr.ErrInvalidOTP
	}

	return user, otp, nil
}

// Usecase check OTP (hanya cek, OTP tidak dihanguskan)
func (uc *PrintyUsecase) CheckOTP(ctx context.Context, param models.VerifyOTPParam) error {
	_, _, err := uc.verifyForgotPasswordOTP(ctx, param.Email, param.OTP)
	return err
}

func (uc *PrintyUsecase) ResetPassword(ctx context.Context, param models.ResetPasswordParam) error {
	// 1. validasi OTP (user, expired, attempts, hash)
	user, otp, err := uc.verifyForgotPasswordOTP(ctx, param.Email, param.OTP)
	if err != nil {
		return err
	}

	// 2. hash password baru
	hashedPassword, err := helper.HashPassword(param.NewPassword)
	if err != nil {
		log.Logger.Errorf("helper.HashPassword %v", err)
		return err
	}

	// 3. hanguskan OTP secara atomik (hanya satu request yang boleh lolos)
	marked, err := uc.PrintyService.MarkOtpUsed(ctx, otp.ID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"OtpID": otp.ID,
		}).Errorf("uc.PrintyService.MarkOtpUsed %v", err)
		return err
	}

	// keduluan dipakai request lain -> tolak
	if !marked {
		return herr.ErrInvalidOTP
	}

	// 4. update password user
	if err := uc.PrintyService.UpdateUserPassword(ctx, user.ID, hashedPassword); err != nil {
		log.Logger.WithFields(logrus.Fields{
			"UserID": user.ID,
		}).Errorf("uc.PrintyService.UpdateUserPassword %v", err)
		return err
	}

	return nil
}
