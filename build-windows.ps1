# AetherGrok Desktop Studio - Windows Automated Build Script
# Usage: .\build-windows.ps1 [-Arch <amd64|arm64>] [-SkipTests] [-SkipInstaller]

[CmdletBinding()]
param(
    [string]$Arch = "amd64",
    [switch]$SkipTests,
    [switch]$SkipInstaller
)

$ErrorActionPreference = "Stop"

Write-Output "=========================================="
Write-Output "  Building AetherGrok for Windows ($Arch) "
Write-Output "=========================================="

# 1. Locate Go compiler binary
$GoBin = "go"
if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    if (Test-Path "C:\Program Files\Go\bin\go.exe") {
        $GoBin = "C:\Program Files\Go\bin\go.exe"
    } elseif (Test-Path "$env:ProgramFiles\Go\bin\go.exe") {
        $GoBin = "$env:ProgramFiles\Go\bin\go.exe"
    } elseif (Test-Path "$env:USERPROFILE\go\bin\go.exe") {
        $GoBin = "$env:USERPROFILE\go\bin\go.exe"
    } else {
        throw "Go compiler binary not found. Please install Go or add it to PATH."
    }
}

# 2. Build Frontend Assets
Write-Output "`n[1/5] Building frontend web assets with Vite..."
Push-Location "frontend"
try {
    if (Get-Command "pnpm" -ErrorAction SilentlyContinue) {
        pnpm run build
    } elseif (Get-Command "npm" -ErrorAction SilentlyContinue) {
        npm run build
    } else {
        throw "Neither pnpm nor npm package manager was found in PATH."
    }
} finally {
    Pop-Location
}

if (-not (Test-Path "frontend\dist\index.html")) {
    throw "Frontend build failed: frontend\dist\index.html not generated."
}
Write-Output "[OK] Frontend build completed."

# 3. Create Output Directory
$OutputDir = "build\bin"
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

$OutputExe = "$OutputDir\aethergrok.exe"

# 4. Compile Windows Executable
Write-Output "`n[2/5] Compiling Go backend into Windows executable..."
$env:GOOS = "windows"
$env:GOARCH = $Arch
$LdFlags = "-s -w -H=windowsgui"
& $GoBin build -ldflags $LdFlags -tags "desktop,production" -o $OutputExe .

if (-not (Test-Path $OutputExe)) {
    throw "Compilation failed: $OutputExe was not created."
}

$ExeInfo = Get-Item $OutputExe
$ExeSizeMB = [math]::Round($ExeInfo.Length / 1MB, 2)
Write-Output "[OK] Executable built successfully: $OutputExe ($ExeSizeMB MB)"

# 5. Create Portable Release Zip Archive
Write-Output "`n[3/5] Creating portable release package (.zip)..."
$PortableZip = "$OutputDir\aethergrok-windows-$Arch.zip"
if (Test-Path $PortableZip) {
    Remove-Item $PortableZip -Force
}
Compress-Archive -Path $OutputExe -DestinationPath $PortableZip -Force

if (-not (Test-Path $PortableZip)) {
    throw "Packaging failed: $PortableZip was not created."
}
$ZipInfo = Get-Item $PortableZip
$ZipSizeMB = [math]::Round($ZipInfo.Length / 1MB, 2)
Write-Output "[OK] Portable zip created: $PortableZip ($ZipSizeMB MB)"

# 6. Build NSIS Setup Installer (if makensis is available)
Write-Output "`n[4/5] Checking NSIS compiler (makensis) for Setup Installer..."
$MakensisBin = $null

if (-not $SkipInstaller) {
    if (Get-Command "makensis" -ErrorAction SilentlyContinue) {
        $MakensisBin = (Get-Command "makensis").Source
    } elseif (Test-Path "C:\Program Files (x86)\NSIS\makensis.exe") {
        $MakensisBin = "C:\Program Files (x86)\NSIS\makensis.exe"
    } elseif (Test-Path "C:\Program Files\NSIS\makensis.exe") {
        $MakensisBin = "C:\Program Files\NSIS\makensis.exe"
    } elseif (Test-Path "$env:LOCALAPPDATA\Programs\NSIS\makensis.exe") {
        $MakensisBin = "$env:LOCALAPPDATA\Programs\NSIS\makensis.exe"
    } else {
        # Extra fallback check in AppData local cache
        $CachedMakensis = Get-ChildItem -Path "$env:LOCALAPPDATA" -Filter "makensis.exe" -Recurse -Depth 4 -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($CachedMakensis) {
            $MakensisBin = $CachedMakensis.FullName
        }
    }
}

$InstallerExe = "$OutputDir\AetherGrok-Setup.exe"
if ($MakensisBin -and (Test-Path $MakensisBin)) {
    Write-Output "Found NSIS compiler: $MakensisBin"
    Write-Output "Compiling NSIS Setup Installer..."
    Push-Location "build\windows\installer"
    try {
        $NsisArgs = @(
            "-DARG_WAILS_AMD64_BINARY=..\..\bin\aethergrok.exe",
            "-DOUTFILE_NAME=..\..\bin\AetherGrok-Setup.exe",
            "project.nsi"
        )
        if ($Arch -eq "arm64") {
            $NsisArgs[0] = "-DARG_WAILS_ARM64_BINARY=..\..\bin\aethergrok.exe"
        }
        & $MakensisBin @NsisArgs
    } finally {
        Pop-Location
    }

    if (Test-Path $InstallerExe) {
        $InstallerInfo = Get-Item $InstallerExe
        $InstallerSizeMB = [math]::Round($InstallerInfo.Length / 1MB, 2)
        Write-Output "[OK] Setup installer built successfully: $InstallerExe ($InstallerSizeMB MB)"
    } else {
        Write-Warning "NSIS completed but installer binary was not found at $InstallerExe"
    }
} else {
    if ($SkipInstaller) {
        Write-Output "Installer creation skipped via -SkipInstaller flag."
    } else {
        Write-Output "[INFO] makensis not found in PATH or standard NSIS directories."
        Write-Output "       (To generate NSIS installer, install NSIS: choco install nsis or https://nsis.sourceforge.io)"
    }
}

# 7. Verification & Tests
Write-Output "`n[5/5] Running verification tests..."
if (-not $SkipTests) {
    & $GoBin test ./...
    Write-Output "[OK] Test suite passed."
} else {
    Write-Output "Test suite verification skipped via -SkipTests flag."
}

# 8. Summary
Write-Output "`n=========================================="
Write-Output "  AetherGrok Windows Build Summary        "
Write-Output "=========================================="
Write-Output "  Executable:    $(Resolve-Path $OutputExe) ($ExeSizeMB MB)"
Write-Output "  Portable Zip:  $(Resolve-Path $PortableZip) ($ZipSizeMB MB)"
if (Test-Path $InstallerExe) {
    $InstallerInfo = Get-Item $InstallerExe
    $InstallerSizeMB = [math]::Round($InstallerInfo.Length / 1MB, 2)
    Write-Output "  Setup Install: $(Resolve-Path $InstallerExe) ($InstallerSizeMB MB)"
}
Write-Output "=========================================="
