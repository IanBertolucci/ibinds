# AGENTS.md

Guidance for AI agents working on this repository. Read this before making
changes; it captures the non-obvious rules that the code and tests depend on.

## What this is

iBinds: a single Windows binary (Go) + embedded Svelte 5 SPA that edits iRacing
control bindings in bulk and copies calibration between control profiles —
without the simulator open. It parses the binary `controls.cfg` and the
`joyCalib.yaml` text, applies byte-exact surgical edits, backs up before every
write, and exposes a local HTTP API.

## Commands

```bash
go test ./...                          # backend unit tests (run on Linux)
GOOS=windows GOARCH=amd64 go build ./...  # verify Windows-only code compiles
go vet ./...

cd web
npm install --legacy-peer-deps          # npm 10.8.x needs --legacy-peer-deps
npx vitest run                          # frontend unit tests
npm run build                           # outputs web/dist (committed + embedded)

.\build.ps1                             # full Windows build -> ibinds.exe
```

Run locally: `go run ./cmd/ibinds --token dev --port 8791` (backend) +
`npm run dev` in `web/` (Vite proxies `/api` to 8791; token `dev` matches the
`import.meta.env.DEV` fallback).

## Architecture

| Package | Role |
|---|---|
| `internal/cfgparse` | read-only `controls.cfg` parser (magic + record walk) |
| `internal/editor` | surgical writer (`ApplyBinding`) + conflict detection |
| `internal/calib` | `joyCalib.yaml` parser/validation |
| `internal/ops` | pure plan logic (`PlanBinding`) |
| `internal/server` | HTTP API (127.0.0.1 + per-session token) |
| `internal/fileops` | timestamped backup, byte diff, atomic write |
| `internal/inputcap` | DirectInput button capture (Windows) + enumeration |
| `internal/simcheck` | iRacing process check |
| `internal/config`, `internal/profiles`, `internal/labels` | config, profile scan, friendly names |

Windows-only code is isolated behind `//go:build windows` with a matching
stub for other platforms, so tests run on Linux.

## The binary format is the source of truth

`docs/FORMAT.md` documents `controls.cfg` — read it before touching the
parser/editor. It was reverse-engineered from **real fixtures**
(`internal/cfgparse/testdata/*.cfg`) captured from the author's profiles, and
the tests assert byte-exact reproduction of the simulator's own edits. Do NOT
guess the format; if a fixture reveals a new structure, update FORMAT.md and
the tests to match reality (the fixtures win over any assumption).

## Invariants (do not violate)

- **Surgical editing**: change only the target record's fields; every other
  byte stays identical. Verified by `TestApplyPreservesUnknownBytes` and the
  byte-exact reproduction tests.
- **Never edit calibration values**: `joyCalib.yaml` is copied whole, never
  modified.
- **Simulator closed before write**: the `iRacingSim64DX11.exe` process must
  not be running; the API returns 409 and touches nothing.
- **Backup before every write**: timestamped `.bak-YYYYMMDD-HHMMSS` first.
- **Validate `CCFG` magic** before writing.

## Gotchas

- **Go nil slices marshal to JSON `null`.** Always `make([]T, 0)` for fields
  that appear in API responses (e.g. `conflicts`); the frontend reads `.length`.
- **Conflict detection is namespace-scoped.** Replay/camera controls
  (`Cam*`, `DCam*`, `Rpy*`, `NextRpyCam*`, `PrevRpyCam*`) are a separate
  namespace from driving controls and are excluded from conflicts. A control
  may bind **multiple** buttons (`Record.Buttons`); `Record.Button` is the
  lowest one.
- **The parser walks records as `name + NUL + 68-byte fixed region`** from a
  `LRTC` chunk marker; the binding type is a uint32 at `+8`, key byte at `+12`,
  five 32-bit button-mask words at `+12..+31`, GUID at `+36`, ProductID at
  `+52`, `PIDVID` at `+62` (see FORMAT.md).
- **Test fixtures are committed real files.** Don't replace them with
  synthetic guesses.

## Known limitations / open items

- DirectInput capture compiles for Windows but the button-index conversion
  vs iRacing's "Dev N Btn M" has not been validated on hardware.
- Friendly car-name map (`internal/labels`) only has `f4` and
  `Low Fuel Accept`; the rest fall back to their technical slug.
- Device enumeration falls back to the profiles' `joyCalib.yaml` on
  non-Windows (button count unknown/0).

## Process

New features should follow a fresh brainstorming → spec → plan cycle (the
superpowers workflow), with the plan citing `docs/FORMAT.md` as the format
authority and this file for the invariants. All code, comments, docs, and
commit messages are in English.
