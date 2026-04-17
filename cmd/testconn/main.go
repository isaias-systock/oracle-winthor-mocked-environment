package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"oracle-winthor-mocked-environment/internal/config"
	"oracle-winthor-mocked-environment/internal/db"
)

func main() {
	cfg := config.Load()

	target := "admin"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}

	var (
		conn      interface{ Close() error }
		connLabel string
		err       error
	)

	switch target {
	case "admin":
		connLabel = cfg.OracleAdminUser
		conn, err = db.OpenAdmin(cfg)
	case "app":
		connLabel = cfg.OracleAppUser
		conn, err = db.OpenApp(cfg)
	default:
		log.Fatalf("uso: go run ./cmd/testconn [admin|app]")
	}

	if err != nil {
		log.Fatalf("falha ao conectar com usuario %s: %v", connLabel, err)
	}
	defer conn.Close()

	fmt.Printf(
		"Conectado com sucesso ao Oracle em %s.",
		time.Now().Format(time.RFC3339),
	)
}
