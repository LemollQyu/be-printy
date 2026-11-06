package main

import (
	"fmt"

	"github.com/LemollQyu/be-printy/app/handler"
	"github.com/LemollQyu/be-printy/app/repository"
	"github.com/LemollQyu/be-printy/app/resource"
	"github.com/LemollQyu/be-printy/app/service"
	"github.com/LemollQyu/be-printy/app/usecase"
	"github.com/LemollQyu/be-printy/config"
	"github.com/LemollQyu/be-printy/infrastructure/log"
	"github.com/LemollQyu/be-printy/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Welcome in App Restfull API Printy")

	cfg := config.LoadConfig()
	port := cfg.App.Port

	fmt.Println("Config semua disni")
	fmt.Println("APP CONFIG:", cfg.App)
	fmt.Println("DATABASE CONFIG:", cfg.Database)
	fmt.Println("REDIS CONFIG:", cfg.Redis)
	fmt.Println("EMAIL CONFIG:", cfg.Email)

	db := resource.InitDB(&cfg)
	redis := resource.InitRedis(&cfg)
	log.SetupLogger()

	printyRepository := repository.NewPrintyRepository(db, redis)
	printyService := service.NewPrintyService(*printyRepository, cfg.Email)
	printyUsecase := usecase.NewPrintyUsecase(*printyService, cfg.App)
	printyHandler := handler.NewPrintyHandler(*printyUsecase, cfg.App)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	router.Static("/public", "./public")

	routes.SetupRoutes(router, *printyHandler)
	log.Logger.Printf("HTTP server running on port : %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Logger.Fatalf("HTTP server error: %v", err)
	}

}
