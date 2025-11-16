package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
}

type DBConfig struct {
	PgUser     string `env:"PGUSER"`
	PgPassword string `env:"PGPASSWORD"`
	PgHost     string `env:"PGHOST"`
	PgPort     int    `env:"PGPORT"`
	PgDatabase string `env:"PGDATABASE"`
	PgSSLMode  string `env:"PGSSLMODE"`
}

type ServerConfig struct {
	HTTPPort string `env:"HTTP_PORT"`
}

func MustLoad() *Config {
	cfg := &Config{}

	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	return cfg
}
