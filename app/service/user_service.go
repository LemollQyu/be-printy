package service

import (
	"context"

	"github.com/google/uuid"

	models "github.com/LemollQyu/be-printy/models/authentikasi"
)

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
