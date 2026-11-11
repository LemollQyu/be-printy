package service

import (
	"bytes"
	"context"
	"errors"
	"mime"
	"mime/quotedprintable"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/LemollQyu/be-printy/infrastructure/log"
	models "github.com/LemollQyu/be-printy/models/authentikasi"
)

func (svc *PrintyService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := svc.PrintyRepositry.FindUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (svc *PrintyService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := svc.PrintyRepositry.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (svc *PrintyService) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {

	user, err := svc.PrintyRepositry.FindUserByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (svc *PrintyService) InsertUser(ctx context.Context, user *models.User) (uuid.UUID, error) {
	userID, err := svc.PrintyRepositry.InsertUser(ctx, user)

	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func (svc *PrintyService) SendEmail(ctx context.Context, req models.EmailParameter) error {
	if strings.ContainsAny(req.To+req.Subject, "\r\n") {
		return errors.New("invalid email header")
	}

	from := svc.ConfigEmail.EmailHost
	addr := svc.ConfigEmail.SmtpHost + ":" + svc.ConfigEmail.SmtpPort

	// body -> CRLF + quoted-printable
	body := strings.ReplaceAll(strings.ReplaceAll(req.Body, "\r\n", "\n"), "\n", "\r\n")
	var qp bytes.Buffer
	w := quotedprintable.NewWriter(&qp)
	w.Write([]byte(body))
	w.Close()

	msg := []byte("From: " + from + "\r\n" +
		"To: " + req.To + "\r\n" +
		"Subject: " + mime.QEncoding.Encode("UTF-8", req.Subject) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: quoted-printable\r\n\r\n" +
		qp.String())

	auth := smtp.PlainAuth("", from, svc.ConfigEmail.EmailSecret, svc.ConfigEmail.SmtpHost)

	// kirim dengan timeout
	done := make(chan error, 1)
	go func() { done <- smtp.SendMail(addr, auth, from, []string{req.To}, msg) }()

	select {
	case err := <-done:
		if err != nil {
			log.Logger.WithFields(logrus.Fields{"email_to": req.To, "service": "SendEmail"}).
				Error("failed to send email: ", err)
		}
		return err
	case <-time.After(20 * time.Second):
		return errors.New("send email timeout")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (svc *PrintyService) InsertVerifyToken(ctx context.Context, verifyToken models.EmailVerificationToken) (*models.EmailVerificationToken, error) {
	token, err := svc.PrintyRepositry.InsertVerifyToken(ctx, verifyToken)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (svc *PrintyService) GetVerifyTokenByHash(ctx context.Context, tokenHash string) (*models.EmailVerificationToken, error) {
	token, err := svc.PrintyRepositry.FindVerifyTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (svc *PrintyService) UpdateAccountVerify(ctx context.Context, id uuid.UUID) error {
	err := svc.PrintyRepositry.UpdateAccountVerify(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (svc *PrintyService) UpdateVerifyTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	err := svc.PrintyRepositry.UpdateUseToken(ctx, tokenID)
	if err != nil {
		return err
	}

	return nil
}

func (svc *PrintyService) VerifyAccountTx(ctx context.Context, tokenID, userID uuid.UUID) error {
	err := svc.PrintyRepositry.VerifyAccountTx(ctx, tokenID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (svc *PrintyService) GetLatestVerifyTokenByUserID(ctx context.Context, userID uuid.UUID) (*models.EmailVerificationToken, error) {
	token, err := svc.PrintyRepositry.FindLatestVerifyTokenByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (svc *PrintyService) InvalidateVerifyTokens(ctx context.Context, userID uuid.UUID) error {
	err := svc.PrintyRepositry.InvalidateVerifyTokens(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (svc *PrintyService) InsertOtp(ctx context.Context, otp models.OTP) error {
	err := svc.PrintyRepositry.InsertOtp(ctx, otp)
	if err != nil {
		return err
	}

	return nil
}

func (svc *PrintyService) InvalidateOtps(ctx context.Context, userID uuid.UUID, otpType string) error {
	return svc.PrintyRepositry.InvalidateOtps(ctx, userID, otpType)
}

func (svc *PrintyService) GetLatestOtp(ctx context.Context, userID uuid.UUID, otpType string) (*models.OTP, error) {
	return svc.PrintyRepositry.GetLatestOtp(ctx, userID, otpType)
}

func (svc *PrintyService) IncrementOtpAttempts(ctx context.Context, otpID uuid.UUID) error {
	return svc.PrintyRepositry.IncrementOtpAttempts(ctx, otpID)
}

func (svc *PrintyService) MarkOtpUsed(ctx context.Context, otpID uuid.UUID) (bool, error) {
	return svc.PrintyRepositry.MarkOtpUsed(ctx, otpID)
}

func (svc *PrintyService) UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	return svc.PrintyRepositry.UpdateUserPassword(ctx, userID, hashedPassword)
}

func (svc *PrintyService) CountOtpsSince(ctx context.Context, userID uuid.UUID, otpType string, since time.Time) (int64, error) {
	return svc.PrintyRepositry.CountOtpsSince(ctx, userID, otpType, since)
}
