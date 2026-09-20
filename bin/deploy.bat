@echo off

chcp 65001
setlocal enabledelayedexpansion

REM 第一个参数是环境名
set ENV=%~1
shift

REM 收集剩余的服务器组
set GROUPS=
:collect
if "%~1"=="" goto run
set GROUPS=!GROUPS! %~1
shift
goto collect

:run
if "%ENV%"=="" (
    echo 错误: 缺少参数 环境名
    exit /b 1
)
if "!GROUPS!"=="" (
    echo 错误: 缺少参数 服务器组
    exit /b 1
)

java -Dfile.encoding=UTF-8 -jar taskssh-1.2.0.jar deploy !GROUPS! -i "%ENV%.yaml"