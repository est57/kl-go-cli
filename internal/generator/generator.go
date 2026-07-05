package generator

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed all:templates
var templatesFS embed.FS

const templatesRoot = "templates"

// Data adalah nilai yang di-inject ke setiap template saat generate.
type Data struct {
	ServiceName string // nama tampilan, misal "user-service"
	PackageName string // nama folder/binary, snake/kebab-safe
	Module      string // go module path, misal github.com/kodelokal/user-service
	Port        string
	Database    string // postgres atau none
	Transport   string // http, grpc, atau both
	GRPCPort    string
}

func (d Data) HasPostgres() bool {
	return d.Database == "" || d.Database == "postgres"
}

func (d Data) HasHTTP() bool {
	return d.Transport == "" || d.Transport == "http" || d.Transport == "both"
}

func (d Data) HasGRPC() bool {
	return d.Transport == "grpc" || d.Transport == "both"
}

// GenerateService membuat skeleton microservice baru di outDir.
func GenerateService(outDir string, data Data) error {
	if _, err := os.Stat(outDir); err == nil {
		return fmt.Errorf("folder %q sudah ada, hapus dulu atau pilih nama lain", outDir)
	}

	return fs.WalkDir(templatesFS, templatesRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == templatesRoot {
			return nil
		}

		rel, err := filepath.Rel(templatesRoot, path)
		if err != nil {
			return err
		}
		rel = strings.TrimSuffix(rel, ".tmpl")
		if !data.HasPostgres() && isPostgresOnlyTemplate(rel) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !data.HasHTTP() && isHTTPOnlyTemplate(rel) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !data.HasGRPC() && isGRPCOnlyTemplate(rel) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		destPath := filepath.Join(outDir, rel)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}

		raw, err := templatesFS.ReadFile(path)
		if err != nil {
			return err
		}

		tmpl, err := template.New(d.Name()).Parse(string(raw))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", path, err)
		}

		f, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer f.Close()

		if err := tmpl.Execute(f, data); err != nil {
			return fmt.Errorf("render template %s: %w", path, err)
		}

		return nil
	})
}

func isPostgresOnlyTemplate(rel string) bool {
	rel = filepath.ToSlash(rel)
	postgresOnlyPrefixes := []string{
		"cmd/migrate/",
		"cmd/seed/",
		"internal/infrastructure/database/",
		"migrations/",
	}
	for _, prefix := range postgresOnlyPrefixes {
		if strings.HasPrefix(rel, prefix) {
			return true
		}
	}
	return false
}

func isHTTPOnlyTemplate(rel string) bool {
	return strings.HasPrefix(filepath.ToSlash(rel), "internal/delivery/http/")
}

func isGRPCOnlyTemplate(rel string) bool {
	return strings.HasPrefix(filepath.ToSlash(rel), "internal/delivery/grpc/")
}
