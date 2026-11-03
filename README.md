# be-printy

Backend service untuk **Printy**, dibangun dengan Go menggunakan framework [Gin](https://github.com/gin-gonic/gin), PostgreSQL (via GORM), dan Redis.

## Tech Stack

| Kebutuhan        | Library                                   |
| ---------------- | ----------------------------------------- |
| Bahasa           | Go 1.27.3                                 |
| HTTP Framework   | `github.com/gin-gonic/gin` v1.12.0        |
| ORM              | `gorm.io/gorm` v1.31.2                    |
| Database Driver  | `gorm.io/driver/postgres` v1.6.3          |
| Cache            | `github.com/redis/go-redis/v9` v9.23.0    |
| Config           | `github.com/spf13/viper` v1.21.0          |
| Logger           | `github.com/sirupsen/logrus` v1.10.2      |
| UUID (Request ID)| `github.com/google/uuid` v1.6.0           |

## Struktur Project

```
be-printy/
├── main.go                  # Entry point aplikasi
├── app/
│   ├── handler/             # HTTP handler (parsing request & response)
│   ├── usecase/             # Business logic / orkestrasi alur
│   ├── service/             # Service layer (logic pendukung)
│   ├── repository/          # Akses data (query database)
│   └── resource/            # Inisialisasi koneksi (PostgreSQL, Redis)
├── config/                  # Loader & struct konfigurasi (viper)
├── files/config/config.yml  # File konfigurasi aplikasi
├── helper/                  # Helper umum (request ID, response)
├── infrastructure/
│   ├── log/                 # Setup logger (logrus)
│   └── migrations/          # File migrasi database
├── middleware/              # Middleware Gin (mis. request middleware)
├── models/                  # Model & struct response
├── routes/                  # Definisi route
└── testing/                 # Test
```

### Alur Layer

```
Request → Middleware → Route → Handler → Usecase → Service → Repository → Database
```

## Prasyarat

- [Go](https://go.dev/dl/) 1.27.3 atau lebih baru
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/)

## Instalasi

```bash
# Clone repository
git clone <url-repository>
cd be-printy

# Install dependency
go mod tidy
```

## Konfigurasi

Konfigurasi dibaca dari `files/config/config.yml`. Sesuaikan dengan environment kamu. Contoh (sesuaikan dengan struct di `config/type.go`):

```yaml
app:
  name: be-printy
  port: 8080
  env: development

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  name: printy
  sslmode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
```

> **Catatan:** Jangan commit kredensial asli ke repository. Gunakan file config terpisah per environment atau environment variable.

## Menjalankan Aplikasi

```bash
# Development
go run main.go

# Build binary
go build -o bin/be-printy main.go
./bin/be-printy
```

Server akan berjalan di port yang diatur pada konfigurasi (default contoh: `8080`).

## Testing

```bash
go test ./...
```

## Fitur Utama

- Arsitektur berlapis (handler → usecase → service → repository)
- Koneksi PostgreSQL dengan GORM dan Redis
- Konfigurasi terpusat dengan Viper
- Structured logging dengan Logrus
- Request ID tracing menggunakan UUID untuk setiap request
- Format response yang konsisten (`helper/response.go` dan `models/response.go`)

## Kontribusi

1. Fork repository ini
2. Buat branch fitur: `git checkout -b feat/nama-fitur`
3. Commit perubahan: `git commit -m "feat: deskripsi singkat"`
4. Push ke branch: `git push origin feat/nama-fitur`
5. Buat Pull Request

## Lisensi

Tentukan lisensi project di sini (mis. MIT).