package repository

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PrintyRepository struct {
	Database *gorm.DB
	Redis    *redis.Client
}

func NewPrintyRepository(db *gorm.DB, redis *redis.Client) *PrintyRepository {
	return &PrintyRepository{
		Database: db,
		Redis:    redis,
	}
}
