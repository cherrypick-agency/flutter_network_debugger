param(
  [string]$Version = "1.0.0"
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Green
Write-Host "Windows MSI Packaging Script (WiX)" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

# Normalize version for MSI (strip leading 'v', ensure x.y.z)
if ($Version.StartsWith("v")) {
  $Version = $Version.Substring(1)
}
if ($Version -notmatch '^\d+(\.\d+){0,2}$') {
  Write-Host "Provided version '$Version' is not MSI-friendly. Fallback to 1.0.0" -ForegroundColor Yellow
  $Version = "1.0.0"
}
if ($Version -notmatch '^\d+\.\d+\.\d+$') {
  # Expand to 3 parts if needed (e.g., 1 -> 1.0.0, 1.2 -> 1.2.0)
  $parts = $Version.Split('.')
  if ($parts.Length -eq 1) { $Version = "$($parts[0]).0.0" }
  elseif ($parts.Length -eq 2) { $Version = "$($parts[0]).$($parts[1]).0" }
}

# Configuration
$APP_NAME = "Network Debugger"
$MANUFACTURER = "CherryPick Agency"
$DIST_DIR = Join-Path $PSScriptRoot "..\dist"
$PKG_DIR = Join-Path $DIST_DIR "package"
$WIX_DIR = Join-Path $DIST_DIR "wix"
$OUTPUT_MSI = Join-Path $DIST_DIR "NetworkDebugger-$Version-windows-amd64.msi"

Write-Host "`nConfiguration:" -ForegroundColor Yellow
Write-Host "  App Name: $APP_NAME"
Write-Host "  Version:  $Version"
Write-Host "  Output:   $OUTPUT_MSI"

# Ensure WiX tools available
$heat = Get-Command heat.exe -ErrorAction SilentlyContinue
$candle = Get-Command candle.exe -ErrorAction SilentlyContinue
$light = Get-Command light.exe -ErrorAction SilentlyContinue
if (-not $heat -or -not $candle -or -not $light) {
  Write-Host "`nWiX Toolset is required (heat, candle, light). Install with:" -ForegroundColor Yellow
  Write-Host "  choco install wixtoolset -y"
  throw "WiX Toolset not found in PATH."
}

# Clean previous builds
Write-Host "`nStep 1/6: Cleaning previous builds..." -ForegroundColor Yellow
if (Test-Path $DIST_DIR) { Remove-Item -Path $DIST_DIR -Recurse -Force }
New-Item -ItemType Directory -Path $DIST_DIR | Out-Null
New-Item -ItemType Directory -Path $WIX_DIR | Out-Null

# Build Go server binary
Write-Host "`nStep 2/6: Building Go server binary..." -ForegroundColor Yellow
Set-Location (Join-Path $PSScriptRoot "..")
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
Write-Host "  Building for Windows amd64..."
go build -ldflags="-s -w" -o "$DIST_DIR\server_windows_amd64.exe" .\cmd\network-debugger
if (-not (Test-Path "$DIST_DIR\server_windows_amd64.exe")) {
  throw "Failed to build Go binary"
}
Write-Host "  ✓ Go binary built successfully" -ForegroundColor Green

# Build Flutter Windows app
Write-Host "`nStep 3/6: Building Flutter Windows app..." -ForegroundColor Yellow
Set-Location frontend
flutter clean
flutter pub get
flutter build windows --release
$BUILD_DIR = "build\windows\x64\runner\Release"
if (-not (Test-Path $BUILD_DIR)) {
  throw "Flutter build failed (folder not found: $BUILD_DIR)"
}
Write-Host "  ✓ Flutter app built successfully" -ForegroundColor Green

# Stage package contents
Write-Host "`nStep 4/6: Staging package contents..." -ForegroundColor Yellow
New-Item -ItemType Directory -Path $PKG_DIR -Force | Out-Null
# Copy Flutter app output
Copy-Item -Path "$BUILD_DIR\*" -Destination $PKG_DIR -Recurse -Force
# Resources folder
$RESOURCES_DIR = Join-Path $PKG_DIR "resources"
New-Item -ItemType Directory -Path $RESOURCES_DIR -Force | Out-Null
Copy-Item -Path "$DIST_DIR\server_windows_amd64.exe" -Destination $RESOURCES_DIR -Force
Write-Host "  ✓ Package prepared" -ForegroundColor Green

# Generate WiX source files
Write-Host "`nStep 5/6: Generating WiX sources..." -ForegroundColor Yellow
Set-Location $WIX_DIR

# Stable GUIDs (do not change across versions)
$UpgradeCode = "C5EBDC5B-0A33-4D49-8D5B-EE359B2BD704"
$ShortcutsComponentGuid = "7F7E5C51-44A5-4A7C-9E0F-3C27B5F12E9E"

# Product.wxs
$productWxs = @"
<?xml version="1.0" encoding="UTF-8"?>
<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi">
  <Product
    Id="*"
    Name="$APP_NAME"
    Language="1033"
    Version="$Version"
    Manufacturer="$MANUFACTURER"
    UpgradeCode="$UpgradeCode">
    <Package InstallerVersion="500" Compressed="yes" InstallScope="perMachine" />
    <MajorUpgrade DowngradeErrorMessage="A newer version of [ProductName] is already installed." />
    <MediaTemplate EmbedCab="yes" CompressionLevel="high"/>

    <Feature Id="MainFeature" Title="$APP_NAME" Level="1">
      <ComponentGroupRef Id="AppFiles" />
      <ComponentRef Id="ShortcutsComponent" />
    </Feature>

    <Property Id="WIXUI_INSTALLDIR" Value="INSTALLDIR" />
    <UIRef Id="WixUI_InstallDir" />

    <Directory Id="TARGETDIR" Name="SourceDir">
      <Directory Id="ProgramFilesFolder">
        <Directory Id="INSTALLDIR" Name="NetworkDebugger">
          <!-- Files harvested into INSTALLDIR -->
        </Directory>
      </Directory>
      <Directory Id="ProgramMenuFolder">
        <Directory Id="ApplicationProgramsFolder" Name="$APP_NAME" />
      </Directory>
      <Directory Id="DesktopFolder" />
    </Directory>

    <!-- Shortcuts component uses a registry KeyPath for reference -->
    <Component Id="ShortcutsComponent" Guid="$ShortcutsComponentGuid">
      <RegistryValue Root="HKCU" Key="Software\NetworkDebugger" Name="installed" Type="integer" Value="1" KeyPath="yes" />
      <Shortcut Id="StartMenuShortcut"
                Name="$APP_NAME"
                Directory="ApplicationProgramsFolder"
                Target="[INSTALLDIR]frontend.exe"
                WorkingDirectory="INSTALLDIR"
                Advertise="no" />
      <Shortcut Id="DesktopShortcut"
                Name="$APP_NAME"
                Directory="DesktopFolder"
                Target="[INSTALLDIR]frontend.exe"
                WorkingDirectory="INSTALLDIR"
                Advertise="no" />
      <RemoveFolder Id="RemoveShortcutsDir" Directory="ApplicationProgramsFolder" On="uninstall" />
    </Component>
  </Product>
</Wix>
"@
Set-Content -Path (Join-Path $WIX_DIR "Product.wxs") -Value $productWxs -Encoding UTF8

# Harvest application files into AppFiles.wxs
& $heat.Path dir $PKG_DIR -cg AppFiles -dr INSTALLDIR -srd -sreg -gg -sfrag -var var.SourceDir -out (Join-Path $WIX_DIR "AppFiles.wxs")

# Compile WiX sources
& $candle.Path (Join-Path $WIX_DIR "Product.wxs") (Join-Path $WIX_DIR "AppFiles.wxs") -dSourceDir="$PKG_DIR" -ext WixUIExtension -ext WixUtilExtension -o "$WIX_DIR\"

# Link to MSI
& $light.Path "$WIX_DIR\Product.wixobj" "$WIX_DIR\AppFiles.wixobj" -b "$PKG_DIR" -ext WixUIExtension -ext WixUtilExtension -o "$OUTPUT_MSI"

if (-not (Test-Path $OUTPUT_MSI)) {
  throw "Failed to create MSI"
}
Write-Host "  ✓ MSI created: $OUTPUT_MSI" -ForegroundColor Green

# Summary
Write-Host "`n========================================" -ForegroundColor Green
Write-Host "Build completed successfully!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "`nOutput:" -ForegroundColor Yellow
Write-Host "  $OUTPUT_MSI"
Write-Host ""


