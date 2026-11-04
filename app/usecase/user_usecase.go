package usecase

import (
	"context"

	"github.com/LemollQyu/be-printy/helper"
	herr "github.com/LemollQyu/be-printy/helper/error"
	"github.com/LemollQyu/be-printy/infrastructure/log"
	models "github.com/LemollQyu/be-printy/models/authentikasi"
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

	return id, nil
}
