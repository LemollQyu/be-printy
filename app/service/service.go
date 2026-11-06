package service

import (
	"github.com/LemollQyu/be-printy/app/repository"
	"github.com/LemollQyu/be-printy/config"
)

type PrintyService struct {
	PrintyRepositry repository.PrintyRepository
	ConfigEmail     config.EmailSecret
}

func NewPrintyService(printyRepository repository.PrintyRepository, cfgEmail config.EmailSecret) *PrintyService {
	return &PrintyService{
		PrintyRepositry: printyRepository,
		ConfigEmail:     cfgEmail,
	}
}
