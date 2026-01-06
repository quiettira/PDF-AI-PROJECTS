@echo off
echo ========================================
echo 🗄️ Setting up PostgreSQL Database
echo ========================================

echo.
echo Creating database 'pdf_summarizer'...
createdb -h localhost -U postgres pdf_summarizer

if %errorlevel% equ 0 (
    echo ✅ Database created successfully!
) else (
    echo ⚠️ Database might already exist or PostgreSQL not accessible
)

echo.
echo 🔧 Database will be auto-migrated when Go backend starts
echo ========================================
pause