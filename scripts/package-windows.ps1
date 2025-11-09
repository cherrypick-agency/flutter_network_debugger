# PowerShell script for creating Windows installer
param(
    [string]$Version = "1.0.0"
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Green
Write-Host "Windows Installer Packaging Script" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

# Configuration
$APP_NAME = "Network Debugger"
$BUNDLE_NAME = "NetworkDebugger"
$DIST_DIR = Join-Path $PSScriptRoot "..\dist"
$INSTALLER_NAME = "NetworkDebugger-$Version-windows-amd64.zip"

Write-Host "`nConfiguration:" -ForegroundColor Yellow
Write-Host "  App Name: $APP_NAME"
Write-Host "  Version: $Version"
Write-Host "  Output: $INSTALLER_NAME"

# Clean previous builds
Write-Host "`nStep 1/5: Cleaning previous builds..." -ForegroundColor Yellow
if (Test-Path $DIST_DIR) {
    Remove-Item -Path $DIST_DIR -Recurse -Force
}
New-Item -ItemType Directory -Path $DIST_DIR -Force | Out-Null

# Build Go server binary
Write-Host "`nStep 2/5: Building Go server binary..." -ForegroundColor Yellow
Set-Location (Join-Path $PSScriptRoot "..")
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

Write-Host "  Building for Windows amd64..."
go build -ldflags="-s -w" -o "$DIST_DIR\server_windows_amd64.exe" .\cmd\network-debugger

if (-not (Test-Path "$DIST_DIR\server_windows_amd64.exe")) {
    Write-Host "Error: Failed to build Go binary" -ForegroundColor Red
    exit 1
}
Write-Host "  ✓ Go binary built successfully" -ForegroundColor Green

# Build Flutter Windows app
Write-Host "`nStep 3/5: Building Flutter Windows app..." -ForegroundColor Yellow
Set-Location frontend
flutter clean
flutter pub get
flutter build windows --release

$BUILD_DIR = "build\windows\x64\runner\Release"
if (-not (Test-Path $BUILD_DIR)) {
    Write-Host "Error: Flutter build failed" -ForegroundColor Red
    exit 1
}
Write-Host "  ✓ Flutter app built successfully" -ForegroundColor Green

# Package app
Write-Host "`nStep 4/5: Packaging app..." -ForegroundColor Yellow

$PACKAGE_DIR = Join-Path $DIST_DIR "package"
New-Item -ItemType Directory -Path $PACKAGE_DIR -Force | Out-Null

# Copy Flutter app
Write-Host "  Copying Flutter app files..."
Copy-Item -Path "$BUILD_DIR\*" -Destination $PACKAGE_DIR -Recurse

# Create resources directory
$RESOURCES_DIR = Join-Path $PACKAGE_DIR "resources"
New-Item -ItemType Directory -Path $RESOURCES_DIR -Force | Out-Null

# Copy Go binary
Write-Host "  Copying Go server binary..."
Copy-Item -Path "$DIST_DIR\server_windows_amd64.exe" -Destination $RESOURCES_DIR

# Create install script
Write-Host "  Creating installer script..."
$INSTALL_SCRIPT = @"
@echo off
echo ========================================
echo Network Debugger Installer
echo ========================================
echo.

set INSTALL_DIR=%LOCALAPPDATA%\NetworkDebugger

echo Installing to: %INSTALL_DIR%
echo.

if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

echo Copying files...
xcopy /E /I /Y * "%INSTALL_DIR%"

echo Creating desktop shortcut...
powershell -Command `$WshShell = New-Object -comObject WScript.Shell; `$Shortcut = `$WshShell.CreateShortcut('%USERPROFILE%\Desktop\Network Debugger.lnk'); `$Shortcut.TargetPath = '%INSTALL_DIR%\frontend.exe'; `$Shortcut.WorkingDirectory = '%INSTALL_DIR%'; `$Shortcut.Save()

echo Creating start menu shortcut...
if not exist "%APPDATA%\Microsoft\Windows\Start Menu\Programs\Network Debugger" mkdir "%APPDATA%\Microsoft\Windows\Start Menu\Programs\Network Debugger"
powershell -Command `$WshShell = New-Object -comObject WScript.Shell; `$Shortcut = `$WshShell.CreateShortcut('%APPDATA%\Microsoft\Windows\Start Menu\Programs\Network Debugger\Network Debugger.lnk'); `$Shortcut.TargetPath = '%INSTALL_DIR%\frontend.exe'; `$Shortcut.WorkingDirectory = '%INSTALL_DIR%'; `$Shortcut.Save()

echo.
echo ========================================
echo Installation completed successfully!
echo ========================================
echo.
echo The application has been installed to:
echo %INSTALL_DIR%
echo.
echo Shortcuts created on:
echo - Desktop
echo - Start Menu
echo.
pause
"@

$INSTALL_SCRIPT | Out-File -FilePath (Join-Path $PACKAGE_DIR "install.bat") -Encoding ASCII

# Create uninstall script
$UNINSTALL_SCRIPT = @"
@echo off
echo ========================================
echo Network Debugger Uninstaller
echo ========================================
echo.

set INSTALL_DIR=%LOCALAPPDATA%\NetworkDebugger

echo This will remove Network Debugger from your computer.
echo Installation directory: %INSTALL_DIR%
echo.
pause

echo Removing files...
if exist "%INSTALL_DIR%" rmdir /S /Q "%INSTALL_DIR%"

echo Removing shortcuts...
if exist "%USERPROFILE%\Desktop\Network Debugger.lnk" del "%USERPROFILE%\Desktop\Network Debugger.lnk"
if exist "%APPDATA%\Microsoft\Windows\Start Menu\Programs\Network Debugger" rmdir /S /Q "%APPDATA%\Microsoft\Windows\Start Menu\Programs\Network Debugger"

echo.
echo ========================================
echo Uninstallation completed!
echo ========================================
echo.
pause
"@

$UNINSTALL_SCRIPT | Out-File -FilePath (Join-Path $PACKAGE_DIR "uninstall.bat") -Encoding ASCII

# Create README
$README = @"
Network Debugger v$Version

INSTALLATION
============
1. Double-click install.bat
2. Follow the on-screen instructions
3. The app will be installed to %LOCALAPPDATA%\NetworkDebugger
4. Shortcuts will be created on Desktop and Start Menu

UNINSTALLATION
==============
1. Double-click uninstall.bat in the installation directory
   OR
2. Run: %LOCALAPPDATA%\NetworkDebugger\uninstall.bat

SYSTEM REQUIREMENTS
===================
- Windows 10 or later (64-bit)
- 200 MB disk space
- Internet connection (for updates)

TROUBLESHOOTING
===============
If the app doesn't start:
1. Check if port 9092 is available
2. Run as Administrator
3. Check Windows Firewall settings

For more help, visit:
https://github.com/YOUR_USERNAME/YOUR_REPO
"@

$README | Out-File -FilePath (Join-Path $PACKAGE_DIR "README.txt") -Encoding UTF8

Write-Host "  ✓ Package prepared successfully" -ForegroundColor Green

# Create ZIP archive
Write-Host "`nStep 5/5: Creating ZIP archive..." -ForegroundColor Yellow

Set-Location $DIST_DIR
Compress-Archive -Path "package\*" -DestinationPath $INSTALLER_NAME -Force

if (-not (Test-Path $INSTALLER_NAME)) {
    Write-Host "Error: Failed to create ZIP archive" -ForegroundColor Red
    exit 1
}

$ZIP_SIZE = (Get-Item $INSTALLER_NAME).Length / 1MB
$ZIP_SIZE = [math]::Round($ZIP_SIZE, 2)

Write-Host "  ✓ ZIP archive created successfully" -ForegroundColor Green

# Summary
Write-Host "`n========================================" -ForegroundColor Green
Write-Host "Build completed successfully!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "`nOutput:" -ForegroundColor Yellow
Write-Host "  Location: $DIST_DIR\$INSTALLER_NAME"
Write-Host "  Size: $ZIP_SIZE MB"
Write-Host "`nNext steps:" -ForegroundColor Yellow
Write-Host "  1. Test the installer"
Write-Host "  2. Upload to GitHub Releases"
Write-Host ""
