@echo off
echo Starting Fiber Backend...
cd /d "%~dp0"
go run cmd/main.go
pause