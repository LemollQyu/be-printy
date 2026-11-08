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
	rawToken, err := uc.CreateVerifyToken(ctx, id)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"user_id": id,
		}).Errorf("uc.createVerifyToken got error %v", err)
		return id, nil // user sudah tersimpan, nanti bisa kirim ulang
	}

	// kirim email di background supaya register tidak menunggu SMTP
	email, fullName := user.Email, user.FullName
	go func() {
		if err := uc.SendVerifyEmail(context.Background(), email, fullName, rawToken); err != nil {
			log.Logger.WithFields(logrus.Fields{
				"email": email,
			}).Errorf("uc.SendVerifyEmail got error %v", err)
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
	rawToken, err := uc.CreateVerifyToken(ctx, user.ID)
	if err != nil {
		log.Logger.Errorf("uc.createVerifyToken got error %v", err)
		return err
	}

	// sinkron, supaya user tahu kalau pengiriman gagal
	if err := uc.SendVerifyEmail(ctx, user.Email, user.FullName, rawToken); err != nil {
		log.Logger.Errorf("uc.SendVerifyEmail got error %v", err)
		return err
	}

	return nil
}

// buat token baru, simpan hash-nya, kembalikan raw token
func (uc *PrintyUsecase) CreateVerifyToken(ctx context.Context, userID uuid.UUID) (string, error) {
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
func (uc *PrintyUsecase) SendVerifyEmail(ctx context.Context, email, fullName, rawToken string) error {
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

// signin
func (uc *PrintyUsecase) Signin(ctx context.Context, param *models.LoginParameter) (string, error) {
	userEmail, err := uc.PrintyService.GetUserByEmail(ctx, param.Email)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"email": param.Email,
		}).Errorf("uc.PrintyService.GetUserByEmail got error %v", err)
		return "", err
	}

	// user tidak ada, atau tidak punya password
	if userEmail == nil || userEmail.PasswordHash == nil {
		return "", herr.ErrUserSigninWrong
	}

	isPassword, err := helper.CheckPasswordHash(param.Password, *userEmail.PasswordHash)
	if err != nil {
		log.Logger.Errorf("helper.CheckPasswordHash got error %v", err)
		return "", err
	}

	// password salah
	if !isPassword {
		return "", herr.ErrUserSigninWrong
	}

	// check user verified
	if !userEmail.IsEmailVerified {
		return "", herr.ErrAccountNotVerified
	}

	// generate token session (raw dikirim ke client, hash disimpan di DB)
	rawToken, tokenHash, err := helper.GenerateSessionToken()
	if err != nil {
		log.Logger.Errorf("helper.GenerateSessionToken got error %v", err)
		return "", err
	}

	session := &models.Session{
		UserID:    userEmail.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if param.IPAddress != "" {
		session.IPAddress = &param.IPAddress
	}
	if param.UserAgent != "" {
		session.UserAgent = &param.UserAgent
	}

	if err := uc.PrintyService.CreateSession(ctx, session); err != nil {
		log.Logger.WithFields(logrus.Fields{
			"user_id": userEmail.ID,
		}).Errorf("uc.PrintyService.CreateSession got error %v", err)
		return "", err
	}

	return rawToken, nil

}

func (uc *PrintyUsecase) ValidateSession(ctx context.Context, rawToken string) (*models.User, error) {
	if !helper.IsValidSessionTokenFormat(rawToken) {
		return nil, herr.ErrUnauthorized
	}

	session, err := uc.PrintyService.GetActiveSession(ctx, helper.HashToken(rawToken))
	if err != nil {
		log.Logger.Errorf("uc.PrintyService.GetActiveSession got error %v", err)
		return nil, err
	}

	// token tidak ada atau sudah expired
	if session == nil {
		return nil, herr.ErrUnauthorized
	}

	return &session.User, nil
}
