@echo off
REM Script untuk menjalankan unboost franchise service di Windows
REM Pastikan environment variables sudah diset

REM Set working directory
cd /d "%~dp0"

REM Run the application
echo Starting unboost franchise service...
go run main.go

if %ERRORLEVEL% EQU 0 (
    echo Unboost franchise service completed successfully
) else (
    echo Unboost franchise service failed
    exit /b 1
) 