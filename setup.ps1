# =============================================================================
#  ZomCleaner — setup.ps1
#  Run this in PowerShell (as Administrator for full functionality)
#
#  Usage:
#    Right-click setup.ps1 -> "Run with PowerShell"
#    OR in PowerShell terminal: .\setup.ps1
# =============================================================================

#Requires -Version 5.1

# ─── Self-elevate if not admin ────────────────────────────────────────────────
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]"Administrator")

if (-not $isAdmin) {
    Write-Host ""
    Write-Host "  [WARN] Not running as Administrator." -ForegroundColor Yellow
    Write-Host "  Relaunching with elevated privileges..." -ForegroundColor Yellow
    Start-Process -FilePath "powershell.exe" `
        -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`"" `
        -Verb RunAs
    exit
}

# ─── Config ───────────────────────────────────────────────────────────────────
$SourceFile  = "clean1.go"
$BinaryName  = "zomcleaner.exe"
$DashboardURL = "http://localhost:8080"

# ─── Helpers ──────────────────────────────────────────────────────────────────
function Write-Banner {
    Clear-Host
    Write-Host ""
    Write-Host "  ╔══════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "  ║       ZomCleaner  —  Setup Wizard        ║" -ForegroundColor Cyan
    Write-Host "  ║   System-Aware Resource Reaper for Win   ║" -ForegroundColor Cyan
    Write-Host "  ╚══════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
}

function Write-Step($msg) {
    Write-Host ""
    Write-Host "  ▶ $msg" -ForegroundColor White
}

function Write-OK($msg)   { Write-Host "  [  OK ] $msg" -ForegroundColor Green  }
function Write-Warn($msg) { Write-Host "  [ WARN] $msg" -ForegroundColor Yellow }
function Write-Fail($msg) { Write-Host "  [ FAIL] $msg" -ForegroundColor Red    }
function Write-Info($msg) { Write-Host "  [INFO ] $msg" -ForegroundColor Cyan   }

# ─── Banner ───────────────────────────────────────────────────────────────────
Write-Banner
Write-OK "Running as Administrator."

# ─── Step 1: Check Go installation ───────────────────────────────────────────
Write-Step "Checking Go installation..."

$goCmd = Get-Command go -ErrorAction SilentlyContinue

if ($goCmd) {
    $goVer = & go version
    Write-OK "Go found: $goVer"
} else {
    Write-Warn "Go not found. Attempting install via winget..."

    $winget = Get-Command winget -ErrorAction SilentlyContinue
    if ($winget) {
        try {
            winget install --id GoLang.Go -e --silent `
                --accept-package-agreements --accept-source-agreements
            # Refresh PATH for current session
            $env:PATH += ";C:\Program Files\Go\bin;$env:USERPROFILE\go\bin"
            $goCmd = Get-Command go -ErrorAction SilentlyContinue
            if ($goCmd) {
                Write-OK "Go installed: $(& go version)"
            } else {
                Write-Fail "Go installed but not in PATH. Please restart PowerShell and re-run setup.ps1"
                Pause; exit 1
            }
        } catch {
            Write-Fail "winget install failed: $_"
            Write-Info "Download Go manually from https://go.dev/dl/ then re-run setup.ps1"
            Pause; exit 1
        }
    } else {
        Write-Fail "winget not available on this system."
        Write-Info "Download Go from https://go.dev/dl/ and re-run setup.ps1"
        Pause; exit 1
    }
}

# ─── Step 2: Verify source file ───────────────────────────────────────────────
Write-Step "Locating source file..."

if (-not (Test-Path $SourceFile)) {
    Write-Fail "'$SourceFile' not found in: $(Get-Location)"
    Write-Info "Make sure setup.ps1 is in the same folder as clean1.go"
    Pause; exit 1
}

Write-OK "Found: $SourceFile"

# ─── Step 3: Build binary ─────────────────────────────────────────────────────
Write-Step "Building ZomCleaner..."

$env:GOOS   = "windows"
$env:GOARCH = "amd64"

try {
    & go build -o $BinaryName $SourceFile
    if ($LASTEXITCODE -ne 0) { throw "Build exited with code $LASTEXITCODE" }
    Write-OK "Build complete -> $BinaryName"
} catch {
    Write-Fail "Build failed: $_"
    Pause; exit 1
} finally {
    Remove-Item Env:\GOOS   -ErrorAction SilentlyContinue
    Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
}

# ─── Step 4: Desktop shortcut ─────────────────────────────────────────────────
Write-Step "Creating desktop shortcut..."

try {
    $desktop = [Environment]::GetFolderPath("Desktop")
    $ws  = New-Object -ComObject WScript.Shell
    $lnk = $ws.CreateShortcut("$desktop\ZomCleaner.lnk")
    $lnk.TargetPath      = "$PWD\$BinaryName"
    $lnk.WorkingDirectory = "$PWD"
    $lnk.Description     = "ZomCleaner - System Resource Reaper"
    $lnk.Save()
    Write-OK "Shortcut created on Desktop: ZomCleaner.lnk"
} catch {
    Write-Warn "Could not create shortcut (non-critical): $_"
}

# ─── Step 5: Launch dashboard ─────────────────────────────────────────────────
Write-Step "Launching ZomCleaner..."
Write-Host ""
Write-Host "  Dashboard -> $DashboardURL" -ForegroundColor Green
Write-Host "  Press Ctrl+C to stop the server." -ForegroundColor DarkGray
Write-Host ""

# Open browser after 2 seconds (background job)
Start-Job -ScriptBlock {
    Start-Sleep -Seconds 2
    Start-Process $using:DashboardURL
} | Out-Null

# Run binary in foreground
& ".\$BinaryName"

# Use "Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser" command 
