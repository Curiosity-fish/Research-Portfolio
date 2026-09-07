// Command mockgen generates safe, tagged mock data for development and test
// environments. It refuses to run against production databases.
//
// Usage:
//
//	go run ./cmd/mockgen -config=configs/config.yaml
//	go run ./cmd/mockgen -cleanup -config=configs/config.yaml
//	APP_SERVER__ENV=test go run ./cmd/mockgen
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/school-api/school-api-v1/ent/adminuser"
	"github.com/school-api/school-api-v1/internal/config"
	"github.com/school-api/school-api-v1/internal/mockdata"
	"github.com/school-api/school-api-v1/internal/repository"
)

func main() {
	if err := run(); err != nil {
		slog.Error("mockgen failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		cleanup       = flag.Bool("cleanup", false, "remove all mockgen-tagged data instead of generating")
		adminUser     = flag.String("admin-username", "mockgen-admin", "username for the generated admin")
		adminPass     = flag.String("admin-password", "MockGen@123", "password for the generated admin")
		groupCount    = flag.Int("groups", 2, "number of groups to generate")
		rootCount     = flag.Int("root-departments", 2, "number of root departments to generate")
		children      = flag.Int("children-per-root", 2, "number of children per root department")
		userCount     = flag.Int("users", 5, "number of users to generate")
		tokensPerUser = flag.Int("tokens-per-user", 2, "number of API tokens per user")
		seed          = flag.Int64("seed", mockdata.DefaultSeed, "random seed for deterministic output")
	)
	flag.Parse()

	cfgPath := os.Getenv("APP_CONFIG_PATH")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err := mockdata.MustBeNonProduction(cfg.Server.Env); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := repository.OpenEntClient(cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer client.Close()

	gen := mockdata.NewGenerator(client, *seed)

	if *cleanup {
		if err := gen.Cleanup(ctx); err != nil {
			return fmt.Errorf("cleanup mock data: %w", err)
		}
		slog.Info("mockgen cleanup complete")
		return nil
	}

	scenario := mockdata.ScenarioConfig{
		AdminUsername:       *adminUser,
		AdminPassword:       *adminPass,
		AdminRole:           adminRole(),
		GroupCount:          *groupCount,
		RootDepartmentCount: *rootCount,
		ChildrenPerRoot:     *children,
		UserCount:           *userCount,
		TokensPerUser:       *tokensPerUser,
	}

	set, err := gen.Generate(ctx, scenario)
	if err != nil {
		return fmt.Errorf("generate mock data: %w", err)
	}

	printResults(set)
	return nil
}

func adminRole() adminuser.Role {
	// The default role is admin rather than super_admin to avoid accidentally
	// creating privileged accounts during routine testing.
	return adminuser.RoleAdmin
}

func printResults(set *mockdata.GeneratedSet) {
	fmt.Println("=== Mock data generation complete ===")
	fmt.Printf("Admin:     %s (%s)\n", set.Admin.Username, set.Admin.Role)
	fmt.Printf("Groups:    %d\n", len(set.Groups))
	fmt.Printf("Departments: %d\n", len(set.Departments))
	fmt.Printf("Users:     %d\n", len(set.Users))
	fmt.Printf("Tokens:    %d\n", len(set.Tokens))
	fmt.Println()
	fmt.Println("Sample API keys (save these now; they are not stored in plaintext):")
	for i, tok := range set.Tokens {
		if i >= 5 {
			fmt.Printf("... and %d more\n", len(set.Tokens)-i)
			break
		}
		fmt.Printf("  user=%s key=%s\n", tok.UserID, tok.Plaintext)
	}
	fmt.Println()
	fmt.Println("Run with -cleanup to remove all mockgen-tagged records.")
}
