# iBinds

iBinds edits iRacing control bindings in bulk and copies calibration between
control profiles — without opening the simulator. It reads the local
`controls.cfg` (binary) and `joyCalib.yaml` (YAML) files under
`Documents/iRacing/profiles/controls/`, applies surgical edits to only the
target record's fields, and keeps everything else byte-for-byte intact.

## Features

- **Bulk binding**: pick a control and a key or device button, select any set
  of car profiles (baseline included), preview a per-profile diff, and apply.
- **Conflict detection**: warns when the key/button is already bound to another
  control in a profile (iRacing silently resolves such conflicts on load — iBinds
  surfaces them first).
- **Calibration copy**: copy a whole `joyCalib.yaml` from one profile to many.
  Calibration values are never edited.
- **Safety**: writes are blocked while the simulator is running
  (`iRacingSim64DX11.exe`), a timestamped backup is made before every write, and
  the `CCFG` magic is validated before any edit.
- **Local only**: a single binary starts an HTTP server on `127.0.0.1` with a
  per-session token and opens the browser with an embedded Svelte UI.

## Requirements

- Windows (iRacing is Windows-only).
- To run: just `ibinds.exe` (no runtime, no install).
- To build from source: Go 1.22+ and Node.js 20+.

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
```

The profiles directory is resolved in this order: `--dir` flag,
`IBINDS_PROFILES_DIR` environment variable, the saved config file, or the
default `Documents/iRacing/profiles/controls`.

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

## Build

```powershell
.\build.ps1   # produces ibinds.exe (builds web/ then the Go binary)
```

## Binary format

`controls.cfg` decoding is documented in [`docs/FORMAT.md`](docs/FORMAT.md),
validated against real fixtures captured from the author's profiles.

## License

MIT — see [LICENSE](LICENSE).
