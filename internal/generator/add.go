package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type ResourceData struct {
	Name             string
	RouteName        string
	Package          string
	Type             string
	Module           string
	HasPostgres      bool
	RouterWired      bool
	MigrationCreated bool
}

func AddHTTPHandler(projectDir, name string) (*ResourceData, error) {
	if !validResourceName(name) {
		return nil, fmt.Errorf("resource name cuma boleh huruf kecil, angka, dan dash. contoh: customer atau sales-order")
	}
	if _, err := os.Stat(filepath.Join(projectDir, "internal", "delivery", "http")); err != nil {
		return nil, fmt.Errorf("project ini tidak punya HTTP transport, generate dengan -transport=http atau -transport=both")
	}

	module, err := readModule(projectDir)
	if err != nil {
		return nil, err
	}

	data := &ResourceData{
		Name:        name,
		RouteName:   pluralize(name),
		Package:     toSnakeName(name),
		Type:        toPascalName(name),
		Module:      module,
		HasPostgres: hasPostgresProject(projectDir),
	}

	files := map[string]string{
		filepath.Join("internal", "domain", data.Package+".go"):                              domainResourceTemplate,
		filepath.Join("internal", "usecase", data.Package+"_usecase.go"):                     usecaseResourceTemplate,
		filepath.Join("internal", "repository", "postgres", data.Package+"_repo.go"):         repositoryResourceTemplate,
		filepath.Join("internal", "delivery", "http", "handler", data.Package+"_handler.go"): handlerResourceTemplate,
	}

	for rel, tmpl := range files {
		if err := renderNewFile(filepath.Join(projectDir, rel), tmpl, data); err != nil {
			return nil, err
		}
	}

	data.RouterWired, err = wireHTTPRouter(projectDir, data)
	if err != nil {
		return nil, err
	}

	if data.HasPostgres {
		data.MigrationCreated, err = createPlaceholderMigration(projectDir, data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func createPlaceholderMigration(projectDir string, data *ResourceData) (bool, error) {
	migrationsDir := filepath.Join(projectDir, "migrations")
	if _, err := os.Stat(migrationsDir); err != nil {
		return false, nil
	}

	next, err := nextMigrationVersion(migrationsDir)
	if err != nil {
		return false, err
	}

	base := fmt.Sprintf("%06d_create_%s", next, data.RouteName)
	upPath := filepath.Join(migrationsDir, base+".up.sql")
	downPath := filepath.Join(migrationsDir, base+".down.sql")
	up := fmt.Sprintf(`-- TODO: define %s table schema.
-- Example starting point:
--
-- CREATE TABLE IF NOT EXISTS %s (
--     id UUID PRIMARY KEY,
--     name TEXT NOT NULL,
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );
`, data.RouteName, data.RouteName)
	down := fmt.Sprintf(`-- TODO: define rollback for %s table schema.
-- Example:
--
-- DROP TABLE IF EXISTS %s;
`, data.RouteName, data.RouteName)

	if err := os.WriteFile(upPath, []byte(up), 0o644); err != nil {
		return false, err
	}
	if err := os.WriteFile(downPath, []byte(down), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func nextMigrationVersion(migrationsDir string) (int, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return 0, err
	}

	re := regexp.MustCompile(`^(\d+)_.*\.(up|down)\.sql$`)
	maxVersion := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := re.FindStringSubmatch(entry.Name())
		if len(matches) != 3 {
			continue
		}
		version := 0
		for _, r := range matches[1] {
			version = version*10 + int(r-'0')
		}
		if version > maxVersion {
			maxVersion = version
		}
	}
	return maxVersion + 1, nil
}

func wireHTTPRouter(projectDir string, data *ResourceData) (bool, error) {
	path := filepath.Join(projectDir, "internal", "delivery", "http", "router.go")
	b, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	content := string(b)
	if strings.Contains(content, "New"+data.Type+"Handler") || strings.Contains(content, `v1.Group("/`+data.RouteName+`")`) {
		return false, fmt.Errorf("router sudah punya wiring untuk %q", data.Name)
	}

	wiringAnchor := "\texampleHandler := handler.NewExampleHandler(exampleUsecase)\n"
	routeAnchor := "\t\texamples.GET(\"/:id\", exampleHandler.GetByID)\n"
	if !strings.Contains(content, wiringAnchor) || !strings.Contains(content, routeAnchor) {
		return false, nil
	}

	wiring := renderRouterWiring(data)
	routes := renderRouterRoutes(data)
	content = strings.Replace(content, wiringAnchor, wiringAnchor+wiring, 1)
	content = strings.Replace(content, routeAnchor, routeAnchor+routes, 1)

	formatted, err := format.Source([]byte(content))
	if err != nil {
		return false, fmt.Errorf("format router.go: %w", err)
	}
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func renderRouterWiring(data *ResourceData) string {
	if data.HasPostgres {
		return fmt.Sprintf(`
	%sRepo := postgres.New%sRepository(db)
	%sUsecase := usecase.New%sUsecase(%sRepo)
	%sHandler := handler.New%sHandler(%sUsecase)
`, data.Package, data.Type, data.Package, data.Type, data.Package, data.Package, data.Type, data.Package)
	}
	return fmt.Sprintf(`
	%sRepo := postgres.NewInMemory%sRepository()
	%sUsecase := usecase.New%sUsecase(%sRepo)
	%sHandler := handler.New%sHandler(%sUsecase)
`, data.Package, data.Type, data.Package, data.Type, data.Package, data.Package, data.Type, data.Package)
}

func renderRouterRoutes(data *ResourceData) string {
	return fmt.Sprintf(`

		%s := v1.Group("/%s")
		%s.POST("", %sHandler.Create)
		%s.GET("", %sHandler.List)
		%s.GET("/:id", %sHandler.GetByID)
`, data.RouteName, data.RouteName, data.RouteName, data.Package, data.RouteName, data.Package, data.RouteName, data.Package)
}

func renderNewFile(path, raw string, data *ResourceData) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file %q sudah ada", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmpl, err := template.New(filepath.Base(path)).Parse(raw)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return err
	}
	if filepath.Ext(path) == ".go" {
		formatted, err := format.Source(b.Bytes())
		if err != nil {
			return fmt.Errorf("format %q: %w", path, err)
		}
		return os.WriteFile(path, formatted, 0o644)
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

func readModule(projectDir string) (string, error) {
	b, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("baca go.mod: %w", err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module path tidak ditemukan di go.mod")
}

func hasPostgresProject(projectDir string) bool {
	_, err := os.Stat(filepath.Join(projectDir, "internal", "infrastructure", "database", "postgres.go"))
	return err == nil
}

var resourceNameRe = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func validResourceName(s string) bool {
	return resourceNameRe.MatchString(s)
}

func toSnakeName(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

func toPascalName(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func pluralize(s string) string {
	if strings.HasSuffix(s, "s") {
		return s
	}
	if strings.HasSuffix(s, "y") {
		return strings.TrimSuffix(s, "y") + "ies"
	}
	return s + "s"
}

const domainResourceTemplate = `package domain

import (
	"context"
	"time"
)

type {{.Type}} struct {
	ID        string    ` + "`json:\"id\"{{if .HasPostgres}} gorm:\"type:uuid;primaryKey\"{{end}}`" + `
	Name      string    ` + "`json:\"name\"{{if .HasPostgres}} gorm:\"not null\"{{end}}`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"{{if .HasPostgres}} gorm:\"autoCreateTime\"{{end}}`" + `
}

type {{.Type}}Repository interface {
	Create(ctx context.Context, e *{{.Type}}) error
	GetByID(ctx context.Context, id string) (*{{.Type}}, error)
	List(ctx context.Context) ([]*{{.Type}}, error)
}
`

const usecaseResourceTemplate = `package usecase

import (
	"context"

	"{{.Module}}/internal/domain"
)

type {{.Type}}Usecase struct {
	repo domain.{{.Type}}Repository
}

func New{{.Type}}Usecase(repo domain.{{.Type}}Repository) *{{.Type}}Usecase {
	return &{{.Type}}Usecase{repo: repo}
}

func (u *{{.Type}}Usecase) Create(ctx context.Context, name string) (*domain.{{.Type}}, error) {
	if name == "" {
		return nil, domain.ErrInvalid
	}
	e := &domain.{{.Type}}{Name: name}
	if err := u.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (u *{{.Type}}Usecase) GetByID(ctx context.Context, id string) (*domain.{{.Type}}, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *{{.Type}}Usecase) List(ctx context.Context) ([]*domain.{{.Type}}, error) {
	return u.repo.List(ctx)
}
`

const repositoryResourceTemplate = `package postgres

import (
	"context"
{{if .HasPostgres}}	"errors"
{{else}}	"sync"
{{end}}

	"github.com/google/uuid"
{{if .HasPostgres}}	"gorm.io/gorm"
{{end}}

	"{{.Module}}/internal/domain"
)

{{if .HasPostgres}}type {{.Type}}Repository struct {
	db *gorm.DB
}

func New{{.Type}}Repository(db *gorm.DB) *{{.Type}}Repository {
	return &{{.Type}}Repository{db: db}
}

func (r *{{.Type}}Repository) Create(ctx context.Context, e *domain.{{.Type}}) error {
	e.ID = uuid.NewString()
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *{{.Type}}Repository) GetByID(ctx context.Context, id string) (*domain.{{.Type}}, error) {
	var e domain.{{.Type}}
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *{{.Type}}Repository) List(ctx context.Context) ([]*domain.{{.Type}}, error) {
	var list []*domain.{{.Type}}
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
{{else}}type InMemory{{.Type}}Repository struct {
	mu   sync.RWMutex
	data map[string]*domain.{{.Type}}
}

func NewInMemory{{.Type}}Repository() *InMemory{{.Type}}Repository {
	return &InMemory{{.Type}}Repository{data: make(map[string]*domain.{{.Type}})}
}

func (r *InMemory{{.Type}}Repository) Create(ctx context.Context, e *domain.{{.Type}}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	e.ID = uuid.NewString()
	r.data[e.ID] = e
	return nil
}

func (r *InMemory{{.Type}}Repository) GetByID(ctx context.Context, id string) (*domain.{{.Type}}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return e, nil
}

func (r *InMemory{{.Type}}Repository) List(ctx context.Context) ([]*domain.{{.Type}}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*domain.{{.Type}}, 0, len(r.data))
	for _, e := range r.data {
		out = append(out, e)
	}
	return out, nil
}
{{end}}`

const handlerResourceTemplate = `package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"{{.Module}}/internal/domain"
	"{{.Module}}/internal/usecase"
)

type {{.Type}}Handler struct {
	usecase *usecase.{{.Type}}Usecase
}

func New{{.Type}}Handler(u *usecase.{{.Type}}Usecase) *{{.Type}}Handler {
	return &{{.Type}}Handler{usecase: u}
}

type create{{.Type}}Request struct {
	Name string ` + "`json:\"name\" binding:\"required\"`" + `
}

func (h *{{.Type}}Handler) Create(c *gin.Context) {
	var req create{{.Type}}Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	e, err := h.usecase.Create(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, e)
}

func (h *{{.Type}}Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	e, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, e)
}

func (h *{{.Type}}Handler) List(c *gin.Context) {
	list, err := h.usecase.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}
`
