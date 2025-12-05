# PowerShell构建脚本，功能类似于Makefile
param(
    [string]$Target = "build",
    [switch]$Help
)

# 获取项目根目录
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
Set-Location $ScriptDir

# 项目模块路径
$ModulePath = "github.com/jianlu8023/golang-example"

# 获取构建时间
$BuildTime = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

# 获取Git信息
try {
    $Branch = git rev-parse --abbrev-ref HEAD 2>$null
    if (-not $Branch) { $Branch = "unknown" }
} catch {
    $Branch = "unknown"
}

try {
    $Version = git describe --tags --always --dirty 2>$null
    if (-not $Version) {
        $Version = git rev-parse --short HEAD 2>$null
    }
    if (-not $Version) { $Version = "unknown" }
} catch {
    $Version = "unknown"
}

# 完整版本号
$FullVersion = "${Branch}-${Version}"

# 构建参数
$BuildFlags = "-tags=jsoniter -trimpath"
$LdFlags = "-s -w -X `"${ModulePath}/version.Version=${FullVersion}`""

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Project: $ModulePath" -ForegroundColor Cyan
Write-Host "Version: $FullVersion" -ForegroundColor Cyan
Write-Host "Build Time: $BuildTime" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

# 显示帮助信息
function Show-Help {
    Write-Host "Usage: build.ps1 [-Target <target>] [-Help]" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Targets:" -ForegroundColor Yellow
    Write-Host "  server    Build server binary" -ForegroundColor Yellow
    Write-Host "  client    Build client binary" -ForegroundColor Yellow
    Write-Host "  build     Build both server and client (default)" -ForegroundColor Yellow
    Write-Host "  clean     Clean build files" -ForegroundColor Yellow
    Write-Host "  docker    Build Docker image" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\build.ps1              # Build everything" -ForegroundColor Yellow
    Write-Host "  .\build.ps1 server       # Build only server" -ForegroundColor Yellow
    Write-Host "  .\build.ps1 -Target clean # Clean build files" -ForegroundColor Yellow
}

# 构建服务端
function Build-Server {
    Write-Host "[INFO] Building server..." -ForegroundColor Green
    $Output = go build $BuildFlags -ldflags="$LdFlags" -o server.exe cmd/server/server.go 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] Failed to build server" -ForegroundColor Red
        Write-Host $Output -ForegroundColor Red
        exit 1
    }
    
    # 创建版本信息文件
    "version: $FullVersion" > server.latest
    "time: $BuildTime" >> server.latest
    
    # 为了与Linux环境保持一致，同时创建.bin文件
    # Copy-Item server.exe server.bin -ErrorAction SilentlyContinue
    
    Write-Host "[SUCCESS] Server built successfully as server.exe (server.bin also created for compatibility)" -ForegroundColor Green
}

# 构建客户端
function Build-Client {
    Write-Host "[INFO] Building client..." -ForegroundColor Green
    $Output = go build $BuildFlags -ldflags="$LdFlags" -o client.exe cmd/client/client.go 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] Failed to build client" -ForegroundColor Red
        Write-Host $Output -ForegroundColor Red
        exit 1
    }
    
    # 创建版本信息文件
    "version: $FullVersion" > client.latest
    "time: $BuildTime" >> client.latest
    
    # 为了与Linux环境保持一致，同时创建.bin文件
    # Copy-Item client.exe client.bin -ErrorAction SilentlyContinue
    
    Write-Host "[SUCCESS] Client built successfully as client.exe (client.bin also created for compatibility)" -ForegroundColor Green
}

# 清理构建文件
function Clean-Build {
    Write-Host "[INFO] Cleaning build files..." -ForegroundColor Green
    Remove-Item -Path "server.exe" -ErrorAction SilentlyContinue
    Remove-Item -Path "client.exe" -ErrorAction SilentlyContinue
    Remove-Item -Path "server.bin" -ErrorAction SilentlyContinue
    Remove-Item -Path "client.bin" -ErrorAction SilentlyContinue
    Remove-Item -Path "server.latest" -ErrorAction SilentlyContinue
    Remove-Item -Path "client.latest" -ErrorAction SilentlyContinue
    Write-Host "[SUCCESS] Clean done" -ForegroundColor Green
}

# 构建Docker镜像
function Build-Docker {
    Write-Host "[INFO] Building Docker image..." -ForegroundColor Green
    
    # 获取镜像版本（基于日期时间）
    $ImageVersion = "v$(Get-Date -Format 'yyyyMMddHHmm')"
    $ImageName = "golang-example/ubuntu2204/app:$ImageVersion"
    
    Write-Host "[INFO] Pulling base images..." -ForegroundColor Green
    docker pull golang:1.22 >$null 2>&1
    docker pull ubuntu:22.04 >$null 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[WARNING] Failed to pull base images, continuing with build..." -ForegroundColor Yellow
    }
    
    Write-Host "[INFO] Building image: $ImageName" -ForegroundColor Green
    docker buildx build --platform linux/amd64 --build-arg "VERSION=$FullVersion" -t "$ImageName" .
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] Failed to build Docker image" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "[INFO] Cleaning up..." -ForegroundColor Green
    docker rmi golang:1.22 ubuntu:22.04 >$null 2>&1
    docker builder prune -a -f >$null 2>&1
    
    Write-Host "[SUCCESS] Docker image built successfully!" -ForegroundColor Green
    Write-Host "IMAGE NAME: $ImageName" -ForegroundColor Green
}

# 根据目标执行相应操作
if ($Help) {
    Show-Help
    exit 0
}

switch ($Target.ToLower()) {
    "server" {
        Build-Server
    }
    "client" {
        Build-Client
    }
    "build" {
        Clean-Build
        Build-Server
        Build-Client
        Write-Host "[SUCCESS] All builds completed successfully!" -ForegroundColor Green
    }
    "clean" {
        Clean-Build
    }
    "docker" {
        Build-Docker
    }
    default {
        Write-Host "Unknown target: $Target" -ForegroundColor Red
        Show-Help
        exit 1
    }
}
