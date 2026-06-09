$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
$ToolsDir = Join-Path $Root ".tools"
$GoRoot = Join-Path $ToolsDir "go"
$GoExe = Join-Path $GoRoot "bin\go.exe"
$DistDir = Join-Path $Root "dist"
$OutFile = Join-Path $DistDir "Hypnos.exe"

if (-not (Test-Path $GoExe)) {
    $SystemGo = Get-Command "go.exe" -ErrorAction SilentlyContinue
    if ($SystemGo) {
        $GoExe = $SystemGo.Source
    }
}

if (-not (Test-Path $GoExe)) {
    New-Item -ItemType Directory -Force -Path $ToolsDir | Out-Null

    Write-Host "Go nicht gefunden. Lade portable Go-Toolchain..."
    $Release = Invoke-RestMethod -Uri "https://go.dev/dl/?mode=json" |
        Where-Object { $_.stable -eq $true } |
        Select-Object -First 1

    $Archive = $Release.files |
        Where-Object { $_.os -eq "windows" -and $_.arch -eq "amd64" -and $_.kind -eq "archive" } |
        Select-Object -First 1

    if (-not $Archive) {
        throw "Kein passendes Go Windows amd64 Archiv gefunden."
    }

    $ZipFile = Join-Path $ToolsDir $Archive.filename
    $Curl = Get-Command "curl.exe" -ErrorAction SilentlyContinue
    if ($Curl) {
        & $Curl.Source -L --fail --retry 3 --connect-timeout 30 -o $ZipFile "https://go.dev/dl/$($Archive.filename)"
    } else {
        Invoke-WebRequest -Uri "https://go.dev/dl/$($Archive.filename)" -OutFile $ZipFile
    }

    $ActualHash = (Get-FileHash -Algorithm SHA256 -Path $ZipFile).Hash.ToLowerInvariant()
    if ($ActualHash -ne $Archive.sha256) {
        throw "SHA256 passt nicht fuer $($Archive.filename)."
    }

    Expand-Archive -Path $ZipFile -DestinationPath $ToolsDir -Force
}

New-Item -ItemType Directory -Force -Path $DistDir | Out-Null

$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

& $GoExe build `
    -trimpath `
    -buildvcs=false `
    -ldflags "-s -w -buildid= -H=windowsgui" `
    -o $OutFile `
    ".\cmd\hypnos"

$Size = [Math]::Round((Get-Item $OutFile).Length / 1MB, 2)
Write-Host "Fertig: dist\Hypnos.exe ($Size MB)"
