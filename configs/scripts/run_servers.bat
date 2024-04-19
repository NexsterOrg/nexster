@echo off

set "PROJECT_ROOT_DIR=E:\nexster-runtime\nexster"

cd /d "%PROJECT_ROOT_DIR%"

if not exist "%PROJECT_ROOT_DIR%\logs" (
    mkdir "%PROJECT_ROOT_DIR%\logs"
    echo logs directory created.
) else (
    echo logs directory already exists.
)

rem Run content server
cd content\cmd
start "" /B go run main.go > "%PROJECT_ROOT_DIR%\logs\content_server.log" 2>&1

rem Run search server
cd ..\..\search\cmd
start "" /B go run main.go > "%PROJECT_ROOT_DIR%\logs\search_server.log" 2>&1

rem Run space server
cd ..\..\space\cmd
start "" /B go run main.go > "%PROJECT_ROOT_DIR%\logs\space_server.log" 2>&1

rem Run timeline server
cd ..\..\timeline\cmd
start "" /B go run main.go > "%PROJECT_ROOT_DIR%\logs\timeline_server.log" 2>&1

rem Run usrmgmt server
cd ..\..\usrmgmt\cmd
start "" /B go run main.go > "%PROJECT_ROOT_DIR%\logs\usrmgmt_server.log" 2>&1

rem Run boarding_finder server
cd ..\..\boarding_finder\cmd
start "" /B go run main.go > "%PROJECT_ROOT_DIR%\logs\bdFinder_server.log" 2>&1

cd /d "%PROJECT_ROOT_DIR%\configs\scripts"

echo --done--
