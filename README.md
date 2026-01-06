# PDF AI Summarizer

Aplikasi web untuk meringkas file PDF secara otomatis menggunakan AI Gemini dengan CRUD operations dan manajemen file PDF.

## 🏗️ Tech Stack

- **Frontend**: Next.js dengan React (App Router)
- **Backend**: Golang dengan Gin framework
- **AI Service**: Python FastAPI dengan Google Gemini 2.5 Flash
- **Database**: PostgreSQL

## 🚀 Panduan Instalasi Cepat

### Prasyarat
- Go 1.21+
- Python 3.9+
- Node.js 18+
- PostgreSQL 15+
- [Gemini API Key](https://aistudio.google.com)

### 1. Setup Awal
```bash
# Clone repository
# Buat database PostgreSQL
createdb pdf_summarizer
```

### 2. Konfigurasi Environment
Buat file `.env` di root project:
```env
# Backend
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=pdf_summarizer
PYTHON_API_URL=http://localhost:8000/summarize

# Python Service
GEMINI_API_KEY=your_gemini_api_key
```

### 3. Menjalankan Aplikasi
#### Backend (Golang)
```bash
cd backend-go
go mod tidy
go run .
```

#### AI Service (Python)
```bash
cd pdf-ai-summarizer
python -m venv venv
# Windows
venv\Scripts\activate
# Linux/Mac
# source venv/bin/activate

pip install -r requirements.txt
python main.py
```

#### Frontend (Next.js)
```bash
cd pdf-ai-frontend
npm install
npm run dev
```

## 🔧 Troubleshooting

### 1. Koneksi Ditolak (Connection Refused)
**Gejala**:
```
dial tcp [::1]:8001: connectex: No connection could be made
```

**Solusi**:
- Pastikan Python service berjalan di port yang benar (default: 8000)
- Periksa variabel environment `PYTHON_API_URL` di backend
- Pastikan tidak ada firewall yang memblokir koneksi

### 2. Error API Key Tidak Ditemukan
**Solusi**:
- Pastikan `GEMINI_API_KEY` sudah diatur di environment
- Restart service Python setelah mengubah environment variable

### 3. Error Database
**Solusi**:
- Pastikan PostgreSQL berjalan
- Periksa kredensial database di file `.env`
- Jalankan migrasi database jika diperlukan

## 📂 Struktur Proyek
- `backend-fiber/`: Server utama (Golang)
- `pdf-ai-summarizer/`: Layanan AI (Python)
- `pdf-ai-frontend/`: Antarmuka pengguna (Next.js)

## 🎯 Fitur
- Upload dan manajemen file PDF
- Ringkasan otomatis dengan berbagai gaya (standard, eksekutif, poin-poin)
- Ekspor hasil ringkasan (TXT, PDF)
- Riwayat ringkasan

## 📝 Catatan
- Pastikan Python service berjalan sebelum mengakses fitur ringkasan
- Batas ukuran file default: 10MB
- Format file yang didukung: PDF
python main.py

# 3. Frontend Next.js
cd pdf-ai-frontend
npm install
npm run dev
```

### 4. Akses Aplikasi
- **Frontend**: http://localhost:3000
- **Go Backend API**: http://localhost:8080
- **Python AI Service**: http://localhost:8000
- **PostgreSQL**: localhost:5432

## 📋 Fitur & Requirements

### ✅ CRUD File PDF
- **Upload PDF** dengan validasi ukuran (max 10MB)
- **View/List** semua PDF yang sudah diupload
- **Delete PDF** beserta summary dan file fisiknya
- **Download** file PDF original

### ✅ Proses Summarization
- **Auto-summarize** setelah upload PDF
- **Multiple styles** (standard, executive, bullets, detailed)
- **Language detection** otomatis
- **Process time tracking** untuk monitoring performa

### ✅ Manajemen Data
- **Database storage** untuk metadata PDF dan hasil summarization
- **File size limitation** (10MB per file)
- **Process time logging** untuk analisis performa
- **Export** hasil summarization ke TXT/PDF

### 🎯 API Endpoints

#### Go Backend (Port 8080)
- `POST /upload` - Upload PDF dan auto-summarize
- `GET /pdfs` - Ambil daftar semua PDF
- `GET /pdf/{id}` - Detail PDF dan summary
- `DELETE /pdf/{id}` - Hapus PDF dan summarynya
- `GET /pdf/{id}/download` - Download file PDF

#### Python AI Service (Port 8001)
- `POST /summarize` - Summarize PDF dengan style tertentu
- `POST /extract-text` - Extract teks dari PDF
- `GET /health` - Health check service

## 🛠️ Development Setup

### Backend Go
```bash
cd backend-go
go mod tidy
go run main.go
```

### Python AI Service
```bash
cd pdf-ai-summarizer
python -m venv venv
venv\Scripts\activate  # Windows
# source venv/bin/activate  # Linux/Mac
pip install -r requirements.txt
python main.py
```

### Frontend Next.js
```bash
cd pdf-ai-frontend
npm install
npm run dev
```

### PostgreSQL Setup
```sql
-- Buat database
CREATE DATABASE pdf_summarizer;

-- Gunakan database
\c pdf_summarizer;

-- Tabel akan dibuat otomatis oleh Go backend saat startup
```

## 🔧 Configuration

### Environment Variables (.env)
```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=pdf_summarizer

# AI Service
GEMINI_API_KEY=your_api_key_here

# File Upload Limits
MAX_FILE_SIZE=10485760  # 10MB in bytes
ALLOWED_EXTENSIONS=pdf

# API URLs
PYTHON_API_URL=http://localhost:8000
GO_API_URL=http://localhost:8080
```

## 📊 Database Schema

### pdf_files
- `id` (SERIAL PRIMARY KEY)
- `filename` (VARCHAR(255))
- `original_filename` (VARCHAR(255))
- `filepath` (TEXT)
- `filesize` (BIGINT)
- `upload_time` (TIMESTAMP)
- `created_at` (TIMESTAMP DEFAULT NOW())

### summaries
- `id` (SERIAL PRIMARY KEY)
- `pdf_id` (INT REFERENCES pdf_files(id) ON DELETE CASCADE)
- `summary_text` (TEXT)
- `summary_style` (VARCHAR(50))
- `process_time_ms` (BIGINT)
- `language_detected` (VARCHAR(10))
- `created_at` (TIMESTAMP DEFAULT NOW())

### File Size Limits
- **Maximum file size**: 10MB per PDF
- **Supported format**: PDF only
- **Storage**: Local filesystem dengan path tracking di database

## � Troubrleshooting

### Port Conflicts
```bash
# Windows - Cek port yang digunakan
netstat -an | findstr :3000
netstat -an | findstr :8080
netstat -an | findstr :8000
netstat -an | findstr :5432
```

### Database Connection Issues
```bash
# Test koneksi PostgreSQL
psql -h localhost -U postgres -d pdf_summarizer

# Check tables
\dt
\d pdf_files
\d summaries
```

### File Upload Issues
- Pastikan folder `backend-go/uploads` ada dan writable
- Check file size limit (max 10MB)
- Verify file format (hanya PDF yang diizinkan)

### CORS Issues
- Backend Go sudah include CORS middleware
- Pastikan frontend menggunakan URL yang benar (localhost:8080)

---

**Made with ❤️ using Go, Python, Next.js, and PostgreSQL**
