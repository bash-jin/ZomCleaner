# =============================================================================
# ZomCleaner — setup.ps1 (STRICT SYNTAX CLEANUP)
# =============================================================================

# 1. Admin Check
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]"Administrator")

if (-not $isAdmin) {
    Write-Host "Relaunching as Admin..." -ForegroundColor Yellow
    Start-Process powershell.exe -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`"" -Verb RunAs
    exit
}

# 2. Config
$SourceFile  = "clean1.go"
$BinaryName  = "zomcleaner.exe"
$DashboardURL = "http://localhost:8080"
$WorkDir = Split-Path -Parent $MyInvocation.MyCommand.Definition

# 3. Helpers
function Write-OK($msg)   { Write-Host "[ OK ] $msg" -ForegroundColor Green }
function Write-Step($msg) { Write-Host "`n▶ $msg" -ForegroundColor White }
function Write-Fail($msg) { Write-Host "[ FAIL ] $msg" -ForegroundColor Red }

# 4. Check Go
Write-Step "Checking Go..."
if (Get-Command go -ErrorAction SilentlyContinue) {
    Write-OK "Go is ready."
} else {
    Write-Host "Installing Go via winget..." -ForegroundColor Cyan
    winget install --id GoLang.Go -e --silent --accept-package-agreements --accept-source-agreements
    $env:PATH += ";C:\Program Files\Go\bin;$env:USERPROFILE\go\bin"
}

# 5. Build
Write-Step "Building Binary..."
Set-Location $WorkDir
if (Test-Path $SourceFile) {
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    & go build -o $BinaryName $SourceFile
    if ($LASTEXITCODE -eq 0) {
        Write-OK "Build successful: $BinaryName"
    } else {
        Write-Fail "Go build failed."; Pause; exit
    }
} else {
    Write-Fail "Missing $SourceFile in $WorkDir"; Pause; exit
}

# 6. Shortcut
Write-Step "Creating Shortcut..."
try {
    $desktop = [Environment]::GetFolderPath("Desktop")
    $ws = New-Object -ComObject WScript.Shell
    $lnk = $ws.CreateShortcut("$desktop\ZomCleaner.lnk")
    $lnk.TargetPath = "$WorkDir\$BinaryName"
    $lnk.WorkingDirectory = "$WorkDir"
    $lnk.Save()
    Write-OK "Shortcut created."
} catch {
    Write-Host "Shortcut failed, but that's okay." -ForegroundColor Gray
}

# 7. Launch
Write-Step "Launching..."
Start-Job -ScriptBlock { Start-Sleep -Seconds 2; Start-Process "http://localhost:8080" } | Out-Null
& ".\$BinaryName"
