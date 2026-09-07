package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/adminuser"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/config"
	"github.com/school-api/school-api-v1/internal/repository"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var (
		username = flag.String("username", os.Getenv("APP_ADMIN__SEED_USERNAME"), "admin username")
		password = flag.String("password", os.Getenv("APP_ADMIN__SEED_PASSWORD"), "admin password")
		role     = flag.String("role", os.Getenv("APP_ADMIN__SEED_ROLE"), "admin role (super_admin or admin)")
	)
	flag.Parse()

	if flag.NArg() == 0 || flag.Arg(0) != "create-admin" {
		return errors.New("usage: admin create-admin -username=NAME -password=PASSWORD -role=ROLE")
	}

	if *password == "" {
		*password = os.Getenv("APP_ADMIN__SEED_PASSWORD")
	}
	if *username == "" || *password == "" {
		return errors.New("username and password are required (use -password or APP_ADMIN__SEED_PASSWORD)")
	}

	cfgPath := os.Getenv("APP_CONFIG_PATH")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := repository.OpenEntClient(cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer client.Close()

	repo := repository.NewEntAdminRepository(client)

	hash, err := auth.HashPassword(auth.BcryptCostForEnv(cfg.Server.Env), *password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	adminRole := adminuser.Role(*role)
	if adminRole != adminuser.RoleSuperAdmin && adminRole != adminuser.RoleAdmin {
		return fmt.Errorf("invalid role %q", *role)
	}

	admin, err := repo.CreateAdmin(ctx, *username, hash, adminRole)
	if err != nil {
		if ent.IsConstraintError(err) {
			return fmt.Errorf("username %q already exists", *username)
		}
		return fmt.Errorf("create admin: %w", err)
	}

	fmt.Printf("created admin: id=%s username=%s role=%s\n", admin.ID.String(), admin.Username, admin.Role)
	return nil
}
