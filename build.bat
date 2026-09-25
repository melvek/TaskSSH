@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion


REM ============================================================
REM  TaskSSH Build Script
REM ============================================================

set BINARY=taskssh
set DIST=dist
set RELEASE=release

echo.
echo ============================================================
echo  TaskSSH Build
echo  Output:     %RELEASE%\
echo ============================================================
echo.

REM ---- 清理旧产物 ----
if exist "%DIST%" rmdir /s /q "%DIST%"
if exist "%RELEASE%" rmdir /s /q "%RELEASE%"
mkdir "%DIST%"
mkdir "%RELEASE%"

REM ---- 依赖 ----
echo [0/6] Resolving dependencies...
go mod tidy
if errorlevel 1 (
    echo [ERROR] go mod tidy failed
    exit /b 1
)

set CGO_ENABLED=0

REM ---- 1. Linux amd64 ----
echo [1/6] Building Linux amd64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "!LDFLAGS!" -o "%DIST%\taskssh-linux-amd64" main.go
if errorlevel 1 (
    echo [ERROR] Linux amd64 build failed
    exit /b 1
)

REM ---- 2. Linux arm64 ----
echo [2/6] Building Linux arm64...
set GOOS=linux
set GOARCH=arm64
go build -ldflags "!LDFLAGS!" -o "%DIST%\taskssh-linux-arm64" main.go
if errorlevel 1 (
    echo [ERROR] Linux arm64 build failed
    exit /b 1
)

REM ---- 3. Windows amd64 ----
echo [3/6] Building Windows amd64...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "!LDFLAGS!" -o "%DIST%\taskssh-windows-amd64.exe" main.go
if errorlevel 1 (
    echo [ERROR] Windows amd64 build failed
    exit /b 1
)

REM ---- 4. macOS Intel ----
echo [4/6] Building macOS amd64...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "!LDFLAGS!" -o "%DIST%\taskssh-darwin-amd64" main.go
if errorlevel 1 (
    echo [ERROR] macOS amd64 build failed
    exit /b 1
)

REM ---- 5. macOS Apple Silicon ----
echo [5/6] Building macOS arm64...
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
echo [6/6] Preparing package files...
if exist testdata/inventory.yaml (
    copy testdata/inventory.yaml "%DIST%\" >nul
    echo   Copied: inventory.yaml
) else (
    echo   [WARN] inventory.yaml not found, skipped
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
call :package "%DIST%\taskssh-linux-arm64"       "taskssh"     "taskssh-linux-arm64"
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
if exist "inventory.yaml" copy "inventory.yaml" "%PKG_DIR%\" >nul
if exist "README.md" copy "README.md" "%PKG_DIR%\" >nul

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