package usecase

import "github.com/LemollQyu/be-printy/app/service"

type PrintyUsecase struct {
	PrintyService service.PrintyService
}

func NewPrintyUsecase(printyService service.PrintyService) *PrintyUsecase {
	return &PrintyUsecase{
		PrintyService: printyService,
	}
}
