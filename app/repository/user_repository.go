package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	models "github.com/LemollQyu/be-printy/models/authentikasi"
	"gorm.io/gorm"
)

func (r *PrintyRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.Database.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *PrintyRepository) FindUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.Database.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *PrintyRepository) InsertUser(ctx context.Context, user *models.User) (uuid.UUID, error) {
	if err := r.Database.WithContext(ctx).Create(user).Error; err != nil {
		return uuid.Nil, err
	}

	return user.ID, nil
}

func (r *PrintyRepository) FindUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User

	err := r.Database.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *PrintyRepository) InsertVerifyToken(ctx context.Context, verifyToken models.EmailVerificationToken) (*models.EmailVerificationToken, error) {
	if err := r.Database.WithContext(ctx).Omit("User").Create(&verifyToken).Error; err != nil {
		return nil, err
	}

	return &verifyToken, nil
}

// update user email verified true
func (r *PrintyRepository) UpdateAccountVerify(ctx context.Context, userID uuid.UUID) error {
	res := r.Database.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_email_verified", true)
	if res.Error != nil {
		return res.Error
	}

	// user tidak ditemukan
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *PrintyRepository) FindVerifyTokenByHash(ctx context.Context, tokenHash string) (*models.EmailVerificationToken, error) {
	var token models.EmailVerificationToken

	err := r.Database.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &token, nil
}

func (r *PrintyRepository) UpdateUseToken(ctx context.Context, tokenID uuid.UUID) error {
	result := r.Database.WithContext(ctx).
		Model(&models.EmailVerificationToken{}).
		Where("id = ? AND used_at IS NULL", tokenID).
		Update("used_at", time.Now())

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *PrintyRepository) VerifyAccountTx(ctx context.Context, tokenID, userID uuid.UUID) error {
	return r.Database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// tandai token terpakai (atomik, hanya jika belum dipakai & belum expired)
		res := tx.Model(&models.EmailVerificationToken{}).
			Where("id = ? AND used_at IS NULL AND expired_at > ?", tokenID, time.Now()).
			Update("used_at", time.Now())
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		// verify akun
		res = tx.Model(&models.User{}).
			Where("id = ?", userID).
			Update("is_email_verified", true)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

func (r *PrintyRepository) FindLatestVerifyTokenByUserID(ctx context.Context, userID uuid.UUID) (*models.EmailVerificationToken, error) {
	var token models.EmailVerificationToken

	err := r.Database.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &token, nil
}

// tandai semua token milik user yang belum dipakai sebagai terpakai
func (r *PrintyRepository) InvalidateVerifyTokens(ctx context.Context, userID uuid.UUID) error {
	return r.Database.WithContext(ctx).
		Model(&models.EmailVerificationToken{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", time.Now()).Error
}

func (r *PrintyRepository) InsertSession(ctx context.Context, session *models.Session) error {
	return r.Database.WithContext(ctx).Omit("User").Create(session).Error
}

func (r *PrintyRepository) FindActiveSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	var session models.Session

	err := r.Database.WithContext(ctx).
		Preload("User").
		Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &session, nil
}
