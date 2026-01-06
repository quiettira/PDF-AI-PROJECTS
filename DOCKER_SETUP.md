# Docker Setup untuk PDF AI Summarizer

## Arsitektur Container

Aplikasi ini menggunakan **4 container terpisah**:

1. **Frontend (Next.js)** - Port 3000
2. **Backend (Go Fiber)** - Port 8080  
3. **Python AI Service** - Port 8000
4. **PostgreSQL Database** - Port 5432

## Konfigurasi yang Sudah Diperbaiki

### ✅ Docker Compose
- Semua 4 service sudah dikonfigurasi dengan benar
- Network internal menggunakan service names (db, python-service, backend, frontend)
- Volume untuk uploads dan database persistence
- Environment variables sudah disesuaikan untuk container networking

### ✅ Dockerfiles
- **Backend**: Multi-stage build dengan Go 1.21, optimized binary
- **Frontend**: Multi-stage build dengan standalone output untuk production
- **Python**: Slim image dengan health check dan non-root user
- **Database**: PostgreSQL 15 Alpine (official image)

### ✅ Next.js Configuration
- Added `output: 'standalone'` untuk Docker deployment
- Optimized untuk production build

### ✅ Environment Variables
- Backend menggunakan `DB_HOST=db` (bukan localhost)
- Python service URL: `http://python-service:8000`
- CORS sudah dikonfigurasi untuk container networking

## Cara Menjalankan

### 1. Persiapan
```bash
# Copy environment file
copy .env.example .env

# Edit .env dan isi GEMINI_API_KEY
notepad .env
```

### 2. Build & Run
```bash
# Build semua container
docker-build.bat

# Jalankan semua service
docker-run.bat

# Stop semua service
docker-stop.bat
```

### 3. Manual Commands
```bash
# Build
docker-compose build --no-cache

# Run detached
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

## Service URLs

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Python AI**: http://localhost:8000
- **Database**: localhost:5432

## Health Checks

- Python service memiliki health check endpoint: `/health`
- Semua service menggunakan `restart: unless-stopped`

## Troubleshooting

### Docker Desktop Not Running
```bash
# Start Docker Desktop manually
"C:\Program Files\Docker\Docker\Docker Desktop.exe"

# Wait for Docker to be ready
docker --version
docker ps
```

### Build Errors
```bash
# Clean build
docker-compose build --no-cache --pull

# Check individual service
docker build -t test-backend ./backend-fiber
docker build -t test-frontend ./pdf-ai-frontend  
docker build -t test-python ./pdf-ai-summarizer
```

### Network Issues
- Pastikan semua service menggunakan service names dalam environment variables
- Backend: `DB_HOST=db`, `PYTHON_API_URL=http://python-service:8000`
- Frontend: `NEXT_PUBLIC_API_URL=http://localhost:8080` (dari browser)

## Keuntungan Containerization

1. **Isolation**: Setiap service terisolasi dengan dependencies sendiri
2. **Scalability**: Mudah scale individual services
3. **Portability**: Berjalan sama di semua environment
4. **Development**: Consistent environment untuk semua developer
5. **Production Ready**: Optimized builds dengan multi-stage Dockerfiles

## File Structure
```
├── docker-compose.yml          # Orchestration
├── backend-fiber/
│   └── Dockerfile             # Go Fiber backend
├── pdf-ai-frontend/
│   └── Dockerfile             # Next.js frontend
├── pdf-ai-summarizer/
│   └── Dockerfile             # Python FastAPI
├── docker-build.bat           # Build script
├── docker-run.bat             # Run script
└── docker-stop.bat            # Stop script
```

Semua konfigurasi sudah benar dan siap untuk production deployment!