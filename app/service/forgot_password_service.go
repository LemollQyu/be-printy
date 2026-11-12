package service

import (
	"context"
	"time"

	models "github.com/LemollQyu/be-printy/models/authentikasi"
	"github.com/google/uuid"
)

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
