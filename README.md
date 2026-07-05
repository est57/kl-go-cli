# kl-go-cli

CLI ringan untuk generate Go microservice siap jalan dengan clean architecture, Gin, optional Postgres, migration, seed, dan gRPC.

`kl-go-cli` cocok untuk backend developer atau tim yang sering membuat service baru dan ingin struktur project yang konsisten tanpa copy-paste dari service lama.

## Quickstart 3 Menit

Install versi terbaru:

```bash
go install github.com/est57/kl-go-cli@latest
```

Generate service:

```bash
kl-go-cli new payment-service -module=github.com/acme/payment-service -tidy
cd payment-service
make run
```

Cek service:

```bash
curl http://localhost:8080/health
```

Untuk service dengan Postgres local:

```bash
docker compose up -d db
make migrate-up
make seed
make run
```

## Untuk Apa?

Project ini adalah scaffolder, bukan framework runtime. Output-nya adalah project Go biasa yang bisa diedit, dipindahkan, dicommit, dan dideploy seperti service Go normal.

Yang dibuat otomatis:

- struktur clean architecture
- REST API dengan Gin
- optional gRPC health server
- optional Postgres + Gorm repository
- SQL migration runner bawaan
- seed command untuk local development
- env config loader
- Dockerfile dan docker-compose
- Makefile untuk command umum
- command untuk menambah skeleton HTTP handler baru

## Contoh Command

REST API + Postgres, cocok untuk service backend standar:

```bash
kl-go-cli new order-service -module=github.com/acme/order-service -db=postgres -tidy
```

REST API ringan tanpa database:

```bash
kl-go-cli new webhook-service -db=none -tidy
```

gRPC-only service:

```bash
kl-go-cli new inventory-rpc -transport=grpc -grpc-port=9090 -db=none -tidy
```

HTTP + gRPC dalam satu service:

```bash
kl-go-cli new payment-service -transport=both -port=8080 -grpc-port=9090 -db=postgres -tidy
```

Tambah skeleton resource HTTP di service yang sudah dibuat:

```bash
cd payment-service
kl-go-cli add handler customer
```

## Contoh Terminal Demo

```bash
$ kl-go-cli new payment-service -module=github.com/acme/payment-service -transport=both -db=postgres -tidy

Service "payment-service" berhasil dibuat di ./payment-service

menjalankan go mod tidy...
selesai.

Next steps:
  cd payment-service
  make run

Health check: curl http://localhost:8080/health
gRPC health service: localhost:9090
```

## Struktur Output

Contoh struktur untuk `-transport=both -db=postgres`:

```text
payment-service/
├── cmd/
│   ├── api/main.go
│   ├── migrate/main.go
│   └── seed/main.go
├── internal/
│   ├── config/
│   ├── delivery/
│   │   ├── grpc/
│   │   └── http/
│   ├── domain/
│   ├── infrastructure/database/
│   ├── repository/postgres/
│   └── usecase/
├── migrations/
├── deployments/Dockerfile
├── docker-compose.yml
├── Makefile
├── .env.example
└── go.mod
```

Untuk `-db=none`, folder database, migration, migrate command, dan seed command tidak dibuat.

Untuk `-transport=grpc`, folder HTTP tidak dibuat.

## Menambah Handler Baru

Jalankan command ini dari root project hasil generate:

```bash
kl-go-cli add handler customer
```

Command ini membuat file:

```text
internal/domain/customer.go
internal/usecase/customer_usecase.go
internal/repository/postgres/customer_repo.go
internal/delivery/http/handler/customer_handler.go
```

Setelah file dibuat, CLI akan menampilkan snippet wiring untuk ditambahkan ke `internal/delivery/http/router.go`.

Catatan: untuk saat ini wiring router masih manual agar `kl-go-cli` tidak menimpa perubahan custom di router project kamu.

## Kapan Pakai Flag Tertentu?

| Flag | Pakai saat | Catatan |
|---|---|---|
| `-db=postgres` | Service butuh persistence dari awal | Default. Include Gorm, migration, seed, dan docker-compose Postgres. |
| `-db=none` | Service ringan, worker, webhook, proxy, atau belum butuh DB | Output lebih kecil dan dependency lebih sedikit. |
| `-transport=http` | Service REST API umum | Default. Include Gin router, handler, middleware. |
| `-transport=grpc` | Service internal antar backend | Include gRPC health service tanpa perlu `protoc`. |
| `-transport=both` | Service butuh REST API publik dan gRPC internal | Menjalankan HTTP + gRPC dalam satu binary. |
| `-tidy` | Ingin dependency langsung siap | Menjalankan `go mod tidy` setelah generate. |
| `-git` | Ingin output langsung jadi repo baru | Menjalankan `git init`, `git add`, dan initial commit. |

## Semua Flag

| Flag | Fungsi | Default |
|---|---|---|
| `-module` | Go module path | `github.com/kodelokal/<nama-service>` |
| `-port` | Port default HTTP service | `8080` |
| `-transport` | Transport scaffold: `http`, `grpc`, atau `both` | `http` |
| `-grpc-port` | Port default gRPC service | `9090` |
| `-db` | Database scaffold: `postgres` atau `none` | `postgres` |
| `-out` | Folder output | sama dengan nama service |
| `-tidy` | Otomatis menjalankan `go mod tidy` | off |
| `-git` | Otomatis `git init` + commit pertama | off |

## Install Manual

```bash
git clone https://github.com/est57/kl-go-cli.git
cd kl-go-cli
go build -o kl-go-cli .
```

Cek versi:

```bash
kl-go-cli -v
```

Bantuan:

```bash
kl-go-cli -h
```

## Target Pengguna

`kl-go-cli` paling berguna untuk:

- Go backend developer
- tim microservice
- project internal tools
- service REST/gRPC kecil sampai menengah
- tim yang ingin struktur service konsisten

Kurang cocok untuk:

- aplikasi sangat kecil yang cukup satu file `main.go`
- project yang sudah punya framework internal sendiri
- user yang mencari full-stack framework seperti Laravel, Rails, atau NestJS

## Roadmap

- [ ] Auto-wiring route untuk `kl-go-cli add handler <name>`
- [ ] Generate custom gRPC service/proto untuk resource baru
- [ ] Template auth/JWT opsional
