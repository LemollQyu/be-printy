package service

import "github.com/LemollQyu/be-printy/app/repository"

type PrintyService struct {
	PrintyRepositry repository.PrintyRepository
}

func NewPrintyService(printyRepository repository.PrintyRepository) *PrintyService {
	return &PrintyService{
		PrintyRepositry: printyRepository,
	}
}
