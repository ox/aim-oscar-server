package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type config struct {
	AppConfig   AppConfig   `yaml:"app"`
	DBConfig    DBConfig    `yaml:"db"`
	OscarConfig OscarConfig `yaml:"oscar"`
}

type AppConfig struct {
	LogLevel string `yaml:"log_level" env-default:"debug"`
}

type OscarConfig struct {
	Addr string `yaml:"addr" env:"OSCAR_ADDR" env-required:"true"`
	BOS  string `yaml:"bos" env:"OSCAR_BOS" env-required:"true"`
}

type DBConfig struct {
	User     string `yaml:"user" env:"DB_USERNAME" env-required:"true"`
	Password string `yaml:"password" env:"DB_PASSWORD" env-required:"true"`
	Name     string `yaml:"name" env:"DB_NAME" env-required:"true"`
	Host     string `yaml:"host" env:"DB_HOST" env-required:"true"`
	Port     int    `yaml:"port" env:"DB_PORT" env-required:"true"`
	SSLMode  string `yaml:"ssl_mode" env:"DB_SSLMODE" env-default:"disable"`
}

func FromFile(filepath string) (*config, error) {
	var cfg config

	err := cleanenv.ReadConfig(filepath, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
