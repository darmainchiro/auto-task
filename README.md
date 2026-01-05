# SRS Automation API

API untuk automatisasi pembuatan Software Requirements Specification (SRS) berdasarkan dokumen BRD atau dokumen lainnya menggunakan AI.

## Fitur

- Upload dokumen BRD/dokumen lainnya
- Ekstraksi konten dokumen menggunakan AI (Google Gemini)
- Generate SRS otomatis dari dokumen BRD
- CRUD operations untuk dokumen dan SRS
- Clean Architecture dengan Separation of Concerns

## Struktur Proyek

```
project-root/
├── cmd/
│   └── main.go                 # Entry point aplikasi
├── internal/
│   ├── api/
│   │   ├── handler/           # HTTP handlers (Gin/Fiber)
│   │   └── router/            # Route configuration
│   ├── core/
│   │   ├── domain/            # Domain models/entities
│   │   ├── service/           # Business logic (Use Cases)
│   │   └── ports/             # Interfaces (Repository, External)
│   └── infra/
│       ├── database/          # Database connection
│       ├── repository/        # Repository implementations
│       └── external/          # External API integrations
└── pkg/                       # Shared utilities
```

## Prerequisites

- Go 1.21+
- PostgreSQL
- Google Gemini API Key

## Setup

1. Clone repository
2. Copy `.env.example` ke `.env` dan sesuaikan konfigurasi:
   - Dapatkan Gemini API Key dari: https://makersuite.google.com/app/apikey
   - Setup PostgreSQL database
3. Install dependencies:
```bash
go mod download
```

4. Jalankan aplikasi:
```bash
go run cmd/main.go
```

## API Endpoints

### Documents
- `POST /api/v1/documents` - Upload dokumen
- `GET /api/v1/documents` - List semua dokumen
- `GET /api/v1/documents/:id` - Detail dokumen
- `POST /api/v1/documents/:id/process` - Proses dokumen dengan AI
- `DELETE /api/v1/documents/:id` - Hapus dokumen

### SRS
- `POST /api/v1/srs` - Generate SRS dari dokumen
- `GET /api/v1/srs` - List semua SRS
- `GET /api/v1/srs/:id` - Detail SRS
- `GET /api/v1/srs/document/:documentId` - SRS berdasarkan dokumen
- `PUT /api/v1/srs/:id` - Update SRS
- `DELETE /api/v1/srs/:id` - Hapus SRS

## Contoh Penggunaan

### 1. Upload Dokumen BRD
```bash
curl -X POST http://localhost:8080/api/v1/documents \
  -F "file=@brd.pdf" \
  -F "type=BRD"
```

### 2. Proses Dokumen
```bash
curl -X POST http://localhost:8080/api/v1/documents/1/process
```

### 3. Generate SRS
```bash
curl -X POST http://localhost:8080/api/v1/srs \
  -H "Content-Type: application/json" \
  -d '{"document_id": 1, "title": "SRS untuk Aplikasi XYZ"}'
```
