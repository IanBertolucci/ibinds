# iBinds Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Windows app that edits bindings in bulk and copies calibration between iRacing control profiles, with the simulator closed, automatic backups, diff preview and conflict detection.

**Architecture:** Single Go binary that starts an HTTP server on `127.0.0.1` (per-session token) and opens the browser with a Svelte SPA embedded via `go:embed`. The backend parses `controls.cfg` (binary) and `joyCalib.yaml`, performs surgical edits (only the target record's fields), timestamped backups, diffs, conflict detection, simulator process checks, and button capture via DirectInput (Windows).

**Tech Stack:** Go 1.22+, Svelte 5 + Vite + Vitest, `gopkg.in/yaml.v3`, `golang.org/x/sys` (Windows Known Folder API), `go-ole` (Windows-only, DirectInput vtables hand-written), stdlib `net/http` + `embed`.

**Spec:** `docs/superpowers/specs/2026-10-18-ibinds-design.md`

## Global Constraints

- Target: Windows (`GOOS=windows`); Windows-only code isolated in files with build tags (`//go:build windows`) and a matching stub for other platforms (dev/test runs on Linux/WSL).
- Writes forbidden while the simulator is open: check the `iRacingSim64DX11.exe` process before any write; return a clear error when running.
- Surgical editing: change only the target record's fields; every byte outside the target fields stays identical (including unknown structures).
- Validate the `CCFG` magic and file consistency before any write.
- Never edit calibration values: `joyCalib.yaml` is only copied whole.
- Timestamped backup required before every write.
- HTTP server only on `127.0.0.1`, ephemeral port, per-session one-time token validated on every `/api/` route; `--token` flag fixes a token for development.
- Go module: `ibinds`. No heavy backend dependencies beyond the listed ones.
- Frontend: Svelte 5 + Vite; build output in `web/dist` committed to the repo (embedded via `go:embed`); `web/.gitignore` must not ignore `dist`.
- Unit tests must run on Linux; parser/writer/conflict tests are driven by **real** fixtures captured from the author's iRacing profiles (committed with the author's consent), not synthetic guesses.
- All code, comments, docs and commit messages in English.

## Review Focus

1. **`controls.cfg` with invalid magic or truncated** → parser returns a typed error; writer refuses. Tests in Task 2 (`TestParseRejectsBadMagic`, `TestParseRejectsTruncated`) and Task 4 (`TestWriteRefusesUnvalidatedInput`).
2. **Simulator running at write time** → API returns 409 and no file is touched. Test in Task 8 (`TestApplyBlockedWhenSimRunning`).
3. **Key/button already used by another function in the same profile** → apply plan reports a conflict; applying without the force flag skips the profile. Tests in Task 5 (`TestDetectKeyConflict`, `TestDetectButtonConflictSameDevice`, `TestDetectButtonConflictDifferentDeviceIsNotConflict`) and Task 8 (`TestApplyWithoutForceSkipsConflict`).
4. **Binding's device missing from the target profile's `joyCalib.yaml`** → warning in the plan; never auto-edit calibration. Test in Task 8 (`TestPlanWarnsDeviceMissing`).
5. **Unknown bytes/structures in the file** → preserved byte-for-byte on write. Tests in Task 4 (`TestApplyPreservesUnknownBytes`, `TestApplyButtonReplacementReproducesRealFixture`).

---

### Task 1: Scaffold, config and profile enumeration

**Files:**
- Create: `go.mod`, `cmd/ibinds/main.go`, `internal/config/config.go`, `internal/config/config_windows.go`, `internal/config/config_other.go`, `internal/profiles/profiles.go`, `internal/labels/labels.go`, `LICENSE` (MIT)
- Test: `internal/config/config_test.go`, `internal/profiles/profiles_test.go`, `internal/labels/labels_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `config.Load() (*Config, error)`; `type Config struct { ProfilesDir string }` — priority: `--dir` flag > `IBINDS_PROFILES_DIR` env > `%APPDATA%/ibinds/config.json` file > default resolved from the user's Documents folder.
  - `config.Save(c *Config) error` — writes JSON to the user config dir.
  - `config.DocumentsDir() (string, error)` — on Windows (`config_windows.go`, `//go:build windows`) resolves via `SHGetKnownFolderPath(FOLDERID_Documents)` from `golang.org/x/sys/windows` (handles OneDrive-redirected Documents); other platforms (`config_other.go`) return `$HOME/Documents`. The default profiles dir is `DocumentsDir() + "/iRacing/profiles/controls"`.
  - `profiles.Scan(dir string) ([]Profile, error)`; `type Profile struct { ID string; Dir string; IsBaseline bool; HasControls, HasCalib bool }` — `ID` = folder name (`Baseline` or car slug); a valid profile has `controls.cfg` or `joyCalib.yaml`.
  - `labels.Car(id string) string`, `labels.Control(id string) string` — static maps falling back to the technical id (maps start empty; populated in Task 10).

- [ ] **Step 1: Write failing tests** — `internal/config/config_test.go`:
  - `TestDefaultProfilesDir`: with no flag/env/file, `Load` returns `<DocumentsDir()>/iRacing/profiles/controls` and `Save` writes JSON to the user config dir (use `t.Setenv` to isolate HOME/config dir; the non-Windows `DocumentsDir` is testable on Linux via `$HOME/Documents`).
  - `TestEnvOverride` and `TestFileOverride`: `IBINDS_PROFILES_DIR` and the config file override the default; `TestFlagOverride`: the `--dir` flag wins over everything.
  - `internal/profiles/profiles_test.go`: `TestScanListsBaselineAndCars` — temp dir with `Baseline/controls.cfg`, `f4/controls.cfg`, `f4/joyCalib.yaml`, `audigto/joyCalib.yaml` (calib only) and `junk/` (empty) → 3 profiles, `f4` with `HasControls=true, HasCalib=true`, `audigto` with `HasControls=false, HasCalib=true`, `junk` absent; `TestScanSortsBaselineFirst` — `Baseline` is the first item.
  - `internal/labels/labels_test.go`: `TestCarFallback` — unknown id returns the id itself; `TestControlFallback` same. (Known-mapping tests come in Task 10, when the maps are populated.)

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./...`
Expected: FAIL (packages do not exist).

- [ ] **Step 3: Implement `go.mod`, `config`, `profiles`, `labels`**

`go mod init ibinds`; `go get golang.org/x/sys`. `config.Load` uses `os.UserConfigDir()` for the config file and `config.DocumentsDir()` for the default. `profiles.Scan` uses `os.ReadDir`, ignores folders with neither file, sorts with `Baseline` first and the rest alphabetically. `labels` starts with empty maps + fallback.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./...`
Expected: PASS.

- [ ] **Step 5: Scaffold `cmd/ibinds/main.go` + `LICENSE` and commit**

`main.go` with `--dir` and `--list-profiles` flags (prints profiles and exits; otherwise a `println("ok")` placeholder). `LICENSE` file (MIT).

```bash
git add -A && git commit -m "feat: scaffold, config, profile enumeration"
```

---

### Task 2: Format-mapping milestone + read-only controls.cfg parser

**Files:**
- Create: `internal/cfgparse/parser.go`, `internal/cfgparse/records.go`, `internal/cfgparse/testdata/*.cfg` (real fixtures), `docs/FORMAT.md`
- Test: `internal/cfgparse/parser_test.go`

**Interfaces:**
- Consumes: `profiles.Profile.Dir` (Task 1) only in integration; the parser itself takes `[]byte`.
- Produces:
  - `cfgparse.Parse(data []byte) (*File, error)`; `type File struct { Raw []byte; Records []Record }`; `var ErrBadMagic = errors.New("cfgparse: bad magic")`, `ErrTruncated = errors.New("cfgparse: truncated")`.
  - `type Record struct { Name string; Offset int; PayloadStart int; Kind BindingKind; Key byte; Button int; Device *Device; Opaque []byte }`; `type BindingKind int` with constants `KindUnbound`, `KindKey`, `KindButton`, `KindAxis`, `KindUnknown`; `type Device struct { InstanceGUID string; ProductGUID string }` (`{...}` format as in `joyCalib.yaml`).
  - `Record.Bytes() []byte` — reserialization identical to the original (identity).

- [ ] **Step 1: Copy real fixtures into `internal/cfgparse/testdata/`**

The author's experiment captured before/after pairs written by the real simulator (author's consent recorded in the session). Copy, with these names:
- From `/tmp/ibinds_exp/`: `f4_step1_before.cfg` → `testdata/f4_unbound_oldfmt.cfg` (June-era layout), `f4_step2_before.cfg` → `testdata/f4_btn130.cfg`, `f4_step3_before.cfg` → `testdata/f4_btn128.cfg`, `f4_rt_before.cfg` → `testdata/f4_key_u.cfg`.
- From the live profile `/mnt/c/Users/Ian/Documents/iRacing/profiles/controls/f4/controls.cfg` → `testdata/f4_key_i_lowfuel_n.cfg` (current state: `TractionControlArm`='I', `LowFuelAccept`='N').
If any source is missing, stop and ask the author to re-snapshot before continuing. Fixture names and paths are part of the interface.

- [ ] **Step 2: Write FORMAT.md** (documentation, no code)

Record the validated findings: `CCFG` magic at `0x0` (`47 46 43 43`), size field at `0x8` = len−12, `CTRL` chunk at `0xa4`, records with NUL-terminated names, binding type (`02` button, `04` key, `00` unbound; axes observed with `0x19`/`0x1f`), ASCII key byte, button bitmask (button N = bit `N%32` of word `N/32`, word 4 at `0xbe0` for Btn 128/130), InstanceGUID 16 bytes + ProductID 4 bytes + `PIDVID`. Include the `TractionControlArm` offset table as the annotated example, and a table of the fixture files with the known value of each field (which byte differs between `f4_btn130.cfg` and `f4_btn128.cfg`, etc.).

- [ ] **Step 3: Write failing tests** — `internal/cfgparse/parser_test.go` (real fixtures drive every happy-path test):
  - `TestParseRejectsBadMagic`: `[]byte("XXXX...")` → `ErrBadMagic`.
  - `TestParseRejectsTruncated`: `testdata/f4_key_u.cfg` cut in the middle → `ErrTruncated`.
  - `TestParseRealBtn130`: `f4_btn130.cfg` → record `TractionControlArm` with `Kind==KindButton`, `Button==130`, `Device.InstanceGUID=="{991805D0-2EB6-11F1-800A-444553540000}"`.
  - `TestParseRealBtn128`: `f4_btn128.cfg` → same record with `Button==128`.
  - `TestParseRealKeyU`: `f4_key_u.cfg` → `Kind==KindKey`, `Key=='U'`.
  - `TestParseRealCurrent`: `f4_key_i_lowfuel_n.cfg` → `TractionControlArm` `Key=='I'` and `LowFuelAccept` `Key=='N'`; `Ignition` is `KindUnbound` (the silent-conflict outcome).
  - `TestParseOldFormatSurvives`: `f4_unbound_oldfmt.cfg` parses without error; records present; any `KindUnknown`/`KindAxis` records keep their bytes (no panic).
  - `TestRoundTripIdentity`: for each fixture, concatenation of each record's `Record.Bytes()` with the bytes outside records == `File.Raw`.

- [ ] **Step 4: Run tests to verify they fail**

Run: `go test ./internal/cfgparse/ -v`
Expected: FAIL (parser does not exist).

- [ ] **Step 5: Implement parser in `parser.go`/`records.go`**

Locate the `CTRL` chunk, walk records by NUL-terminated names, read type/key/mask/GUID at the offsets documented in FORMAT.md. `Parse` validates magic and the `0x8` size field. Any record whose framing cannot be decoded keeps its full byte range in `Opaque` with `KindUnknown`. If the real fixtures reveal that the framing assumption in FORMAT.md is wrong (tests fail in a way that requires FORMAT.md to change), update FORMAT.md to match reality — the fixtures are the source of truth, not the guesses.

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./internal/cfgparse/ -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add docs/FORMAT.md internal/cfgparse && git commit -m "feat: read-only controls.cfg parser validated against real fixtures"
```

---

### Task 3: joyCalib.yaml parsing + whole-file copy

**Files:**
- Create: `internal/calib/calib.go`, `internal/calib/testdata/f4_joycalib.yaml`, `internal/calib/testdata/baseline_joycalib.yaml` (real fixtures)
- Test: `internal/calib/calib_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `calib.Parse(data []byte) (*Calib, error)`; `type Calib struct { Devices []CalibDevice }`; `type CalibDevice struct { Name string; InstanceGUID string; ProductGUID string; Axes []CalibAxis }`; `type CalibAxis struct { Axis int; Name string; Min, Center, Max int }`.
  - `calib.Validate(data []byte) error` — YAML parseable (empty `DeviceList` is valid).

- [ ] **Step 1: Copy real `joyCalib.yaml` fixtures into `internal/calib/testdata/`**

From the author's live profiles (consent recorded): `/mnt/c/Users/Ian/Documents/iRacing/profiles/controls/f4/joyCalib.yaml` → `testdata/f4_joycalib.yaml` and `/mnt/c/Users/Ian/Documents/iRacing/profiles/controls/Baseline/joyCalib.yaml` → `testdata/baseline_joycalib.yaml`. If missing, ask the author to re-snapshot.

- [ ] **Step 2: Write failing tests** — `internal/calib/calib_test.go`, driven by the real fixtures plus synthetic cases:
  - `TestParseRealF4Fixture`: `f4_joycalib.yaml` parses; contains device `MOZA R9 Base` with `InstanceGUID == "{991805D0-2EB6-11F1-800A-444553540000}"` and the exact axis list from the file.
  - `TestParseDevicesSynthetic`: synthetic YAML (CRLF, two devices, `CalibMin/Center/Max`) → 2 devices with correct names, GUIDs and axes (exact fixture values).
  - `TestValidateAcceptsValidYAML`: `Validate(f4_joycalib.yaml)` returns nil.
  - `TestValidateAcceptsEmptyDeviceList`: YAML with `CalibrationInfo:` and no devices → nil.
  - `TestValidateRejectsGarbage`: `Validate([]byte("not: [valid"))` returns an error.

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/calib/ -v`
Expected: FAIL.

- [ ] **Step 4: Implement `calib.go` with `gopkg.in/yaml.v3`**

`go get gopkg.in/yaml.v3`. Structs with YAML tags mirroring `CalibrationInfo.DeviceList[].{DeviceName,InstanceGUID,ProductGUID,AxisList[].{Axis,AxisName,CalibMin,CalibCenter,CalibMax}}`. `Validate` = `Parse` without error.

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/calib/ -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/calib go.mod go.sum && git commit -m "feat: joyCalib.yaml parser and validation"
```

---

### Task 4: Surgical binding writer

**Files:**
- Create: `internal/editor/editor.go`
- Test: `internal/editor/editor_test.go`

**Interfaces:**
- Consumes: `cfgparse.File`, `cfgparse.Record`, `cfgparse.ErrBadMagic` (Task 2); GUIDs as `{...}` strings; real fixtures from `../cfgparse/testdata/` (read via `os.ReadFile` — tests run with the package dir as CWD).
- Produces:
  - `editor.ErrUnknownControl = errors.New("editor: unknown control")`; `editor.ErrUnsupportedKind = errors.New("editor: unsupported record kind")`.
  - `editor.ApplyBinding(f *cfgparse.File, control string, b Binding) (newData []byte, record cfgparse.Record, err error)`; `type Binding struct { Key byte; HasKey bool; Button int; HasButton bool; DeviceInstanceGUID string; ProductGUID string }` — exactly one of `HasKey`/`HasButton` true.
  - Rules: find the record by name; not found → `ErrUnknownControl`; target `Kind` is `KindAxis`/`KindUnknown` → `ErrUnsupportedKind` (only `KindUnbound`/`KindKey`/`KindButton` targets are editable); validate the magic of `File.Raw`; write type (`04` key / `02` button), the key byte or the mask bit, and GUID+PID+PIDVID (button) or zeros (key); **button edits replace the mask word wholesale — the previous button bit is cleared, not OR-ed**; every other byte unchanged.

- [ ] **Step 1: Write failing tests** — `internal/editor/editor_test.go` (real-fixture-driven):
  - `TestApplyKeyReproducesRealFixture`: `f4_btn128.cfg` + `Binding{Key:'U',HasKey:true}` on `TractionControlArm` → output byte-identical to `f4_key_u.cfg` (this is the real key→button swap the simulator performed).
  - `TestApplyButtonReplacementReproducesRealFixture`: `f4_btn130.cfg` + `Binding{Button:128,HasButton:true,DeviceInstanceGUID:"{991805D0-2EB6-11F1-800A-444553540000}",ProductGUID:"{0002346E-0000-0000-0000-504944564944}"}` on `TractionControlArm` → output byte-identical to `f4_btn128.cfg` (button→button: the word must go `4`→`1`, proving the old bit is cleared, not OR-ed).
  - `TestApplyKeyToButtonReproducesRealFixture`: `f4_key_u.cfg` + `Binding{Button:130,HasButton:true,...}` on `TractionControlArm` → output byte-identical to `f4_btn130.cfg` (key→button with a fresh mask word).
  - `TestApplyToUnbound`: `f4_key_i_lowfuel_n.cfg`'s unbound `Ignition` + key `'I'` → reparsed output has `KindKey, Key=='I'` (current-format unbound record, not the old-format fixture).
  - `TestApplyPreservesUnknownBytes`: on `f4_key_u.cfg`, apply a binding to one control and assert every byte outside the target record's edited fields is identical to the input.
  - `TestApplyUnknownControl`: nonexistent control → `ErrUnknownControl`.
  - `TestApplyRefusesAxisKind`: on a fixture record with `KindAxis` (e.g. `SteerLeft` in `f4_key_u.cfg`) → `ErrUnsupportedKind`.
  - `TestWriteRefusesUnvalidatedInput`: `File.Raw` with corrupted magic → error (not `ErrUnknownControl`).

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/editor/ -v`
Expected: FAIL.

- [ ] **Step 3: Implement `editor.go`**

Copy `File.Raw` to a new `[]byte`, apply offsets per binding type per FORMAT.md, return the new array and the reparsed `Record`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/editor/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/editor && git commit -m "feat: surgical binding writer with byte-exact preservation"
```

---

### Task 5: Conflict detection

**Files:**
- Create: `internal/editor/conflicts.go`
- Test: `internal/editor/conflicts_test.go`

**Interfaces:**
- Consumes: `cfgparse.File` (Task 2); real fixtures.
- Produces:
  - `editor.FindConflicts(f *cfgparse.File, target string, b Binding) []Conflict`; `type Conflict struct { Control string; Reason string }` — `Reason` like `"key 'U' already bound to Ignition"` or `"button 128 of device {...} already bound to PushToPass"`.
  - Rules: a key conflicts with any other control with `KindKey` and the same `Key`; a button conflicts with another control with `KindButton`, the same `Button` and the same `DeviceInstanceGUID`; the target control itself is ignored; a different device with the same button is not a conflict.

- [ ] **Step 1: Write failing tests** (fixtures from Task 2):
  - `TestDetectKeyConflict`: on `f4_key_u.cfg`, target `Ignition` with binding `'U'` → 1 conflict `{Control:"TractionControlArm"}`.
  - `TestDetectNoConflictWhenUnbound`: target `Ignition` with key `'Z'` → 0 conflicts.
  - `TestDetectButtonConflictSameDevice`: on `f4_btn128.cfg`, a synthetic second control (mutate a copy of the parsed file via `editor.ApplyBinding` on `Ignition` with the same Btn 128 + same GUID), then target `TractionControlArm` Btn 128 → 1 conflict.
  - `TestDetectButtonConflictDifferentDeviceIsNotConflict`: same button, different GUID → 0 conflicts.
  - `TestDetectIgnoresTarget`: target `TractionControlArm` already with `'U'` on `f4_key_u.cfg` → 0 conflicts.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/editor/ -run TestDetect -v`
Expected: FAIL.

- [ ] **Step 3: Implement `conflicts.go`**

Scan `f.Records` per the rules; use the target's `Record` to skip it.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/editor/ -run TestDetect -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/editor/conflicts.go internal/editor/conflicts_test.go && git commit -m "feat: binding conflict detection"
```

---

### Task 6: Timestamped backup + diff

**Files:**
- Create: `internal/fileops/backup.go`, `internal/fileops/diff.go`
- Test: `internal/fileops/backup_test.go`, `internal/fileops/diff_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `fileops.Backup(path string) (backupPath string, err error)` — copies to `<path>.bak-YYYYMMDD-HHMMSS` in the same directory; byte-exact.
  - `fileops.Diff(orig, new []byte) []Hunk`; `type Hunk struct { Offset int; Orig, New []byte }` — simple windowed diff hunks, `[]` empty if identical.
  - `fileops.DiffSummary(h []Hunk) string` — readable text like `"3 bytes changed at 0xbd0"`.

- [ ] **Step 1: Write failing tests**
  - `TestBackupCreatesTimestampedCopy`: temp file with known content → backup exists, identical content, name matching `\.bak-\d{8}-\d{6}$`.
  - `TestBackupOverwritesNothing`: two consecutive backups → two distinct files.
  - `TestDiffFindsSingleByteChange`: orig vs new differing in 1 byte at `0xbd0` → 1 hunk with `Offset==0xbd0`, `Orig==[0x55]`, `New==[0x49]`.
  - `TestDiffIdentical`: → empty and `DiffSummary == "no changes"`.
  - `TestDiffSummaryFormat`: 1 byte at `0xbd0` → `"1 byte changed at 0xbd0"`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/fileops/ -v`
Expected: FAIL.

- [ ] **Step 3: Implement `backup.go` and `diff.go`**

Backup: `os.ReadFile` + `os.WriteFile` with `os.O_EXCL` (retry with the next timestamp on collision). Diff: compare in 16-byte windows to produce contiguous hunks.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/fileops/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/fileops && git commit -m "feat: timestamped backups and byte diff"
```

---

### Task 7: Simulator-closed check

**Files:**
- Create: `internal/simcheck/simcheck.go`, `internal/simcheck/simcheck_windows.go`, `internal/simcheck/simcheck_other.go`
- Test: `internal/simcheck/simcheck_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `simcheck.Running() (bool, error)`; `var ErrSimRunning = errors.New("iRacing is running; close it before writing")`.
  - `simcheck.IsSimProcessLine(line string) bool` — exported for pure testing: true for a `tasklist` line containing `iRacingSim64DX11.exe`.

- [ ] **Step 1: Write failing tests** — pure function only (cross-platform):
  - `TestIsSimProcessLineMatches`: real `tasklist` line (`"iRacingSim64DX11.exe  1234 Console  1  1,234 K"`) → true.
  - `TestIsSimProcessLineNoMatch`: `"iRacingUI.exe ..."` and `"notepad.exe"` → false.
  - `TestIsSimProcessLineCaseInsensitive`: `"iracingsim64dx11.exe"` → true.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/simcheck/ -v`
Expected: FAIL.

- [ ] **Step 3: Implement**

`simcheck.go`: `Running()` delegates to a platform helper; `simcheck_windows.go` (`//go:build windows`) runs `tasklist /FI "IMAGENAME eq iRacingSim64DX11.exe" /NH` and applies `IsSimProcessLine`; `simcheck_other.go` returns `false, nil`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/simcheck/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/simcheck && git commit -m "feat: iRacing process check (windows + stub)"
```

---

### Task 8: HTTP server, API and operation orchestration

**Files:**
- Create: `internal/server/server.go`, `internal/server/api.go`, `internal/ops/apply.go`, `internal/inputcap/inputcap.go` (only the `Capturer` interface + types; implementation is Task 9), `web/embed.go` (`package web; //go:embed all:dist` with a temporary `web/dist/index.html` placeholder)
- Modify: `cmd/ibinds/main.go` (wire the server: start listening, print the URL)
- Test: `internal/server/api_test.go`, `internal/ops/apply_test.go`

**Interfaces:**
- Consumes: `config.Config`, `profiles.Scan`, `cfgparse.Parse`, `calib.Parse`, `editor.ApplyBinding`/`FindConflicts`, `fileops.Backup`/`Diff`, `simcheck.Running` (Tasks 1–7).
- Produces:
  - `server.New(cfg *config.Config, opts ServerOpts) *http.Server`; `type ServerOpts struct { Token string; SimCheck func() (bool, error); InputCapturer inputcap.Capturer }` — all injectable for tests; `Token` overrides the generated one (used by the `--token` flag).
  - `inputcap.Capturer` interface declared here (Task 9 provides the real implementation): `Capture(opts CaptureOpts) (*CaptureResult, error)`; `type CaptureOpts struct { Timeout time.Duration }`; `type CaptureResult struct { Button int; DeviceInstanceGUID string; ProductGUID string }`; `var inputcap.ErrUnsupported = errors.New("input capture only supported on Windows")`; the Task 8 stub implementation returns `ErrUnsupported`.
  - The server holds the config behind an `atomic.Pointer[config.Config]`: every route reads the current config **per request**, so `POST /api/settings` takes effect without a restart.
  - Routes (all `/api/` routes require a valid `?token=`; otherwise 401):
    - `GET /` → serves `web.Assets` with SPA fallback to `index.html`; 404 only for missing assets.
    - `GET /api/health` → `{"ok":true}`
    - `GET /api/profiles` → `[{id, displayName, isBaseline, hasControls, hasCalib}]`
    - `GET /api/profiles/{id}/controls` → `[{name, displayName, kind, key, button, deviceName, axisName}]` (`kind` in `unbound|key|button|axis|unknown`; `deviceName` resolved by cross-referencing the `InstanceGUID` with the profile's `joyCalib.yaml` devices)
    - `POST /api/settings` `{profilesDir}` → persists via `config.Save`, swaps the atomic config pointer, returns the new config.
    - `POST /api/plan/binding` `{control, binding:{key|button+deviceInstanceGUID+productGUID}, profileIds, force, profileOverrides:[{profileId, force}]}` → `{plans:[{profileId, conflicts:[{control,reason}], warning, diffSummary, force}]}` — `force` per profile = global `force` OR matching `profileOverrides` entry (batch AND individual confirmation). `warning` present when the device GUID is not listed in the profile's `joyCalib.yaml`; never modifies the YAML.
    - `POST /api/apply/binding` (same body, including `profileOverrides`) → 409 if `simcheck.Running()`; otherwise, per profile: backup, `ApplyBinding`, diff; a profile with conflicts is applied only when its resolved `force` is true, else `skipped`; response `{applied:[{profileId, backupPath, diffSummary}], skipped:[{profileId, reason}]}`
    - `POST /api/plan/calib` `{sourceProfileId, profileIds}` and `POST /api/apply/calib` — whole-file copy of `joyCalib.yaml` (with `calib.Validate` on the target before and after the copy), same 409/backup/diff pattern.
    - `POST /api/input/capture` `{timeoutMs}` → 200 `{button, deviceInstanceGUID, productGUID}` or 204 on timeout; uses the injected `InputCapturer` (real implementation in Task 9; until then returns 501).
  - Pure plan logic lives in `internal/ops/apply.go`: `ops.PlanBinding(profDir string, profileID string, control string, binding editor.Binding, force bool) (PlanResult, error)`; `type PlanResult struct { Conflicts []editor.Conflict; Warning string; Diff string; Force bool }`.

- [ ] **Step 1: Write failing tests** — `internal/server/api_test.go` with `httptest` + a temp profiles dir built from the Task 2 `controls.cfg` fixtures and the Task 3 `joyCalib.yaml` fixtures:
  - `TestHealthRequiresToken`: no token → 401; with token → 200.
  - `TestServesIndex`: `GET /?token=t` → 200 with `index.html` content.
  - `TestProfilesList`: returns baseline first with a friendly `displayName` via `labels`.
  - `TestControlsListResolvesDeviceName`: `f4_btn128.cfg` profile → `TractionControlArm` entry has `kind=="button"`, `button==128`, `deviceName` matching the `joyCalib.yaml` name for that GUID.
  - `TestSettingsPersistsProfilesDir`: `POST /api/settings` → `config.Load` (fresh process of `Load` with the same env) returns the new dir.
  - `TestPlanBindingReportsConflict`: fixture with a key conflict → `plans[0].conflicts` non-empty; `force:false` on apply → `skipped` with reason; `force:true` → `applied`.
  - `TestApplyPerProfileForce`: two profiles, both with conflicts; global `force:false` + `profileOverrides:[{profileId: A, force:true}]` → A applied, B skipped.
  - `TestApplyBlockedWhenSimRunning`: mocked `SimCheck` returning true → 409 and no file in the temp dir changed (compare bytes before/after).
  - `TestApplyCreatesBackup`: successful apply → timestamped backup exists in the profile dir.
  - `TestPlanWarnsDeviceMissing`: button binding with a GUID not in the profile's `joyCalib.yaml` → non-empty `warning` and `joyCalib.yaml` untouched after apply.
  - `TestSettingsTakesEffectWithoutRestart`: `POST /api/settings` with a new profiles dir, then `GET /api/profiles` → profiles from the new dir, no server restart.
  - `TestCalibCopyFlow`: calibration plan/apply copies the source YAML to the target byte-for-byte and creates a backup.
  - `internal/ops/apply_test.go`: `TestPlanBindingDiffShowsOnlyTargetBytes` — `PlanResult.Diff` summarizes exactly the expected edited bytes.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/server/ ./internal/ops/ -v`
Expected: FAIL.

- [ ] **Step 3: Implement server, API and ops**

Token generated in `server.New` (16 hex bytes, overridable via `ServerOpts.Token` from the `--token` flag), `checkToken` middleware; routes per Interfaces; the config is served from an `atomic.Pointer[config.Config]` swapped by `POST /api/settings`. `apply` uses the injected `simcheck`, `fileops.Backup` + `editor.ApplyBinding` + `fileops.Diff`; missing-device validation reads the `joyCalib.yaml` with `calib.Parse` and compares `InstanceGUID`; `deviceName` resolution via `calib.Parse` on the profile's YAML. `web/embed.go` + `web/dist/index.html` placeholder with `go:embed`.

- [ ] **Step 4: Wire `cmd/ibinds/main.go` to start the server**

`main.go` builds `config.Load()`, calls `server.New`, listens on `127.0.0.1:<port>` (`--port` flag, default `0` = ephemeral), prints `http://127.0.0.1:<port>/?token=<t>` to stdout. Flags now: `--dir`, `--list-profiles`, `--token`, `--port` (browser launch comes in Task 11). `go run ./cmd/ibinds --token dev --port 8791` must serve the placeholder page.

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/server/ ./internal/ops/ -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/server internal/ops cmd/ibinds web/embed.go web/dist && git commit -m "feat: local HTTP API with plan/apply for bindings and calibration copy"
```

---

### Task 9: Input capture (DirectInput on Windows)

**Files:**
- Create: `internal/inputcap/inputcap_windows.go` (go-ole + DirectInput), `internal/inputcap/inputcap_other.go`
- Modify: `internal/inputcap/inputcap.go` (add `DeviceInfo`, `Enumerate`, `New`; keep the Task 8 interface), `cmd/ibinds/main.go` (add `--capture` debug flag)
- Test: `internal/inputcap/inputcap_test.go` (pure logic), manual validation on Windows

**Interfaces:**
- Consumes: the `inputcap.Capturer` interface and types declared in Task 8 (keep them as-is).
- Produces:
  - `inputcap.New() Capturer` — platform implementation of the Task 8 interface (the var `ErrUnsupported` already exists from Task 8; reuse it).
  - `type DeviceInfo struct { Name string; InstanceGUID string; ProductGUID string; ButtonCount int }` — add to `inputcap.go`.
  - `inputcap.Enumerate() ([]DeviceInfo, error)` — connected DirectInput devices with button counts and product GUIDs (used by the dropdown).

**Risk note (validated in the plan, not assumed away):** go-ole has no DirectInput bindings; the implementer must hand-write the `IDirectInput8`/`IDirectInputDevice8` vtable definitions and the `DIEnumDevicesCallback` COM object. This is the most uncertain task in the plan. The `ProductGUID` for a device is derived from DIDEVCAPS `dwProductId` as `{0000<pid:%04X>-0000-0000-0000-504944564944}` (document this rule in `docs/FORMAT.md`). If the DirectInput path proves impractical in this task, the task still completes with `Enumerate` returning the devices listed in the profiles' `joyCalib.yaml` files (offline fallback) — capture mode degrades to the dropdown, and the author is informed. Do not silently drop the feature; record the decision in `docs/FORMAT.md`.

- [ ] **Step 1: Write failing tests** (pure logic, runs on Linux)
  - `TestCaptureOptsDefaults`: zero timeout → default 10s.
  - `TestUnsupportedOnNonWindows`: `New().Capture` on the stub → `ErrUnsupported`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/inputcap/ -v`
Expected: FAIL.

- [ ] **Step 3: Implement**

`inputcap.go` keeps the Task 8 interface, adds `DeviceInfo` and `Enumerate`, and gains build-tag dispatch for `New()`. `inputcap_windows.go`: enumerate DirectInput devices via go-ole (`IDirectInput8`, hand-written vtables), poll `IDirectInputDevice8.GetDeviceState` in a loop until timeout, convert the pressed `rgbButtons` entry to an index; convert GUIDs to `{...}` strings; button count from device capabilities; `ProductGUID` from `dwProductId` per the rule above. `inputcap_other.go`: `New()` returns a stub with `ErrUnsupported`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/inputcap/ -v`
Expected: PASS (on Linux; Windows tests run during final validation).

- [ ] **Step 5: Wire `/api/input/capture` and add `GET /api/input/devices`**

Replace the 501 stub: handler uses the injected `InputCapturer`; add `GET /api/input/devices` → `[{name, instanceGUID, productGUID, buttonCount}]` via `Enumerate`. Add the `--capture` flag to `cmd/ibinds/main.go` now (a CLI capture test: press a button, print the result, exit). `--token` and `--port` already exist from Task 8. Integration test with `httptest` using an injected fake capturer.

- [ ] **Step 6: Run server tests**

Run: `go test ./internal/server/ ./internal/inputcap/ -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/inputcap internal/server cmd/ibinds && git commit -m "feat: input capture (DirectInput on Windows) with device enumeration"
```

**Manual validation (author's Windows, record the result in `docs/FORMAT.md`):** run `go run ./cmd/ibinds --capture` on Windows with the MOZA wheel connected; press PL (Btn 32) and Btn 130 and check the indices against iRacing ("Dev 2 Btn 32" etc.). If they diverge, record the observed conversion rule and adjust the implementation before Task 11.

---

### Task 10: Svelte frontend (SPA) + API integration

**Files:**
- Create: `web/package.json`, `web/vite.config.ts`, `web/svelte.config.js`, `web/tsconfig.json`, `web/src/main.ts`, `web/src/App.svelte`, `web/src/lib/api.ts`, `web/src/lib/store.ts`, `web/src/lib/conflicts.ts`, `web/src/components/ProfileList.svelte`, `web/src/components/BindingWizard.svelte`, `web/src/components/CalibCopy.svelte`, `web/src/components/PreviewModal.svelte`, `web/src/components/Settings.svelte`
- Modify: `web/.gitignore` (ensure `dist` is NOT ignored — add `!dist`), `internal/labels/labels.go` (populate the maps)
- Test: `web/src/lib/api.test.ts`, `web/src/lib/conflicts.test.ts`, `web/src/lib/store.test.ts` (Vitest)

**Interfaces:**
- Consumes: API routes (Task 8); friendly names come from `displayName` fields in responses (the frontend needs no local map).
- Produces:
  - `api.fetchProfiles(): Promise<ProfileDTO[]>`, `api.fetchControls(profileId): Promise<ControlDTO[]>`, `api.planBinding(req: BindingReq): Promise<PlanResp>`, `api.applyBinding(req: BindingReq): Promise<ApplyResp>`, `api.planCalib/applyCalib`, `api.fetchInputDevices()`, `api.capture(opts): Promise<CaptureResp>`, `api.saveSettings(profilesDir)` — all appending `?token=`.
  - `token.init(): void` — on app startup, reads `?token=` from `location.search` and stores it in `sessionStorage.token`; if `import.meta.env.DEV` and no URL token, fall back to `"dev"`. `token.get(): string | null` reads `sessionStorage.token`. Production: the token arrives in the URL because the backend opens the browser with it.
  - `keyToByte(key: string): number | null` — pure function mapping a `keydown` `event.key` to the ASCII byte iRacing stores: single printable A–Z/0–9 → uppercase ASCII (`'U'` → `0x55`, `'u'` → `0x55`); everything else (F-keys, arrows, modifiers) → `null` (v1 limitation: only printable keys are capturable; the UI shows a message). Tested.
  - `conflicts.highlight(plan): Record<profileId, ConflictDTO[]>` — pure testable function.
  - `currentBinding(control: ControlDTO): string` — human text: `"Key U"`, `"Btn 128 (MOZA R9 Base)"`, `"none"`, `"axis"` — pure testable function.
  - Svelte `store` with state: `profiles`, `selected`, `control`, `binding`, `plan`, `phase`, `forceOverrides: Record<profileId, boolean>` (per-profile force).

- [ ] **Step 1: Scaffold Svelte + Vite + Vitest and fix gitignore**

`web/` is not empty (it holds `embed.go` and `dist/` from Task 8), so scaffold in a temp dir and merge: `npm create vite@latest /tmp/ibinds-web -- --template svelte-ts`, then move the generated files (excluding its `.gitignore`) into `web/`, keeping the existing `embed.go` and `dist/`. Then `npm i`, `npm i -D vitest` in `web/`. Edit `web/.gitignore` to add `!dist` (so the built assets commit). Configure Vite dev proxy in `vite.config.ts`: `server: { proxy: { '/api': 'http://127.0.0.1:8791' } }` and document the dev flow: `go run ./cmd/ibinds --token dev --port 8791` + `npm run dev` (token `dev` matches the `import.meta.env.DEV` fallback).
Expected: `npm run dev` opens a placeholder; `npx vitest run` runs with no tests.

- [ ] **Step 2: Write failing tests** (Vitest, pure logic):
  - `api.test.ts`: `buildTokenUrl` appends `?token=` from `token.get()`; parses `ApplyResp` with `applied`/`skipped`.
  - `token.test.ts` (or in `api.test.ts`): with a URL token → stored in sessionStorage; DEV without URL token → `"dev"`; production without URL token → `null`.
  - `keyToByte` tests: `'U'`/`'u'` → `0x55`; `'0'` → `0x30`; `'F1'` → `null`; `'ArrowLeft'` → `null`.
  - `conflicts.test.ts`: `highlight` groups conflicts by profileId; a conflict-free plan → `{}`.
  - `store.test.ts`: select all/deselect one; the global `force` flag and `forceOverrides` change `BindingReq.profileOverrides`.
  - `currentBinding` tests in `store.test.ts`: key/button/unbound/axis inputs → expected strings.
  - `internal/labels/labels_test.go`: `TestKnownMappings` — the entries populated in Step 4 return their friendly names.

- [ ] **Step 3: Run tests to verify they fail**

Run: `npx vitest run` and `go test ./internal/labels/ -v`
Expected: FAIL.

- [ ] **Step 4: Implement libs and components**

`ProfileList` with a checkbox per car + "select all" + a `Baseline` badge, and — once a function is chosen — a per-profile line showing the **current binding** via `currentBinding()` (the anti-wrong-profile defense); `BindingWizard`: function picker (with `displayName`), input via capture (keyboard in the browser via `keydown` + `keyToByte`, rejecting non-printable keys with a message; button via `/api/input/capture`), alternative dropdowns — a device dropdown from `/api/input/devices` (using `instanceGUID`/`productGUID`/`buttonCount` from the response) with a button number dropdown (0..buttonCount−1), and a key dropdown (printable A–Z, 0–9 — the set `keyToByte` supports); axis/unknown-kind controls shown but disabled for editing (backend also refuses); `PreviewModal`: per-profile diff, highlighted conflicts, a global "overwrite conflicts" checkbox AND a per-conflicted-profile force checkbox (`forceOverrides` → `profileOverrides` in the request), apply/cancel buttons; `CalibCopy`: source + targets + preview; `Settings`: profiles dir override persisted via `POST /api/settings`. `main.ts` calls `token.init()` before the app mounts. Populate `internal/labels` with entries confirmed in this session: control `LowFuelAccept` → `"Low Fuel Accept"` plus the author's car ids (`f4` → `"F4"` etc.; the implementer asks the author for the car-name list at this point if not already provided).

- [ ] **Step 5: Run tests to verify they pass**

Run: `npx vitest run` and `go test ./...`
Expected: PASS.

- [ ] **Step 6: Frontend build and integration**

Run: `npm run build` (outputs `web/dist`); `go build ./...`; `go test ./internal/server/ -v`.
Expected: all pass; the embed serves the real assets.

- [ ] **Step 7: Commit** (ensure `web/dist` is included)

```bash
git add web internal/labels && git commit -m "feat: Svelte SPA with binding wizard, conflict highlight, preview and calibration copy"
```

---

### Task 11: Packaging, README and final round-trip validation

**Files:**
- Create: `README.md`, `build.ps1`
- Modify: `cmd/ibinds/main.go` (open browser with token; `--no-browser` flag)

**Interfaces:**
- Consumes: everything prior.
- Produces: `ibinds.exe` via `build.ps1`; README with usage, build and contribution instructions.

- [ ] **Step 1: Implement automatic browser launch and the `--no-browser` flag in `main.go`**

`--port`, `--token` and `--capture` already exist (Tasks 8–9). On startup: print `http://127.0.0.1:<port>/?token=<t>` and, unless `--no-browser`, open it with `rundll32 url.dll,FileProtocolHandler` (Windows) / `xdg-open` (others).

- [ ] **Step 2: Write `build.ps1` and `README.md`**

`build.ps1` (PowerShell): `npm run build` in `web/` then `$env:GOOS="windows"; $env:GOARCH="amd64"; go build -ldflags "-s -w" -o ibinds.exe ./cmd/ibinds`. README: what it is, usage, requirements, development (`go run ./cmd/ibinds --token dev --port 8791` + `npm run dev` in `web/`), contribution, MIT license.

- [ ] **Step 3: Manual round-trip validation (Windows; record the result in README or `docs/FORMAT.md`)**

With the simulator closed: (a) use the tool to apply a new key binding in one profile; (b) open iRacing and check the binding in the UI; (c) save inside the sim and exit; (d) reopen the tool and check the binding persists and the file remains valid. Record the result. Do not finish the task without this record.

- [ ] **Step 4: Commit**

```bash
git add README.md build.ps1 cmd/ibinds && git commit -m "docs: README, build script, browser launch, round-trip validation record"
```
