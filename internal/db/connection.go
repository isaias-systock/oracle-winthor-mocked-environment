package db

import (
	"database/sql"
	"fmt"

	"oracle-winthor-mocked-environment/internal/config"

	_ "github.com/sijms/go-ora/v2"
)

func Open(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"oracle://%s:%s@%s:%s/%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBService,
	)

	db, err := sql.Open("oracle", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}