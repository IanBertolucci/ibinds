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

Requirements: Go 1.27.2+ (see `go.mod`) and Node.js 20.19+ (or 22.12+).

```bash
go test ./...                          # backend unit tests (run on Linux)
GOOS=windows GOARCH=amd64 go build ./...  # verify Windows-only code compiles
go vet ./...

cd web
npm install --legacy-peer-deps          # npm 10.8.x needs --legacy-peer-deps
npx vitest run                          # frontend unit tests
npm run build                           # outputs web/dist (committed + embedded)

./build.sh                              # from Linux/WSL -> ibinds.exe (no Windows shell)
VERSION=v0.1.0 ./build.sh               # optional version metadata
.\build.ps1                             # full Windows build -> ibinds.exe
.\build.ps1 -Version v0.1.0             # optional version metadata
```

Release builds are produced by `.github/workflows/release.yml` on `v*` tags;
see `docs/RELEASING.md`.

`build.sh` finds Go via `$PATH` or `~/.local/go/bin/go` (override with
`GO_BIN=/path/to/go`). If a copy already exists at `%USERPROFILE%\ibinds.exe`
(the one the user double-clicks), it is refreshed too; a running instance is
stopped first. Agents on WSL should use it instead of switching to Windows.

Run locally: `go run ./cmd/ibinds --token dev --port 8791` (backend) +
`npm run dev` in `web/` (Vite proxies `/api` to 8791; token `dev` matches the
`import.meta.env.DEV` fallback).

## Logs

Every run logs to stderr and appends to `%APPDATA%\ibinds\ibinds.log`
(Windows) or `$XDG_CONFIG_HOME/ibinds/ibinds.log` (Linux); override with
`--log PATH`. The log rotates at 5 MB to `ibinds.log.old`. Each API request is
one line: method, path, status, duration,
and the error body for 4xx/5xx. Query strings (session token) are never
logged. From WSL, read the Windows log at
`/mnt/c/Users/<user>/AppData/Roaming/ibinds/ibinds.log`; this is the fastest
way to see a failure the user reports.

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
| `internal/config`, `internal/profiles` | config persistence, profile scan |
| `internal/labels` | friendly names, key labels, in-game category groups |

Windows-only code is isolated behind `//go:build windows` with a matching
stub for other platforms, so tests run on Linux.

## The binary format is the source of truth

`docs/FORMAT.md` documents `controls.cfg` — read it before touching the
parser/editor. It was reverse-engineered from **real fixtures**
(`internal/cfgparse/testdata/*.cfg`) captured from the author's profiles, and
the tests assert byte-exact reproduction of the simulator's own edits. Do NOT
guess the format; if a fixture reveals a new structure, update FORMAT.md and
the tests to match reality (the fixtures win over any assumption).

Keyboard codes are iRacing's own `Key` enum, not ASCII or Windows virtual-key
codes; see the key-code table in FORMAT.md. `internal/labels.Key` is the display
mapping used by the API.

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
  `LRTC` chunk marker; the binding type is a uint32 at `+8`, key code at `+12`
  (low byte read by iBinds; modifiers at `+16`), five 32-bit button-mask words
  at `+12..+31`, GUID at `+36`, ProductID at `+52`, `PIDVID` at `+62` (see
  FORMAT.md).
- **The control list follows the in-game section order, not file order.**
  `labels.ControlGroupFor(name)` supplies `Category`, `Subcategory`, and
  `Order`; the API exposes `category`/`subcategory` and sorts by `Order`.
  Category/subcategory dividers and section search depend on this. Unknown
  controls fall back to `Other` and sort last.
- **Special keys are iRacing key codes, not ASCII.** Use `labels.Key(code)`
  for display (F1–F12, numpad, Pause, Print Screen, punctuation, etc.);
  unknown non-printables render as `0xNN`, never a replacement character.
- **Default device is persisted in config.** `Config.DefaultDeviceInstanceGUID`
  is read/written by `GET/POST /api/settings`; a POST without the field
  preserves the current value. The binding wizard preselects it, falling back
  to the first device if it is not connected.
- **Edit vs New binding profile selection.** Editing a control from a profile
  preselects only that profile; the global "New binding" preselects all.
- **Desktop layout is viewport-height with internal scrolling.** At >=1101px,
  `main` is `height:100vh`/`overflow:hidden`, the layout row is
  `minmax(0,1fr)`, and the sidebar and controls table scroll internally so the
  "Copy selected" footer stays visible. Do not reintroduce page-level scroll.
- **Search matches sections too.** `ControlsTable` filters on name, display
  name, category, and subcategory; keep category search working.
- **Test fixtures are committed real files.** Don't replace them with
  synthetic guesses.
- **Per-profile current bindings load for every profile.** The home screen
  must fetch `/api/profiles/{id}/controls` for all profiles, not only the
  selected ones, and must not gate on `hasControls`; otherwise unselected
  profiles wrongly render `none`. Covered by `web/src/App.test.ts`.
- **Calibration copy excludes the source profile.** Targets are
  `profiles - source`, so "Copy calibration" never copies a profile onto
  itself. `CalibCopy.svelte` also has an "All targets" toggle.
- **The running exe may be a copy at `%USERPROFILE%\ibinds.exe`, not the
  repo binary.** Served assets are embedded at build time, so rebuilding
  `web/dist` alone does not change a built exe. `build.sh` stops a running
  instance and refreshes that copy; compare sha256/timestamp before
  concluding a fix did not apply.

### DirectInput capture (Windows)

- `DirectInput8Create` requires a real HINSTANCE (`GetModuleHandleEx`);
  passing 0 returns `0x80070057` (E_INVALIDARG).
- `SetCooperativeLevel` requires a real top-level HWND (the console window,
  with the desktop shell as fallback); 0 returns `0x80070006` (E_HANDLE).
- `c_dfDIJoystick2` lives in `dinput8.lib`, not in `dinput8.dll`. The code
  builds the equivalent button-only `DIDATAFORMAT` in Go; do not use
  `NewProc("c_dfDIJoystick2")` (it panics).
- `Acquire()` returns `S_FALSE` (1) when the device is already acquired.
  Treat it as success, otherwise the capture loop never reads state.
- **POV hats (sticks) are not buttons.** iRacing maps the four hats to
  virtual buttons `128..159`, eight directions each starting at up and
  going clockwise; `povButton` applies that mapping. For example, MOZA R9
  reports 128 physical buttons plus one POV hat.
- The `DIDATAFORMAT` must declare POV objects (`DIDFT_OPTIONAL |
  DIDFT_POV | DIDFT_ANYINSTANCE`); button objects use a NULL GUID. Capture
  triggers on a rising edge from the state sampled at start, so a held
  stick/button is not captured immediately.
- Debug capture from WSL with `./ibinds.exe --capture` (waits 10s for a
  button) and read the log; the API logs the full 5xx error body.

## Known limitations / open items

- DirectInput capture compiles for Windows but the button-index conversion
  vs iRacing's "Dev N Btn M" has not been validated on hardware.
- Friendly car-name map (`internal/labels`) only has `f4`; control names
  are curated from the in-game list in `internal/labels/control_labels.go`,
  and unmapped controls fall back to their technical id.
- Device enumeration falls back to the profiles' `joyCalib.yaml` on
  non-Windows (button count unknown/0).
- `internal/labels/control_categories.go` and `keys.go` are snapshots of the
  in-game controls list/key codes captured in Oct 2026. Unknown controls fall
  back to `Other`/technical id; unknown non-printable keys render as hex.

## Process

New features should follow a fresh brainstorming → spec → plan cycle (the
superpowers workflow), with the plan citing `docs/FORMAT.md` as the format
authority and this file for the invariants. Small fixes and bug fixes should
still be test-driven and scoped to the behavior being changed; they do not need
a new spec/plan. All code, comments, docs, and commit messages are in English.
