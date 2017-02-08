@echo off
setlocal

set bin=bin
set option=%1
set tag=%2

for %%* in (.) do set appName=%%~nx*

rd %bin% /q/s 2>nul
md %bin%

if /i "%tag%"=="" (
    echo No Tag supplied
    set tag=dev
) else (
    echo package main > version.go
    echo. >> version.go
    echo const appVersion = "%tag%" >> version.go

    git add version.go
    git commit -m "Running release for %tag%"
    git tag %tag%
)

go test 
call :THE_BUILD windows

if /i "%option%"=="release" (
    call :THE_BUILD linux
    call :THE_BUILD darwin
)
exit /b 0

:THE_BUILD
    setlocal
    echo Building :%TIME%
    set GOOS=%1
    if /i "%GOOS%"=="windows" set ext=.exe
    set file=%bin%/%appName%-%tag%-%GOOS%%ext%
    del %file% 2>nul
    go build -o %file%
    echo Done     :%TIME%
goto :EOF