# Build script for iBinds on Windows.
# Builds the embedded Svelte SPA, then compiles the Go binary.
param(
    [string]$Version = "dev"
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

Push-Location web
try {
    if (Test-Path package-lock.json) {
        npm ci --legacy-peer-deps
    } else {
        npm install --legacy-peer-deps
    }
    if ($LASTEXITCODE -ne 0) { throw "npm install failed" }

    npm run build
    if ($LASTEXITCODE -ne 0) { throw "web build failed" }
} finally {
    Pop-Location
}

$commit = "none"
if (Test-Path .git) {
    $gitCommit = git rev-parse --short HEAD 2>$null
    if ($LASTEXITCODE -eq 0 -and $gitCommit) { $commit = $gitCommit }
}
$buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-dd")

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -trimpath `
    -ldflags "-s -w -X main.version=$Version -X main.commit=$commit -X main.buildDate=$buildDate" `
    -o ibinds.exe ./cmd/ibinds
if ($LASTEXITCODE -ne 0) { throw "Go build failed" }

Write-Host "Built ibinds.exe ($Version)"
