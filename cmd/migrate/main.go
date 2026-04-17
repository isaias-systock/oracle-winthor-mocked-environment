package main

import (
	"fmt"
	"log"
	"os"

	"oracle-winthor-mocked-environment/internal/config"
	"oracle-winthor-mocked-environment/internal/migration"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	cfg := config.Load()

	switch command {
	case "create":
		if len(os.Args) < 3 {
			log.Fatal("informe o nome da migration")
		}
		name := os.Args[2]
		if err := migration.CreateMigration(name); err != nil {
			log.Fatal(err)
		}

	case "create-seed":
		if len(os.Args) < 3 {
			log.Fatal("informe o nome da seed")
		}
		name := os.Args[2]
		if err := migration.CreateSeed(name); err != nil {
			log.Fatal(err)
		}

	case "up":
		if err := migration.Up(cfg); err != nil {
			log.Fatal(err)
		}

	case "down":
		if err := migration.Down(cfg); err != nil {
			log.Fatal(err)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("uso:")
	fmt.Println("  go run ./cmd/migrate create <nome>")
	fmt.Println("  go run ./cmd/migrate create-seed <nome>")
	fmt.Println("  go run ./cmd/migrate up")
	fmt.Println("  go run ./cmd/migrate down")
}
