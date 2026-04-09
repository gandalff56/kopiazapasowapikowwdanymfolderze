@echo off
taskkill /IM FolderBackupSilent.exe /F 2>nul
taskkill /IM FolderBackup.exe /F 2>nul
echo.
echo Backup stopped.
echo.
pause
