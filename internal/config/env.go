package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv string

	OracleHost    string
	OraclePort    string
	OracleService string

	OracleAdminUser     string
	OracleAdminPassword string

	OracleAppUser     string
	OracleAppPassword string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		AppEnv: os.Getenv("APP_ENV"),

		OracleHost:    os.Getenv("ORACLE_HOST"),
		OraclePort:    os.Getenv("ORACLE_PORT"),
		OracleService: os.Getenv("ORACLE_SERVICE"),

		OracleAdminUser:     os.Getenv("ORACLE_ADMIN_USER"),
		OracleAdminPassword: os.Getenv("ORACLE_ADMIN_PASSWORD"),

		OracleAppUser:     os.Getenv("ORACLE_APP_USER"),
		OracleAppPassword: os.Getenv("ORACLE_APP_PASSWORD"),
	}

	if cfg.OracleHost == "" || cfg.OraclePort == "" || cfg.OracleService == "" {
		log.Fatal("ORACLE_HOST, ORACLE_PORT e ORACLE_SERVICE são obrigatórios")
	}

	return cfg
}