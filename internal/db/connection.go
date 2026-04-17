package db

import (
	"database/sql"
	"fmt"

	"oracle-winthor-mocked-environment/internal/config"

	_ "github.com/sijms/go-ora/v2"
)

func open(user, password, host, port, service string) (*sql.DB, error) {
	dsn := fmt.Sprintf("oracle://%s:%s@%s:%s/%s", user, password, host, port, service)

	conn, err := sql.Open("oracle", dsn)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return conn, nil
}

func OpenAdmin(cfg config.Config) (*sql.DB, error) {
	return open(
		cfg.OracleAdminUser,
		cfg.OracleAdminPassword,
		cfg.OracleHost,
		cfg.OraclePort,
		cfg.OracleService,
	)
}

func OpenApp(cfg config.Config) (*sql.DB, error) {
	return open(
		cfg.OracleAppUser,
		cfg.OracleAppPassword,
		cfg.OracleHost,
		cfg.OraclePort,
		cfg.OracleService,
	)
}