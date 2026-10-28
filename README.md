# iBinds

[![CI](https://github.com/IanBertolucci/ibinds/actions/workflows/ci.yml/badge.svg)](https://github.com/IanBertolucci/ibinds/actions/workflows/ci.yml)

iBinds edits iRacing control bindings in bulk and copies calibration between
control profiles — without opening the simulator. It reads the local
`controls.cfg` (binary) and `joyCalib.yaml` (YAML) files under
`Documents/iRacing/profiles/controls/`, applies surgical edits to only the
target record's fields, and keeps everything else byte-for-byte intact.

## Features

- **Profile-centric UI**: browse every profile, search controls, and see the
  current key or button binding — including multi-button bindings.
- **Bulk binding**: pick a control and a key or device button, select any set
  of profiles (baseline included), preview a per-profile diff, and apply.
- **Copy bindings**: select controls and copy the exact binding region to other
  profiles. Multi-button masks are preserved.
- **Friendly names**: control labels use the same names shown in the iRacing
  controls screen where a mapping is known; unmapped controls keep their
  technical id.
- **Conflict detection**: warns when the key/button is already bound to another
  control in a profile (iRacing silently resolves such conflicts on load —
  iBinds surfaces them first).
- **Calibration copy**: preview and copy a whole `joyCalib.yaml` from one
  profile to many. Calibration values are never edited.
- **Safety**: writes are blocked while the simulator is running
  (`iRacingSim64DX11.exe`), a timestamped backup is made before every write, and
  the `CCFG` magic is validated before any edit.
- **Local only**: a single binary starts an HTTP server on `127.0.0.1` with a
  per-session token and opens the browser with an embedded Svelte UI.

## Install (Windows)

1. Open the [latest release](https://github.com/IanBertolucci/ibinds/releases/latest).
2. Download `ibinds.exe` and `SHA256SUMS.txt`.
3. Verify the download:

   ```powershell
   Get-FileHash .\ibinds.exe -Algorithm SHA256
   Get-Content .\SHA256SUMS.txt
   ```

   The two hashes must match.
4. Run `ibinds.exe` (double-click or from PowerShell). It starts a local server
   and opens the UI in your browser.

iBinds is portable: no installer, no administrator rights, and no runtime to
install. It only reads and writes the iRacing profiles folder on your machine.

### Windows SmartScreen

The binary is not code-signed, so Windows SmartScreen may show a warning. Do
not disable SmartScreen. Verify the SHA-256 checksum and, if you use the GitHub
CLI, the build provenance:

```powershell
gh attestation verify .\ibinds.exe --repo IanBertolucci/ibinds
```

Every release is built by GitHub Actions from the tagged source. If you prefer
not to trust the prebuilt binary, build it yourself from the source archive.

## Build from source

Requirements:

- Windows 10/11 (iRacing is Windows-only).
- Go 1.27+ (see `go.mod`).
- Node.js 20+.

Download the source zip from the release (or clone the repository and check out
the tag), then run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\build.ps1
```

The script installs the web dependencies, builds the SPA, and cross-compiles
`ibinds.exe` in the repository root. To include a version string:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\build.ps1 -Version v0.1.0
```

From Linux or WSL:

```bash
./build.sh                 # builds web/ then cross-compiles ibinds.exe
VERSION=v0.1.0 ./build.sh  # optional version metadata
```

`build.sh` finds Go via `$PATH` or `~/.local/go/bin/go` (override with
`GO_BIN=/path/to/go`). If a copy already exists at `%USERPROFILE%\ibinds.exe`
(the one the user double-clicks), it tries to refresh it too.

## Usage

Double-click `ibinds.exe` (or run from a terminal). It starts a local server,
prints a URL, and opens your browser. The token in the URL authenticates the
session.

CLI flags:

```
--dir DIR           override the profiles directory (default: auto-detected)
--list-profiles     list detected profiles and exit
--token TOKEN       fix the session token (default: random per run)
--port PORT         port to listen on (default: 0 = ephemeral)
--capture           capture a button press and exit (DirectInput test)
--no-browser        do not open the browser on startup
--version           print version and exit
--log FILE          log file path (default: user config dir/ibinds/ibinds.log)
```

The profiles directory is resolved in this order: `--dir` flag,
`IBINDS_PROFILES_DIR` environment variable, the saved config file, or the
default `Documents/iRacing/profiles/controls`.

Logs go to stderr and the log file (on Windows,
`%APPDATA%\ibinds\ibinds.log`). Each API request is logged with method, path,
status, duration, and the error body for 4xx/5xx; the session token is never
logged.

## Development

Backend (starts the server on port 8791 with a fixed `dev` token):

```bash
go run ./cmd/ibinds --token dev --port 8791
```

Frontend (Vite dev server, proxies `/api` to the backend above):

```bash
cd web
npm install --legacy-peer-deps   # npm 10.8.x needs --legacy-peer-deps
npm run dev
```

Tests:

```bash
go test ./...            # backend unit tests (run on Linux)
cd web && npx vitest run # frontend unit tests
```

The same checks run in GitHub Actions on every push and pull request.

## Releases

The release workflow `.github/workflows/release.yml` runs when a `v*` tag is
pushed. It builds and tests both sides, cross-compiles `ibinds.exe` with version
metadata, writes `SHA256SUMS.txt`, attests build provenance, and publishes:

- `ibinds.exe`
- `SHA256SUMS.txt`
- GitHub's automatic `Source code (zip)`

See [`docs/RELEASING.md`](docs/RELEASING.md) for the maintainer checklist.

## Binary format

`controls.cfg` decoding is documented in [`docs/FORMAT.md`](docs/FORMAT.md),
validated against real fixtures captured from the author's profiles.

## License

MIT — see [LICENSE](LICENSE).
