@echo off
powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File "%~dp0go.ps1" %*
exit /b %ERRORLEVEL%
