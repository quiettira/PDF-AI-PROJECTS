@echo off
echo ========================================
echo 🚀 DOCKER ALL COMPLETE - PDF AI SUMMARIZER
echo ========================================

echo Step 1: Cleanup everything...
echo Stopping all containers...
docker stop $(docker ps -aq) 2>nul
echo Removing all containers...
docker rm $(docker ps -aq) 2>nul
echo Cleaning volumes and networks...
docker volume prune -f
docker network prune -f
docker system prune -f

echo Step 2: Building all services fresh...
docker-compose build --no-cache

echo Step 3: Starting all services...
docker-compose up -d

echo Step 4: Waiting for all services to be ready...
echo Please wait 60 seconds for all services to initialize...
timeout /t 60

echo Step 5: Health checks...
echo.
echo [1/4] Database Health Check...
docker-compose exec -T db pg_isready -U postgres -d pdf_summarizer && echo ✅ Database: READY || echo ❌ Database: FAILED

echo [2/4] Python AI Service Health Check...
curl -f http://localhost:8000/health 2>nul && echo ✅ Python AI: READY || echo ❌ Python AI: FAILED

echo [3/4] Go Backend Health Check...
curl -f http://localhost:8080/health 2>nul && echo ✅ Go Backend: READY || echo ❌ Go Backend: FAILED

echo [4/4] Frontend Health Check...
curl -f http://localhost:3000 2>nul && echo ✅ Frontend: READY || echo ❌ Frontend: FAILED

echo.
echo Step 6: Service Status...
docker-compose ps

echo.
echo ========================================
echo 🎯 ALL SERVICES RUNNING!
echo ========================================
echo.
echo 🌐 Access Points:
echo - Frontend: http://localhost:3000
echo - Backend API: http://localhost:8080
echo - Python AI API: http://localhost:8000
echo - Database: localhost:5432
echo.
echo 📋 Test Steps:
echo 1. Open http://localhost:3000
echo 2. Upload a PDF file
echo 3. Select summarization style
echo 4. Click "Summarize"
echo 5. Verify result is saved to database
echo.
echo 🔧 Troubleshooting:
echo - If any service failed, run: docker-compose logs [service-name]
echo - To restart: docker-compose restart [service-name]
echo - To stop all: STOP-ALL.bat
echo.
echo ========================================

echo Opening browser...
start http://localhost:3000

pause