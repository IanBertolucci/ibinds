# Build script for iBinds on Windows.
# Builds the embedded Svelte SPA, then compiles the Go binary.
Set-Location $PSScriptRoot

Push-Location web
npm install --legacy-peer-deps
npm run build
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
Pop-Location

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags "-s -w" -o ibinds.exe ./cmd/ibinds
