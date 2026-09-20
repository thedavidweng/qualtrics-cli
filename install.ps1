# qualtrics-cli PowerShell installer
# Usage: irm https://raw.githubusercontent.com/thedavidweng/qualtrics-cli/main/install.ps1 | iex

$ErrorActionPreference = 'Stop'

$Repo = "thedavidweng/qualtrics-cli"
$Binary = "qualtrics.exe"

function Step($msg) { Write-Host "==> $msg" -ForegroundColor Cyan }
function Die($msg)  { Write-Error "ERROR: $msg"; exit 1 }

$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "x86_64" }
    "ARM64" { "arm64" }
    default { Die "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

Step "Resolving latest version..."
$ReleaseUrl = "https://api.github.com/repos/$Repo/releases/latest"
$Release = Invoke-RestMethod -Uri $ReleaseUrl -Headers @{ "User-Agent" = "qualtrics-installer" }
$Tag = $Release.tag_name

if (-not $Tag) { Die "Could not resolve latest release" }

$ArchiveName = "qualtrics_windows_$Arch.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Tag/$ArchiveName"

$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

try {
    $ZipPath = Join-Path $TempDir $ArchiveName
    Step "Downloading $DownloadUrl..."
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath

    Step "Extracting..."
    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

    $InstallDir = Join-Path $env:LOCALAPPDATA "Programs\qualtrics"
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

    Move-Item -Path (Join-Path $TempDir "qualtrics.exe") -Destination (Join-Path $InstallDir $Binary) -Force

    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
        $env:Path = "$env:Path;$InstallDir"
    }

    Step "Verifying..."
    & (Join-Path $InstallDir $Binary) version

    Write-Host "`nqualtrics installed successfully to $InstallDir\$Binary" -ForegroundColor Green
}
finally {
    Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
}
