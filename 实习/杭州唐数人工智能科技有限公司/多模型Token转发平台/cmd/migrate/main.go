package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/school-api/school-api-v1/internal/migrate"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return usage()
	}

	dbURL := os.Getenv("APP_DATABASE__URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		return fmt.Errorf("APP_DATABASE__URL or DATABASE_URL is required")
	}

	migrationsPath := os.Getenv("APP_MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	m, err := migrate.New(dbURL, migrationsPath)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}

	ctx := context.Background()
	cmd := os.Args[1]

	switch cmd {
	case "up":
		if len(os.Args) > 2 {
			n, err := strconv.Atoi(os.Args[2])
			if err != nil {
				return fmt.Errorf("invalid step count %q: %w", os.Args[2], err)
			}
			if err := m.Steps(ctx, n); err != nil {
				return err
			}
			fmt.Printf("applied %d migration(s)\n", n)
			return nil
		}
		if err := m.Up(ctx); err != nil {
			return err
		}
		fmt.Println("all migrations applied")
	case "down":
		if len(os.Args) > 2 {
			n, err := strconv.Atoi(os.Args[2])
			if err != nil {
				return fmt.Errorf("invalid step count %q: %w", os.Args[2], err)
			}
			if err := m.Steps(ctx, -n); err != nil {
				return err
			}
			fmt.Printf("rolled back %d migration(s)\n", n)
			return nil
		}
		if err := m.Down(ctx); err != nil {
			return err
		}
		fmt.Println("all migrations rolled back")
	case "version":
		version, dirty, err := m.Version(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("version: %d, dirty: %v\n", version, dirty)
	default:
		return usage()
	}

	return nil
}

func usage() error {
	return fmt.Errorf("usage: %s {up [N]|down [N]|version}", os.Args[0])
}
