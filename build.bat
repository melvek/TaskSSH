@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM ============================================================
REM  TaskSSH 多平台发布脚本
REM  编译 Windows / Linux / macOS 四个平台
REM  打包为 zip，内部二进制统一命名为 taskssh / taskssh.exe
REM ============================================================

set BINARY=taskssh
set DIST=dist
set RELEASE=release

REM ---- 版本信息 ----
for /f "tokens=*" %%i in ('git describe --tags --always --dirty 2^>nul') do set VERSION=%%i
if "!VERSION!"=="" set VERSION=dev

for /f "tokens=*" %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i
if "!GIT_COMMIT!"=="" set GIT_COMMIT=unknown

REM ---- 编译参数 ----
set LDFLAGS=-s -w -X "mestrap.com/taskssh/internal/utils.Version=!VERSION!" -X "mestrap.com/taskssh/internal/utils.GitCommit=!GIT_COMMIT!"

echo.
echo ============================================================
echo  TaskSSH Build
echo ============================================================
echo  Version:    !VERSION!
echo  Git commit: !GIT_COMMIT!
echo  Output:     %RELEASE%\
echo ============================================================
echo.

REM ---- 清理旧产物 ----
if exist "%DIST%" rmdir /s /q "%DIST%"
if exist "%RELEASE%" rmdir /s /q "%RELEASE%"
mkdir "%DIST%"
mkdir "%RELEASE%"

REM ---- 依赖 ----
echo [0/5] Resolving dependencies...
go mod tidy
if errorlevel 1 (
    echo [ERROR] go mod tidy failed
    exit /b 1
)

set CGO_ENABLED=0

REM ---- 1. Linux amd64 ----
echo [1/5] Building Linux amd64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "!LDFLAGS!" -o "%DIST%\taskssh-linux-amd64" main.go
if errorlevel 1 (
    echo [ERROR] Linux amd64 build failed
    exit /b 1
)

REM ---- 2. Windows amd64 ----
echo [2/5] Building Windows amd64...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "!LDFLAGS!" -o "%DIST%\taskssh-windows-amd64.exe" main.go
if errorlevel 1 (
    echo [ERROR] Windows amd64 build failed
    exit /b 1
)

REM ---- 3. macOS Intel ----
echo [3/5] Building macOS amd64...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "!LDFLAGS!" -o "%DIST%\taskssh-darwin-amd64" main.go
if errorlevel 1 (
    echo [ERROR] macOS amd64 build failed
    exit /b 1
)

REM ---- 4. macOS Apple Silicon ----
echo [4/5] Building macOS arm64...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags "!LDFLAGS!" -o "%DIST%\taskssh-darwin-arm64" main.go
if errorlevel 1 (
    echo [ERROR] macOS arm64 build failed
    exit /b 1
)

REM ---- 还原环境变量 ----
set GOOS=
set GOARCH=
set CGO_ENABLED=

REM ---- 复制示例文件 ----
echo.
echo [5/5] Preparing package files...
if exist inventory.example.yaml (
    copy inventory.example.yaml "%DIST%\" >nul
    echo   Copied: inventory.example.yaml
) else (
    echo   [WARN] inventory.example.yaml not found, skipped
)
if exist README.md (
    copy README.md "%DIST%\" >nul
    echo   Copied: README.md
) else (
    echo   [WARN] README.md not found, skipped
)

REM ---- 打包 ----
echo.
echo [PACK] Creating release archives...

call :package "%DIST%\taskssh-windows-amd64.exe" "taskssh.exe" "taskssh-windows-amd64"
call :package "%DIST%\taskssh-linux-amd64"       "taskssh"     "taskssh-linux-amd64"
call :package "%DIST%\taskssh-darwin-amd64"      "taskssh"     "taskssh-darwin-amd64"
call :package "%DIST%\taskssh-darwin-arm64"      "taskssh"     "taskssh-darwin-arm64"

REM ---- 完成 ----
echo.
echo ============================================================
echo  Build complete
echo ============================================================
echo.
echo Release archives:
dir /b "%RELEASE%\*.zip" 2>nul
echo.
echo Output directory: %RELEASE%\
echo ============================================================
echo.

endlocal
pause
goto :eof


REM ============================================================
REM  子过程：打包单个平台
REM  参数1：源二进制路径
REM  参数2：包内二进制名（taskssh 或 taskssh.exe）
REM  参数3：zip 名（不含扩展名）
REM ============================================================
:package
set SRC=%~1
set TARGET_NAME=%~2
set ZIP_NAME=%~3

if not exist "%SRC%" (
    echo   [WARN] Not found: %SRC%, skipped
    goto :eof
)

REM 用独立临时目录，避免与别的包冲突
set PKG_DIR=%DIST%\pkg-%ZIP_NAME%
if exist "%PKG_DIR%" rmdir /s /q "%PKG_DIR%"
mkdir "%PKG_DIR%"

REM 复制二进制并重命名
copy "%SRC%" "%PKG_DIR%\%TARGET_NAME%" >nul

REM 复制示例清单和 README（存在才复制）
if exist "%DIST%\inventory.example.yaml" copy "%DIST%\inventory.example.yaml" "%PKG_DIR%\" >nul
if exist "%DIST%\README.md" copy "%DIST%\README.md" "%PKG_DIR%\" >nul

REM 打包
powershell -Command "Compress-Archive -Path '%PKG_DIR%\*' -DestinationPath '%RELEASE%\%ZIP_NAME%.zip' -Force"
if errorlevel 1 (
    echo   [ERROR] Pack failed: %ZIP_NAME%
) else (
    echo   Packed: %ZIP_NAME%.zip
)

REM 清理临时目录
rmdir /s /q "%PKG_DIR%"
goto :eof