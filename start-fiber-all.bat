@echo off
echo Starting PDF AI Summarizer with Fiber Backend...
echo.

REM Start PostgreSQL (if using local installation)
echo Starting PostgreSQL...
net start postgresql-x64-14 2>nul
if %errorlevel% neq 0 (
    echo PostgreSQL service not found or already running
)

REM Start Python AI Service
echo Starting Python AI Service...
start "Python AI" cmd /k "cd pdf-ai-summarizer && python main.py"

REM Wait a bit for Python service to start
timeout /t 3 /nobreak >nul

REM Start Fiber Backend
echo Starting Fiber Backend...
start "Fiber Backend" cmd /k "cd backend-fiber && go run cmd/main.go"

REM Wait a bit for backend to start
timeout /t 3 /nobreak >nul

REM Start Frontend
echo Starting Frontend...
start "Frontend" cmd /k "cd pdf-ai-frontend && npm run dev"

echo.
echo All services are starting...
echo - Python AI: http://localhost:8000
echo - Fiber Backend: http://localhost:8080
echo - Frontend: http://localhost:3000
echo.
echo Press any key to exit...
pause >nul