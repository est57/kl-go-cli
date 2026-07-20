# kl-go-cli

`kl-go-cli` adalah alat baris perintah untuk membuat struktur awal layanan backend menggunakan Go.

Project yang dibuat mengikuti pola clean architecture dan dapat menggunakan REST API, gRPC, PostgreSQL, migration, dan seed data sesuai pilihan saat membuat service.

## Tujuan Project

Project ini dibuat untuk mempercepat pembuatan service backend baru. Dengan satu perintah, pengguna mendapatkan struktur project Go yang sudah rapi dan siap dikembangkan.

`kl-go-cli` bukan framework runtime. Hasil dari perintah ini adalah project Go biasa yang dapat diedit, diuji, dikomit, dan dideploy seperti project Go lainnya.

## Fitur Utama

- Membuat struktur project Go dengan pola clean architecture.
- Membuat REST API menggunakan Gin.
- Mendukung gRPC health service.
- Mendukung PostgreSQL dengan Gorm.
- Menyediakan migration runner bawaan.
- Menyediakan seed command untuk data awal pengembangan lokal.
- Menyediakan Dockerfile dan docker-compose.
- Menyediakan Makefile untuk perintah umum.
- Dapat menambahkan skeleton HTTP handler baru ke service yang sudah dibuat.

## Instalasi

Install versi terbaru:

```bash
go install github.com/est57/kl-go-cli@latest
```

Install versi tertentu:

```bash
go install github.com/est57/kl-go-cli@v0.6.0
```

Pastikan folder `$(go env GOPATH)/bin` sudah ada di `PATH`.

Periksa versi:

```bash
kl-go-cli -v
```

## Quickstart

Buat service baru:

```bash
kl-go-cli new payment-service -module=github.com/acme/payment-service -tidy
cd payment-service
make run
```

Periksa health check:

```bash
curl http://localhost:8080/health
```

Jika service menggunakan PostgreSQL, jalankan database dan migration terlebih dahulu:

```bash
docker compose up -d db
make migrate-up
make seed
make run
```

## Contoh Penggunaan

Membuat REST API dengan PostgreSQL:

```bash
kl-go-cli new order-service -module=github.com/acme/order-service -db=postgres -tidy
```

Membuat REST API tanpa database:

```bash
kl-go-cli new webhook-service -db=none -tidy
```

Membuat service gRPC tanpa database:

```bash
kl-go-cli new inventory-rpc -transport=grpc -grpc-port=9090 -db=none -tidy
```

Membuat service yang menjalankan HTTP dan gRPC dalam satu binary:

```bash
kl-go-cli new payment-service -transport=both -port=8080 -grpc-port=9090 -db=postgres -tidy
```

Menambahkan HTTP handler baru ke service yang sudah dibuat:

```bash
cd payment-service
kl-go-cli add handler customer
```

## Contoh Output Perintah

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

## Struktur Project yang Dibuat

Contoh struktur untuk service dengan `-transport=both -db=postgres`:

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

Jika menggunakan `-db=none`, folder database, migration, migrate command, dan seed command tidak dibuat.

Jika menggunakan `-transport=grpc`, folder HTTP tidak dibuat.

## Menambahkan Handler Baru

Jalankan perintah ini dari root project hasil generate:

```bash
kl-go-cli add handler customer
```

Perintah tersebut membuat file berikut:

```text
internal/domain/customer.go
internal/usecase/customer_usecase.go
internal/repository/postgres/customer_repo.go
internal/delivery/http/handler/customer_handler.go
```

CLI akan mencoba memperbarui `internal/delivery/http/router.go` secara otomatis.

Jika struktur router sudah banyak berubah dan tidak lagi mengikuti pola awal, CLI tidak akan memaksa perubahan. Dalam kondisi tersebut, CLI akan menampilkan snippet wiring yang dapat ditambahkan secara manual.

Untuk project PostgreSQL, perintah ini juga membuat migration placeholder:

```text
migrations/000002_create_customers.up.sql
migrations/000002_create_customers.down.sql
```

Isi migration hanya berupa komentar `TODO`. Schema tabel tetap harus disesuaikan dengan kebutuhan domain sebelum menjalankan:

```bash
make migrate-up
```

## Pilihan Database dan Transport

| Opsi | Kapan digunakan | Keterangan |
|---|---|---|
| `-db=postgres` | Service membutuhkan database sejak awal. | Membuat konfigurasi PostgreSQL, repository Gorm, migration, seed command, dan docker-compose untuk database. |
| `-db=none` | Service belum membutuhkan database. | Output lebih sederhana dan dependency lebih sedikit. |
| `-transport=http` | Service menyediakan REST API. | Membuat router, handler, dan middleware HTTP menggunakan Gin. |
| `-transport=grpc` | Service digunakan untuk komunikasi internal antar-service. | Membuat gRPC server dengan health service. |
| `-transport=both` | Service membutuhkan REST API dan gRPC sekaligus. | Menjalankan HTTP dan gRPC dalam satu binary. |

## Daftar Flag

| Flag | Fungsi | Default |
|---|---|---|
| `-module` | Go module path. | `github.com/kodelokal/<nama-service>` |
| `-port` | Port HTTP. | `8080` |
| `-transport` | Jenis transport: `http`, `grpc`, atau `both`. | `http` |
| `-grpc-port` | Port gRPC. | `9090` |
| `-db` | Jenis database: `postgres` atau `none`. | `postgres` |
| `-out` | Folder output. | Sama dengan nama service. |
| `-tidy` | Menjalankan `go mod tidy` setelah project dibuat. | Tidak aktif. |
| `-git` | Menjalankan `git init` dan membuat commit pertama. | Tidak aktif. |

## Build Manual dari Source

```bash
git clone https://github.com/est57/kl-go-cli.git
cd kl-go-cli
go build -o kl-go-cli .
```

Periksa bantuan command:

```bash
kl-go-cli -h
```

## Target Pengguna

Project ini cocok untuk:

- developer backend Go;
- tim yang sering membuat microservice baru;
- project internal tools;
- service REST atau gRPC skala kecil sampai menengah;
- tim yang ingin struktur project backend lebih konsisten.

Project ini mungkin kurang cocok untuk:

- aplikasi yang cukup dibuat dalam satu file `main.go`;
- tim yang sudah memiliki template internal sendiri;
- pengguna yang mencari framework lengkap seperti Laravel, Rails, atau NestJS.

## Roadmap

- [ ] Field generator untuk `kl-go-cli add handler <name>`.
- [ ] Generate custom gRPC service/proto untuk resource baru.
- [ ] Template auth/JWT opsional.
