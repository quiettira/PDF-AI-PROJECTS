@echo off
echo ========================================
echo 🚀 Starting PDF AI Summarizer Services
echo ========================================

echo.
echo 📊 Checking PostgreSQL...
pg_isready -h localhost -p 5432
if %errorlevel% neq 0 (
    echo ❌ PostgreSQL is not running! Please start PostgreSQL first.
    echo    - Install PostgreSQL if not installed
    echo    - Create database: createdb pdf_summarizer
    pause
    exit /b 1
)
echo ✅ PostgreSQL is running

echo.
echo 🐍 Starting Python AI Service...
start "Python AI Service" cmd /k "cd pdf-ai-summarizer && python -m venv venv && venv\Scripts\activate && pip install -r requirements.txt && python main.py"

echo.
echo ⏳ Waiting for Python service to start...
timeout /t 5 /nobreak > nul

echo.
echo 🔧 Starting Fiber Backend...
start "Fiber Backend" cmd /k "cd backend-fiber && go mod tidy && go run cmd/main.go"

echo.
echo ⏳ Waiting for Go backend to start...
timeout /t 3 /nobreak > nul

echo.
echo 🌐 Starting Next.js Frontend...
start "Next.js Frontend" cmd /k "cd pdf-ai-frontend && npm install && npm run dev"

echo.
echo ========================================
echo ✅ All services are starting!
echo ========================================
echo.
echo 📋 Service URLs:
echo   Frontend: http://localhost:3000
echo   Go API:   http://localhost:8080
echo   Python:   http://localhost:8000
echo   Database: localhost:5432
echo.
echo 🔧 To stop services, close the terminal windows
echo ========================================
pause