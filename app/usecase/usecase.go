package usecase

import (
	"github.com/LemollQyu/be-printy/app/service"
	"github.com/LemollQyu/be-printy/config"
)

type PrintyUsecase struct {
	PrintyService service.PrintyService
	ConfigApp     config.AppConfig
}

func NewPrintyUsecase(printyService service.PrintyService, cfgApp config.AppConfig) *PrintyUsecase {
	return &PrintyUsecase{
		PrintyService: printyService,
		ConfigApp:     cfgApp,
	}
}
