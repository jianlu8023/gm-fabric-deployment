@echo off
setlocal enabledelayedexpansion

REM 获取脚本所在的目录作为项目根目录
set "PROJECT_DIR=%~dp0"
REM 移除末尾的反斜杠
if "!PROJECT_DIR:~-1!"=="\" set "PROJECT_DIR=!PROJECT_DIR:~0,-1!"

REM 设置模块路径
set "MODULE_PATH=github.com/jianlu8023/golang-example"

REM 设置构建时间
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "year=%dt:~0,4%"
set "month=%dt:~4,2%"
set "day=%dt:~6,2%"
set "hour=%dt:~8,2%"
set "minute=%dt:~10,2%"
set "second=%dt:~12,2%"
set "BUILDTIME=%year%-%month%-%day% %hour%:%minute%:%second%"

REM 获取Git分支和标签信息
set "VERSION=unknown"
for /f "delims=" %%i in ('git rev-parse --abbrev-ref HEAD 2^>nul') do set "BRANCH=%%i"
if not defined BRANCH set "BRANCH=unknown"

for /f "delims=" %%i in ('git describe --tags --always --dirty 2^>nul') do set "VERSION=%%i"
if "!VERSION!"=="unknown" (
    for /f "delims=" %%i in ('git rev-parse --short HEAD 2^>nul') do set "VERSION=%%i"
)
if "!VERSION!"=="" set "VERSION=unknown"

REM 最终版本号格式：分支名-标签信息
set "FULL_VERSION=!BRANCH!-!VERSION!"

REM 设置ldflags参数，包括版本信息
set "LDFLAGS=-s -w -X \"%MODULE_PATH%/version.Version=!FULL_VERSION!\""
set "BUILD_FLAGS=-tags=jsoniter -trimpath"

echo ==========================================
echo Project: %MODULE_PATH%
echo Version: %FULL_VERSION%
echo Build Time: %BUILDTIME%
echo ==========================================

REM 根据命令行参数执行相应操作
if "%1"=="" goto :build
if /i "%1"=="server" goto :build_server
if /i "%1"=="client" goto :build_client
if /i "%1"=="build" goto :build
if /i "%1"=="clean" goto :clean
if /i "%1"=="docker" goto :build_docker
if /i "%1"=="help" goto :show_help
if /i "%1"=="-h" goto :show_help
if /i "%1"=="--help" goto :show_help

echo Unknown target: %1
goto :show_help

REM 构建服务端
:build_server
    echo [INFO] Building server...
    go build %BUILD_FLAGS% -ldflags="%LDFLAGS%" -o server.exe cmd/server/server.go
    if errorlevel 1 (
        echo [ERROR] Failed to build server
        exit /b 1
    )
    echo version: %FULL_VERSION%> server.latest
    echo time: %BUILDTIME%>> server.latest
    copy server.exe server.bin >nul 2>&1
    echo [SUCCESS] Server built successfully as server.exe (server.bin also created for compatibility)
    goto :eof

REM 构建客户端
:build_client
    echo [INFO] Building client...
    go build %BUILD_FLAGS% -ldflags="%LDFLAGS%" -o client.exe cmd/client/client.go
    if errorlevel 1 (
        echo [ERROR] Failed to build client
        exit /b 1
    )
    echo version: %FULL_VERSION%> client.latest
    echo time: %BUILDTIME%>> client.latest
    copy client.exe client.bin >nul 2>&1
    echo [SUCCESS] Client built successfully as client.exe (client.bin also created for compatibility)
    goto :eof

REM 构建所有
:build
    call :clean
    call :build_server
    call :build_client
    echo [SUCCESS] All builds completed successfully!
    goto :eof

REM 清理构建文件
:clean
    echo [INFO] Cleaning build files...
    if exist server.exe del /q server.exe >nul 2>&1
    if exist client.exe del /q client.exe >nul 2>&1
    if exist server.bin del /q server.bin >nul 2>&1
    if exist client.bin del /q client.bin >nul 2>&1
    if exist server.latest del /q server.latest >nul 2>&1
    if exist client.latest del /q client.latest >nul 2>&1
    echo [SUCCESS] Clean done
    goto :eof

REM 构建Docker镜像
:build_docker
    echo [INFO] Building Docker image...
    
    REM 获取镜像版本（基于日期时间）
    set "IMAGE_VERSION=v%year%%month%%day%%hour%%minute%"
    set "IMAGE_NAME=golang-example/ubuntu2204/app:%IMAGE_VERSION%"
    
    echo [INFO] Pulling base images...
    docker pull golang:1.22 >nul 2>&1
    docker pull ubuntu:22.04 >nul 2>&1
    
    if errorlevel 1 (
        echo [WARNING] Failed to pull base images, continuing with build...
    )
    
    echo [INFO] Building image: %IMAGE_NAME%
    docker buildx build --platform linux/amd64 --build-arg "VERSION=%FULL_VERSION%" -t "%IMAGE_NAME%" .
    
    if errorlevel 1 (
        echo [ERROR] Failed to build Docker image
        exit /b 1
    )
    
    echo [INFO] Cleaning up...
    docker rmi golang:1.22 ubuntu:22.04 >nul 2>&1
    docker builder prune -a -f >nul 2>&1
    
    echo [SUCCESS] Docker image built successfully!
    echo IMAGE NAME: %IMAGE_NAME%
    goto :eof

REM 显示帮助信息
:show_help
    echo Usage: build-enhanced.bat [target]
    echo.
    echo Targets:
    echo   server    Build server binary
    echo   client    Build client binary
    echo   build     Build both server and client ^(default^)
    echo   clean     Clean build files
    echo   docker    Build Docker image
    echo   help      Show this help message
    echo.
    echo Examples:
    echo   build-enhanced.bat          ^| Build everything
    echo   build-enhanced.bat server   ^| Build only server
    echo   build-enhanced.bat clean    ^| Clean build files
    goto :eof

:end
endlocal