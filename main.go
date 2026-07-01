package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/kodelokal/kl-go-cli/internal/generator"
)

const usage = `kl-go-cli - scaffold Go microservice dengan clean architecture

Usage:
  kl-go-cli new <service-name> [flags]

Flags:
  -module string   go module path (default: github.com/kodelokal/<service-name>)
  -port string     port default service (default: "8080")
  -out string      output directory (default: nama service)

Contoh:
  kl-go-cli new user-service
  kl-go-cli new order-service -module=github.com/tribina/order-service -port=8081
`

func main() {
	if len(os.Args) < 2 || os.Args[1] != "new" {
		fmt.Print(usage)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("new", flag.ExitOnError)
	module := fs.String("module", "", "go module path")
	port := fs.String("port", "8080", "default service port")
	out := fs.String("out", "", "output directory")

	if len(os.Args) < 3 {
		fmt.Print(usage)
		os.Exit(1)
	}
	serviceName := os.Args[2]
	if err := fs.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	if !validServiceName(serviceName) {
		fmt.Fprintln(os.Stderr, "nama service cuma boleh huruf kecil, angka, dan dash. contoh: user-service")
		os.Exit(1)
	}

	pkgName := toPackageName(serviceName)

	if *module == "" {
		*module = "github.com/kodelokal/" + serviceName
	}
	if *out == "" {
		*out = serviceName
	}

	data := generator.Data{
		ServiceName: serviceName,
		PackageName: pkgName,
		Module:      *module,
		Port:        *port,
	}

	if err := generator.GenerateService(*out, data); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	fmt.Printf(`
✅ Service %q berhasil dibuat di ./%s

Next steps:
  cd %s
  go mod tidy
  make run

Health check: curl http://localhost:%s/health
`, serviceName, *out, *out, *port)
}

var serviceNameRe = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func validServiceName(s string) bool {
	return serviceNameRe.MatchString(s)
}

func toPackageName(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}
