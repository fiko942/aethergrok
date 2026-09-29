<#
.SYNOPSIS
    AetherGrok Desktop GUI Studio - Windows Universal Developer Launcher
.DESCRIPTION
    Launches live desktop development, web preview, unit testing, or builds for Windows.
.PARAMETER Mode
    Development mode: 'desktop' (default), 'ui' (browser only), 'test', or 'build'.
.EXAMPLE
    .\dev.ps1
    .\dev.ps1 ui
    .\dev.ps1 test
    .\dev.ps1 build
#>

param (
    [Parameter(Position = 0)]
    [ValidateSet("desktop", "wails", "ui", "web", "browser", "test", "build")]
    [string]$Mode = "desktop"
)

$ErrorActionPreference = "Stop"

# Colors & Banner
Write-Host "======================================================" -ForegroundColor Cyan
Write-Host "   AetherGrok Desktop Studio - Windows Launcher       " -ForegroundColor Cyan
Write-Host "======================================================" -ForegroundColor Cyan

$ProjectRoot = $PSScriptRoot
$FrontendDir = Join-Path $ProjectRoot "frontend"

# 1. Locate Go Compiler Binary
$GoBin = "go"
if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    if (Test-Path "C:\Program Files\Go\bin\go.exe") {
        $GoBin = "C:\Program Files\Go\bin\go.exe"
        $env:PATH = "C:\Program Files\Go\bin;$env:PATH"
    } elseif (Test-Path "$env:USERPROFILE\go\bin\go.exe") {
        $GoBin = "$env:USERPROFILE\go\bin\go.exe"
        $env:PATH = "$env:USERPROFILE\go\bin;$env:PATH"
    } else {
        Write-Host "Error: 'go' is not installed or not found in PATH." -ForegroundColor Red
        Write-Host "Please install Go from https://go.dev/dl/ and add it to PATH." -ForegroundColor Yellow
        exit 1
    }
}

# Ensure Go bin directory is in PATH for tools like wails
$GoPathBin = Join-Path $env:USERPROFILE "go\bin"
if ($env:GOPATH) {
    $GoPathBin = Join-Path $env:GOPATH "bin"
}
if (-not ($env:PATH -split ';' -contains $GoPathBin)) {
    $env:PATH = "$GoPathBin;$env:PATH"
}

# 2. Locate Package Manager (pnpm or npm)
$PkgManager = "npm"
if (Get-Command "pnpm" -ErrorAction SilentlyContinue) {
    $PkgManager = "pnpm"
} elseif (-not (Get-Command "npm" -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Neither 'pnpm' nor 'npm' was found in PATH." -ForegroundColor Red
    Write-Host "Please install Node.js from https://nodejs.org/" -ForegroundColor Yellow
    exit 1
}

# 3. Handle Development Modes
if ($Mode -eq "ui" -or $Mode -eq "web" -or $Mode -eq "browser") {
    Write-Host "`nStarting Svelte 5 Frontend in Web Browser Mode..." -ForegroundColor Green
    Push-Location $FrontendDir
    try {
        if (-not (Test-Path "node_modules")) {
            Write-Host "Installing frontend dependencies via $PkgManager..." -ForegroundColor Yellow
            & $PkgManager install
        }
        & $PkgManager run dev
    } finally {
        Pop-Location
    }
} elseif ($Mode -eq "test") {
    Write-Host "`nRunning Go unit tests and Frontend verification..." -ForegroundColor Green
    
    Write-Host "`n[1/2] Running Go unit tests across all packages:" -ForegroundColor Yellow
    & $GoBin test ./...

    Write-Host "`n[2/2] Running Svelte 5 frontend check:" -ForegroundColor Yellow
    Push-Location $FrontendDir
    try {
        & $PkgManager run check
    } finally {
        Pop-Location
    }
    Write-Host "`n[OK] All tests and checks passed successfully!" -ForegroundColor Green
} elseif ($Mode -eq "build") {
    Write-Host "`nBuilding standalone Windows executable..." -ForegroundColor Green
    & (Join-Path $ProjectRoot "build-windows.ps1") -SkipInstaller
} else {
    # 'desktop' or 'wails' mode
    $WailsBin = "wails"
    if (-not (Get-Command "wails" -ErrorAction SilentlyContinue)) {
        $CandidateWails = Join-Path $GoPathBin "wails.exe"
        if (Test-Path $CandidateWails) {
            $WailsBin = $CandidateWails
        } else {
            Write-Host "'wails' CLI not found. Installing Wails v2 latest..." -ForegroundColor Yellow
            & $GoBin install github.com/wailsapp/wails/v2/cmd/wails@latest
            if (Test-Path $CandidateWails) {
                $WailsBin = $CandidateWails
            }
        }
    }

    # Ensure frontend dependencies are installed
    if (-not (Test-Path (Join-Path $FrontendDir "node_modules"))) {
        Write-Host "Installing frontend dependencies via $PkgManager..." -ForegroundColor Yellow
        Push-Location $FrontendDir
        try {
            & $PkgManager install
        } finally {
            Pop-Location
        }
    }

    Write-Host "`nStarting Wails Native Desktop Live-Development Studio..." -ForegroundColor Green
    Write-Host "Hot-reload enabled: edits in frontend/ or pkg/ will automatically update." -ForegroundColor Cyan
    & $WailsBin dev
}
