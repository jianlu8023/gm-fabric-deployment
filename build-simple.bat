@echo off

REM 简化版构建脚本

REM 清理构建文件
echo Cleaning build files...
if exist server.exe del /q server.exe
if exist client.exe del /q client.exe
if exist server.latest del /q server.latest
if exist client.latest del /q client.latest

REM 获取当前日期时间用于版本信息
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "year=%dt:~0,4%"
set "month=%dt:~4,2%"
set "day=%dt:~6,2%"
set "hour=%dt:~8,2%"
set "minute=%dt:~10,2%"
set "BUILDTIME=%year%-%month%-%day% %hour%:%minute%"

REM 尝试获取git版本信息
for /f "delims=" %%i in ('git describe --tags --always --dirty 2^>nul') do set "VERSION=%%i"
if "%VERSION%"=="" set "VERSION=unknown"

REM 构建服务端
 echo Building server...
go build -ldflags="-X main.version=%VERSION%" -o server.exe cmd\server\server.go
if errorlevel 1 (
    echo Error: Failed to build server
    pause
    exit /b 1
)
echo version : %VERSION%>server.latest
echo time : %BUILDTIME%>>server.latest
echo Server build done

REM 构建客户端
echo Building client...
go build -ldflags="-X main.version=%VERSION%" -o client.exe cmd\client\client.go
if errorlevel 1 (
    echo Error: Failed to build client
    pause
    exit /b 1
)
echo version : %VERSION%>client.latest
echo time : %BUILDTIME%>>client.latest
echo Client build done

echo All builds completed successfully!
pause