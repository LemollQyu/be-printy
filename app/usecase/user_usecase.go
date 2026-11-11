package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LemollQyu/be-printy/helper"
	herr "github.com/LemollQyu/be-printy/helper/error"
	"github.com/LemollQyu/be-printy/infrastructure/log"
	models "github.com/LemollQyu/be-printy/models/authentikasi"
	"github.com/LemollQyu/be-printy/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func (uc *PrintyUsecase) InsertUser(ctx context.Context, user *models.User) (uuid.UUID, error) {
	// check email dulu, supaya tidak buang waktu hashing kalau email sudah ada
	userEmail, err := uc.PrintyService.GetUserByEmail(ctx, user.Email)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": user.Email,
		}).Errorf("uc.PrintyService.GetUserByEmail got error %v", err)
		return uuid.Nil, err
	}

	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	if userEmail != nil {
		return uuid.Nil, herr.ErrEmailExist
	}

	// check phone
	if user.Phone != nil {
		userPhone, err := uc.PrintyService.GetUserByPhone(ctx, *user.Phone)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"phone": *user.Phone,
			}).Errorf("uc.PrintyService.GetUserByPhone got error %v", err)
			return uuid.Nil, err
		}

		if userPhone != nil {
			return uuid.Nil, herr.ErrPhoneExist
		}
	}

	// hashing password
	if user.PasswordHash == nil {
		return uuid.Nil, herr.ErrPasswordRequired
	}

	hashed, err := helper.HashPassword(*user.PasswordHash)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": user.Email,
		}).Errorf("helper.HashPassword() got error %v", err)
		return uuid.Nil, err
	}
	user.PasswordHash = &hashed

	// insert
	id, err := uc.PrintyService.InsertUser(ctx, user)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": user.Email,
		}).Errorf("uc.PrintyService.InsertUser got error %v", err)
		return uuid.Nil, err
	}

	// buat token verifikasi
	rawToken, err := uc.createVerifyToken(ctx, id)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"user_id": id,
		}).Errorf("uc.createVerifyToken got error %v", err)
		return id, nil // user sudah tersimpan, nanti bisa kirim ulang
	}

	// kirim email di background supaya register tidak menunggu SMTP
	email, fullName := user.Email, user.FullName
	go func() {
		if err := uc.sendVerifyEmail(context.Background(), email, fullName, rawToken); err != nil {
			log.Logger.WithFields(logrus.Fields{
				"email": email,
			}).Errorf("uc.sendVerifyEmail got error %v", err)
		}
	}()

	return id, nil
}

// send again verify account
func (uc *PrintyUsecase) SendTokenVerifyAccount(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := uc.PrintyService.GetUserByEmail(ctx, email)
	if err != nil {
		log.Logger.Errorf("uc.PrintyService.GetUserByEmail got error %v", err)
		return err
	}

	// user tidak ditemukan -> return nil (jangan bocorkan status email)
	if user == nil {
		return herr.ErrUserNotFound
	}

	// check user sudah terverifikasi belum
	if user.IsEmailVerified {
		return herr.ErrAccountIsVerified
	}

	// cooldown, cegah spam kirim email
	lastToken, err := uc.PrintyService.GetLatestVerifyTokenByUserID(ctx, user.ID)
	if err != nil {
		log.Logger.Errorf("uc.PrintyService.GetLatestVerifyTokenByUserID got error %v", err)
		return err
	}
	if lastToken != nil && time.Since(lastToken.CreatedAt) < time.Minute {
		return herr.ErrTooManyRequests
	}

	// hanguskan token lama yang belum dipakai
	if err := uc.PrintyService.InvalidateVerifyTokens(ctx, user.ID); err != nil {
		log.Logger.Errorf("uc.PrintyService.InvalidateVerifyTokens got error %v", err)
		return err
	}

	// buat token baru (hash disimpan di DB, raw dikirim ke email)
	rawToken, err := uc.createVerifyToken(ctx, user.ID)
	if err != nil {
		log.Logger.Errorf("uc.createVerifyToken got error %v", err)
		return err
	}

	// sinkron, supaya user tahu kalau pengiriman gagal
	if err := uc.sendVerifyEmail(ctx, user.Email, user.FullName, rawToken); err != nil {
		log.Logger.Errorf("uc.sendVerifyEmail got error %v", err)
		return err
	}

	return nil
}

// buat token baru, simpan hash-nya, kembalikan raw token
func (uc *PrintyUsecase) createVerifyToken(ctx context.Context, userID uuid.UUID) (string, error) {
	rawToken, tokenHash, err := helper.GenerateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	_, err = uc.PrintyService.InsertVerifyToken(ctx, models.EmailVerificationToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiredAt: time.Now().Add(utils.VerifyTokenTTL),
	})
	if err != nil {
		return "", fmt.Errorf("insert verify token: %w", err)
	}

	return rawToken, nil
}

// render template + kirim email verifikasi
func (uc *PrintyUsecase) sendVerifyEmail(ctx context.Context, email, fullName, rawToken string) error {
	content, err := utils.BuildVerifyAccountEmail(fullName, uc.ConfigApp.BaseUrl, rawToken)
	if err != nil {
		return err
	}

	return uc.PrintyService.SendEmail(ctx, models.EmailParameter{
		To:      email,
		Subject: content.Subject,
		Body:    content.Body,
	})
}

// verify account
func (uc *PrintyUsecase) VerifyAccount(ctx context.Context, rawToken string) error {
	tokenHash := helper.HashToken(rawToken)

	verifyToken, err := uc.PrintyService.GetVerifyTokenByHash(ctx, tokenHash)
	if err != nil {
		log.Logger.Errorf("uc.PrintyService.GetVerifyTokenByHash got error %v", err)
		return err
	}
	if verifyToken == nil {
		return herr.ErrInvalidToken
	}

	if verifyToken.UsedAt != nil {
		return herr.ErrTokenUsed
	}

	if time.Now().After(verifyToken.ExpiredAt) {
		return herr.ErrTokenExpired
	}

	user, err := uc.PrintyService.GetUserByID(ctx, verifyToken.UserID)
	if err != nil {
		log.Logger.Errorf("uc.PrintyService.GetUserByID got error %v", err)
		return err
	}
	if user == nil {
		return herr.ErrInvalidToken
	}

	// sudah verified? anggap sukses
	if user.IsEmailVerified {
		return nil
	}

	// tandai token + verify akun dalam satu transaksi
	if err := uc.PrintyService.VerifyAccountTx(ctx, verifyToken.ID, user.ID); err != nil {
		if !errors.Is(err, herr.ErrTokenUsed) {
			log.Logger.Errorf("uc.PrintyService.VerifyAccountTx got error %v", err)
		}
		return err
	}

	return nil
}

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
