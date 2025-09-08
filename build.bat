@echo off
setlocal enabledelayedexpansion

REM 设置构建时间
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "year=%dt:~0,4%"
set "month=%dt:~4,2%"
set "day=%dt:~6,2%"
set "hour=%dt:~8,2%"
set "minute=%dt:~10,2%"
set "BUILDTIME=%year%-%month%-%day% %hour%:%minute%"

REM 定义颜色
set "GREEN=[92m"
set "RESET=[0m"

REM 清理函数
:clean
    echo Cleaning build files...
    if exist server.exe del /q server.exe
    if exist client.exe del /q client.exe
    if exist server.latest del /q server.latest
    if exist client.latest del /q client.latest
    echo %GREEN%Clean done%RESET%
    goto :eof

REM 构建服务端函数
:server
    echo Building server...
    for /f "delims=" %%i in ('git describe --tags --always --dirty 2^>nul') do set "VERSION=%%i"
    if "%VERSION%"=="" set "VERSION=unknown"
    go build -ldflags="-X main.version=%VERSION%" -o server.exe cmd\server\server.go
    if errorlevel 1 (
        echo Error: Failed to build server
        exit /b 1
    )
    echo version : %VERSION%>server.latest
    echo time : %BUILDTIME%>>server.latest
    echo %GREEN%Server build done%RESET%
    goto :eof

REM 构建客户端函数
:client
    echo Building client...
    for /f "delims=" %%i in ('git describe --tags --always --dirty 2^>nul') do set "VERSION=%%i"
    if "%VERSION%"=="" set "VERSION=unknown"
    go build -ldflags="-X main.version=%VERSION%" -o client.exe cmd\client\client.go
    if errorlevel 1 (
        echo Error: Failed to build client
        exit /b 1
    )
    echo version : %VERSION%>client.latest
    echo time : %BUILDTIME%>>client.latest
    echo %GREEN%Client build done%RESET%
    goto :eof

REM 构建所有
:build
    call :clean
    call :server
    call :client
    goto :eof

REM 查看帮助
:help
    echo Usage: build.bat [target]
    echo Targets:
    echo   server    Build server binary
    echo   client    Build client binary
    echo   build     Build both server and client (default)
    echo   clean     Clean build files
    echo   help      Show this help message
    goto :eof

REM 主逻辑
if "%1"=="" goto :build
if "%1"=="server" goto :server
if "%1"=="client" goto :client
if "%1"=="build" goto :build
if "%1"=="clean" goto :clean
if "%1"=="help" goto :help

REM 如果指定了未知目标，显示帮助
if not "%1"=="" (
    echo Unknown target: %1
    goto :help
)

endlocal