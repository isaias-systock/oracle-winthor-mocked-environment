package main

import (
	"fmt"
	"log"
	"time"

	"oracle-winthor-mocked-environment/internal/config"
	"oracle-winthor-mocked-environment/internal/db"
)

func main() {
	cfg := config.LoadConfig()

	conn, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	
	defer conn.Close()

	now := time.Now()
	fmt.Println("Conectado com sucesso ao Oracle", now)
}