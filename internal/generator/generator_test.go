package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGenerateServiceProducesCompilableProject(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "order-service")
	data := Data{
		ServiceName: "order-service",
		PackageName: "order_service",
		Module:      "github.com/example/order-service",
		Port:        "8081",
		Database:    "postgres",
	}

	if err := GenerateService(outDir, data); err != nil {
		t.Fatalf("GenerateService() error = %v", err)
	}

	requiredFiles := []string{
		"go.mod",
		"cmd/api/main.go",
		"cmd/migrate/main.go",
		"cmd/seed/main.go",
		"internal/config/config.go",
		"internal/domain/entity.go",
		"internal/infrastructure/database/postgres.go",
		"internal/repository/postgres/example_repo.go",
		"internal/delivery/http/router.go",
		"migrations/000001_create_examples.up.sql",
		"migrations/000001_create_examples.down.sql",
	}
	for _, name := range requiredFiles {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("generated file %q missing: %v", name, err)
		}
	}

	runGeneratedCommand(t, outDir, "go", "mod", "tidy")
	runGeneratedCommand(t, outDir, "go", "test", "./...")
	runGeneratedCommand(t, outDir, "go", "run", "./cmd/migrate", "create", "add_customers")

	generatedMigrationFiles := []string{
		"migrations/000002_add_customers.up.sql",
		"migrations/000002_add_customers.down.sql",
	}
	for _, name := range generatedMigrationFiles {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("migration file %q missing after create: %v", name, err)
		}
	}
}

func TestGenerateServiceWithoutDatabaseProducesCompilableProject(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "simple-service")
	data := Data{
		ServiceName: "simple-service",
		PackageName: "simple_service",
		Module:      "github.com/example/simple-service",
		Port:        "8082",
		Database:    "none",
	}

	if err := GenerateService(outDir, data); err != nil {
		t.Fatalf("GenerateService() error = %v", err)
	}

	requiredFiles := []string{
		"go.mod",
		"cmd/api/main.go",
		"internal/config/config.go",
		"internal/domain/entity.go",
		"internal/repository/postgres/example_repo.go",
		"internal/delivery/http/router.go",
	}
	for _, name := range requiredFiles {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("generated file %q missing: %v", name, err)
		}
	}

	forbiddenFiles := []string{
		"cmd/migrate/main.go",
		"cmd/seed/main.go",
		"internal/infrastructure/database/postgres.go",
		"migrations/000001_create_examples.up.sql",
		"migrations/000001_create_examples.down.sql",
	}
	for _, name := range forbiddenFiles {
		if _, err := os.Stat(filepath.Join(outDir, name)); err == nil {
			t.Fatalf("generated file %q exists, want omitted for db=none", name)
		}
	}

	runGeneratedCommand(t, outDir, "go", "mod", "tidy")
	runGeneratedCommand(t, outDir, "go", "test", "./...")
}

func TestGenerateServiceRefusesExistingDirectory(t *testing.T) {
	outDir := t.TempDir()
	err := GenerateService(outDir, Data{
		ServiceName: "order-service",
		PackageName: "order_service",
		Module:      "github.com/example/order-service",
		Port:        "8081",
		Database:    "postgres",
	})
	if err == nil {
		t.Fatal("GenerateService() error = nil, want existing directory error")
	}
}

func runGeneratedCommand(t *testing.T, dir, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated project failed %s %v: %v\n%s", name, args, err, output)
	}
}
