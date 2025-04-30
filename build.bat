@echo off
setlocal enabledelayedexpansion

REM Check and install Wails
where wails >nul 2>nul || (
    echo Installing Wails...
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    FOR /F "tokens=*" %%g IN ('go env GOPATH') do set PATH=!PATH!;%%g\bin
)

REM Create output directory
if not exist "bin" mkdir bin

REM Build CLI version
cd cmd\cli
go build -o ..\..\bin\passwordmanager.exe
cd ..\..

REM Build GUI version
cd cmd\gui
wails build
cd ..\..

REM Copy the GUI executable to the bin directory
copy /Y cmd\gui\build\bin\passwordmanager-gui.exe bin\

REM Clean up build artifacts
echo Cleaning up build artifacts...
rmdir /S /Q cmd\gui\build
rmdir /S /Q cmd\gui\frontend\wailsjs

echo Build completed. Executables are located in the bin directory:
echo - CLI: bin\passwordmanager.exe
echo - GUI: bin\passwordmanager-gui.exe