# PDF AI Summarizer

Aplikasi web untuk upload PDF, menyimpan metadata/file, lalu menghasilkan ringkasan otomatis.

## Alur Singkat

1. User upload PDF lewat **Frontend (Next.js)**
2. Frontend kirim request ke **Backend (Go Fiber)** untuk simpan file + data ke **PostgreSQL**
3. Backend memanggil **Python AI Service (FastAPI)** untuk proses summarization/export
4. Hasil ringkasan disimpan/ditampilkan kembali ke frontend

## Tech Stack

- **Frontend**: Next.js (App Router)
- **Backend**: Go + Fiber (`backend-fiber/`)
- **AI Service**: Python + FastAPI (`pdf-ai-summarizer/`)
- **Database**: PostgreSQL

Catatan:

- Python service akan pakai **Gemini** kalau `GEMINI_API_KEY` tersedia.
- Kalau tidak ada API key, service tetap jalan dengan provider `mock`.

## Struktur Proyek

- `backend-fiber/` - Go Fiber API + DB auto-migrate + file storage
- `pdf-ai-summarizer/` - Python FastAPI service untuk summarization/export
- `pdf-ai-frontend/` - Next.js UI

## Prasyarat

- Go 1.21+
- Python 3.9+
- Node.js 18+
- PostgreSQL
- (Opsional) Gemini API Key: https://aistudio.google.com

## Quick Start (Windows)

1. Buat database `pdf_summarizer`:

```bash
createdb pdf_summarizer
```

Atau jalankan:

```bash
setup-db.bat
```

2. Jalankan semua service (frontend + backend + python):

```bash
start-all.bat
```

Service URLs:

- Frontend: http://localhost:3000
- Go API: http://localhost:8080
- Python AI: http://localhost:8000

## Konfigurasi Environment

Gunakan `.env` sesuai kebutuhan per service.

### Backend Fiber

Backend membaca env berikut (default ada di `backend-fiber/internal/config/config.go`):

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=pdf_summarizer

PYTHON_API_URL=http://localhost:8000/summarize

MAX_FILE_SIZE=10485760
UPLOAD_DIR=./uploads
```

### Frontend Next.js

```env
NEXT_PUBLIC_GO_API_BASE_URL=http://localhost:8080
```

### Python AI Service

```env
GEMINI_API_KEY=your_gemini_api_key
CORS_ORIGINS=http://localhost:3000
```

## Menjalankan Manual (tanpa .bat)

1. Python AI Service

```bash
cd pdf-ai-summarizer
python -m venv venv
venv\Scripts\activate
pip install -r requirements.txt
python main.py
```

2. Backend Fiber

```bash
cd backend-fiber
go mod tidy
go run cmd/main.go
```

3. Frontend

```bash
cd pdf-ai-frontend
npm install
npm run dev
```

## Fitur

- Upload PDF (default max 10MB)
- Ringkas otomatis + pilihan style: `standard`, `executive`, `bullets`, `detailed`
- List/detail/download/delete PDF
- Riwayat ringkasan
- Export ringkasan (TXT/PDF) via Python service

## API (Ringkas)

- Backend Go berjalan di `http://localhost:8080`
- Endpoint yang paling sering dipakai:
  - `POST /upload` (query `style=standard|executive|bullets|detailed`)
  - `GET /simple-pdfs` (list)
  - `GET /pdf/:id/download` (download)
  - `GET /history` (history)
  - `GET /health` (cek service)

## Troubleshooting

### Python service tidak bisa diakses

- Pastikan Python berjalan di port 8000
- Pastikan `PYTHON_API_URL` mengarah ke `http://localhost:8000/summarize`

### PostgreSQL tidak jalan

- Pastikan service Postgres aktif
- Pastikan database `pdf_summarizer` sudah dibuat

### Frontend tidak bisa fetch API

- Pastikan `NEXT_PUBLIC_GO_API_BASE_URL` sesuai URL backend
- Pastikan backend Go running di `http://localhost:8080`

---

Last updated: 2025-12-11
