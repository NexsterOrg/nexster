@echo off

echo Stopping servers

REM Define an array of ports
set ports=8000 8001 8002 8003 8004 8005

REM Loop over each port and stop servers
for %%p in (%ports%) do (
    echo Stopping server running on port %%p...
    for /f "tokens=5" %%i in ('netstat -aon ^| findstr /r /c:":%%p[ ]*"') do (
        taskkill /F /PID %%i > nul 2>&1
    )
    echo Server on port %%p stopped.
)

echo All servers stopped successfully.