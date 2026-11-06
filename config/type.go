package config

type Config struct {
	App      AppConfig      `yaml:"app" validate:"required"`
	Database DatabaseConfig `yaml:"database" validate:"required"`
	Redis    RedisConfig    `yaml:"redis" validate:"required"`
	Email    EmailSecret    `yaml:"email" validate:"required"`
	Google   GoogleSecret   `yaml:"google" validate:"required"`
}

type AppConfig struct {
	Port        string `yaml:"port" validate:"required"`
	BaseUrl     string `yaml:"base_url" validate:"required"`
	FrontendUrl string `yaml:"frontend_url" validate:"required"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host" validate:"required"`
	User     string `yaml:"user" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	Name     string `yaml:"name" validate:"required"`
	Port     string `yaml:"port" validate:"required"`
}

type RedisConfig struct {
	Host     string `yaml:"host" validate:"required"`
	Port     string `yaml:"port" validate:"required"`
	Password string `yaml:"password" validate:"required"`
}

type EmailSecret struct {
	SmtpHost    string `yaml:"smtp_host" validate:"required"`
	SmtpPort    string `yaml:"smtp_port" validate:"required"`
	EmailHost   string `yaml:"email_host" validate:"required"`
	EmailSecret string `yaml:"email_secret" validate:"required"`
}

type GoogleSecret struct {
	ClientId     string `yaml:"client_id" validate:"required"`
	ClientSecret string `yaml:"client_secret" validate:"required"`
	RedirectURL  string `yaml:"redirect_url" validate:"required"`
}
