@echo off
net session >nul 2>&1
if %errorlevel% neq 0 (
    powershell -Command "Start-Process '%~f0' -Verb RunAs"
    exit /b
)

echo Stopping backup processes...
echo.
taskkill /IM FolderBackupSilent.exe /F
taskkill /IM FolderBackup.exe /F
echo.
echo Done.
echo.
pause
