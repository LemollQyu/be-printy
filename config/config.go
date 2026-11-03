package config

import (
	"log"

	"github.com/spf13/viper"
)

func LoadConfig() Config {
	var cfg Config

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./files/config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("error read config file: %v", err)
	}

	cfg.App.Port = viper.GetString("app.port")
	cfg.App.Url = viper.GetString("app.url")

	cfg.Database.Host = viper.GetString("database.host")
	cfg.Database.Port = viper.GetString("database.port")
	cfg.Database.User = viper.GetString("database.user")
	cfg.Database.Password = viper.GetString("database.password")
	cfg.Database.Name = viper.GetString("database.name")

	cfg.Redis.Host = viper.GetString("redis.host")
	cfg.Redis.Port = viper.GetString("redis.port")
	cfg.Redis.Password = viper.GetString("redis.password")

	return cfg

}
