# kl-go-cli

CLI untuk scaffold Go microservice dengan clean architecture (domain → usecase → delivery/repository), siap dipakai buat REST API berbasis Gin atau gRPC dengan opsi Gorm/Postgres dan SQL migration.

## Install

**Opsi 1 — go install (butuh Go 1.22+)**
\`\`\`bash
go install github.com/est57/kl-go-cli@latest
\`\`\`
Pastikan `$(go env GOPATH)/bin` ada di `$PATH`.

**Opsi 2 — clone & build manual**
\`\`\`bash
git clone https://github.com/est57/kl-go-cli.git
cd kl-go-cli
go build -o kl-go-cli .
\`\`\`

## Usage

**Mode interaktif** (paling gampang, tinggal jawab pertanyaan):
\`\`\`bash
kl-go-cli new
\`\`\`

**Mode langsung** (satu baris, cocok buat scripting/CI):
\`\`\`bash
kl-go-cli new order-service -module=github.com/kodelokal/order-service -port=8081 -tidy -git
kl-go-cli new simple-service -db=none
kl-go-cli new inventory-service -transport=both -grpc-port=9091
\`\`\`

| Flag | Fungsi | Default |
|---|---|---|
| `-module` | go module path | `github.com/kodelokal/<nama-service>` |
| `-port` | port default service | `8080` |
| `-transport` | transport scaffold (`http`, `grpc`, atau `both`) | `http` |
| `-grpc-port` | port default gRPC service | `9090` |
| `-db` | database scaffold (`postgres` atau `none`) | `postgres` |
| `-out` | nama folder output | sama dengan nama service |
| `-tidy` | otomatis jalanin `go mod tidy` setelah generate | off |
| `-git` | otomatis `git init` + commit pertama setelah generate | off |

Cek versi: `kl-go-cli -v`
Bantuan: `kl-go-cli -h`

## Roadmap

- [ ] `kl-go-cli add handler <name>` — nambah resource baru ke service yang sudah ada
- [ ] Generate custom gRPC service/proto untuk resource baru
