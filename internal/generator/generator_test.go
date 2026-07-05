package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		Transport:   "both",
		GRPCPort:    "9091",
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
		"internal/delivery/grpc/server.go",
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
		Transport:   "http",
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
		"internal/delivery/grpc/server.go",
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

func TestGenerateServiceWithGRPCTransportProducesCompilableProject(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "grpc-service")
	data := Data{
		ServiceName: "grpc-service",
		PackageName: "grpc_service",
		Module:      "github.com/example/grpc-service",
		Port:        "8083",
		Database:    "none",
		Transport:   "grpc",
		GRPCPort:    "9093",
	}

	if err := GenerateService(outDir, data); err != nil {
		t.Fatalf("GenerateService() error = %v", err)
	}

	requiredFiles := []string{
		"go.mod",
		"cmd/api/main.go",
		"internal/config/config.go",
		"internal/delivery/grpc/server.go",
		"internal/domain/entity.go",
		"internal/repository/postgres/example_repo.go",
	}
	for _, name := range requiredFiles {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("generated file %q missing: %v", name, err)
		}
	}

	forbiddenFiles := []string{
		"internal/delivery/http/router.go",
		"cmd/migrate/main.go",
		"cmd/seed/main.go",
		"internal/infrastructure/database/postgres.go",
		"migrations/000001_create_examples.up.sql",
		"migrations/000001_create_examples.down.sql",
	}
	for _, name := range forbiddenFiles {
		if _, err := os.Stat(filepath.Join(outDir, name)); err == nil {
			t.Fatalf("generated file %q exists, want omitted for grpc db=none", name)
		}
	}

	envExample := readGeneratedFile(t, outDir, ".env.example")
	if !strings.Contains(envExample, "GRPC_PORT=9093") {
		t.Fatalf(".env.example missing GRPC_PORT: %q", envExample)
	}
	if hasLinePrefix(envExample, "PORT=") {
		t.Fatalf(".env.example contains HTTP PORT for grpc transport: %q", envExample)
	}
	if hasLinePrefix(envExample, "DATABASE_URL=") {
		t.Fatalf(".env.example contains DATABASE_URL for db=none: %q", envExample)
	}

	runGeneratedCommand(t, outDir, "go", "mod", "tidy")
	runGeneratedCommand(t, outDir, "go", "test", "./...")
}

func TestAddHTTPHandlerProducesCompilableFiles(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "handler-service")
	data := Data{
		ServiceName: "handler-service",
		PackageName: "handler_service",
		Module:      "github.com/example/handler-service",
		Port:        "8084",
		Database:    "none",
		Transport:   "http",
	}

	if err := GenerateService(outDir, data); err != nil {
		t.Fatalf("GenerateService() error = %v", err)
	}
	resource, err := AddHTTPHandler(outDir, "customer")
	if err != nil {
		t.Fatalf("AddHTTPHandler() error = %v", err)
	}
	if resource.Type != "Customer" {
		t.Fatalf("resource type = %q, want Customer", resource.Type)
	}
	if !resource.RouterWired {
		t.Fatal("RouterWired = false, want true")
	}

	requiredFiles := []string{
		"internal/domain/customer.go",
		"internal/usecase/customer_usecase.go",
		"internal/repository/postgres/customer_repo.go",
		"internal/delivery/http/handler/customer_handler.go",
	}
	for _, name := range requiredFiles {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("generated file %q missing: %v", name, err)
		}
	}

	router := readGeneratedFile(t, outDir, "internal/delivery/http/router.go")
	for _, want := range []string{
		"customerRepo := postgres.NewInMemoryCustomerRepository()",
		"customerUsecase := usecase.NewCustomerUsecase(customerRepo)",
		"customerHandler := handler.NewCustomerHandler(customerUsecase)",
		`customers := v1.Group("/customers")`,
		`customers.POST("", customerHandler.Create)`,
		`customers.GET("", customerHandler.List)`,
		`customers.GET("/:id", customerHandler.GetByID)`,
	} {
		if !strings.Contains(router, want) {
			t.Fatalf("router.go missing %q:\n%s", want, router)
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

func readGeneratedFile(t *testing.T, dir, name string) string {
	t.Helper()

	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read generated file %q: %v", name, err)
	}
	return string(b)
}

func hasLinePrefix(s, prefix string) bool {
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}
