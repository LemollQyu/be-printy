package repository

import (
	"context"
	"errors"

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
