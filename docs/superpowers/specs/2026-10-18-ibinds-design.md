# iBinds — Design

Date: 2026-10-18
Status: awaiting final approval (builder check)

## 1. Intent and why

iRacing keeps control bindings and calibration in a baseline profile and in
per-car profiles. Today, any binding change requires opening the simulator,
loading a test drive per car (~2 min per car) and editing them one by one.
Recalibrating after a hardware reset (e.g., Moza app) requires repeating the
process for every car.

We want a local tool that reads iRacing's configuration files
(`controls.cfg` and `joyCalib.yaml`), applies a binding change across many
profiles with review and confirmation before writing, and copies calibration
between profiles — without opening the simulator, and safely (backup,
preview, simulator closed).

The project will be open source, maintained mostly by the author with the
help of AIs. Lightness matters (avoid the Electron stereotype), but
usability and maintainability with well-known frameworks matter more.

## 2. Validated context (real experiments, Oct 2026)

Everything below was confirmed by controlled experiments on the author's
files (binary diff before/after changes made inside the simulator), not by
assumption:

- Default location: `Documents/iRacing/profiles/controls/`, containing a
  `Baseline` folder and one folder per car with its own profile (e.g. `f4`,
  `audigto`). The path is configurable and the directory is detected
  reliably (resolve the user's `Documents` on Windows; allow manual
  override).
- Each profile is a pair of files: `controls.cfg` (binary) and
  `joyCalib.yaml` (text YAML, CRLF). Profiles are complete copies; there is
  no on-disk inheritance. Cars without a profile use the baseline inside the
  game; custom profiles are independent from creation onward ("Save as..." =
  copy).
- `controls.cfg` — decoded binary format:
  - `CCFG` magic (bytes `GFCC` LE) at the start; a size field at offset
    `0x8` = file size − 12.
  - Control records: readable NUL-terminated name (e.g.
    `TractionControlArm`, `LowFuelAccept`, `Ignition`), followed by fixed
    fields including:
    - binding type (offset varies per record): `02` = device button,
      `04` = keyboard key;
    - for keyboard: ASCII byte of the key (e.g. `0x55` = 'U');
    - for button: bitmask in 32-bit words — button N is bit `N % 32` of
      word `N / 32` (e.g. Btn 130 = word 4, value 4);
    - for device: InstanceGUID (16 bytes) + ProductID (4 bytes) + the
      string `PIDVID`, identical to the identifiers in `joyCalib.yaml`
      (e.g. MOZA R9 Base `{991805D0-2EB6-11F1-800A-444553540000}`).
  - The readable names in the file correspond to the names shown in the
    game (e.g. "Low Fuel Accept" ↔ `LowFuelAccept`).
- Round-trip validated in both directions:
  - A file edited surgically outside the game is loaded without errors and
    the binding shows in the simulator UI.
  - On save, the simulator rewrites the file preserving byte-for-byte
    everything that did not change (when the file is already in the current
    game version's format).
- Silent conflict (critical finding): if the same key/button appears in two
  functions, the simulator resolves it by itself on load, **without
  warning**, removing one of the bindings. The tool MUST detect and warn
  about conflicts before writing — a safety requirement, not a convenience.
- The simulator rewrites `joyCalib.yaml` when saving controls, with
  identical content when nothing changed. `joyCalib.yaml` is plain readable
  YAML.
- Files may only be edited while the simulator is closed (the process
  `iRacingSim64DX11.exe` must not be running).
- Files written by older game versions may have a different layout (new or
  renamed controls). Mitigation: surgical editing + preserving unknown
  structure byte-for-byte.
- The complete record framing (per-type prefix fields, binding counts, the
  fixed size of the payload union, header size recalculation) is not yet
  100% mapped. At least types `02` (button), `04` (key) and axis-like types
  (e.g. `0x19`/`0x1f` on `SteerLeft`/`Throttle` records) have been
  observed. **The plan must start with a milestone to complete the format
  mapping**, with round-trip fixtures (real before/after pairs, re-running
  the experiment if needed), before any write code.
- The v1 parser must read and preserve **all** binding types present in
  real files (including axes and unknown structures, preserved
  byte-for-byte). v1 **editing** is limited to key and button; axis
  bindings are displayed as information, not editable.
- Device missing from the target profile: when the target profile's
  `joyCalib.yaml` does not list the binding's device, the tool **warns**
  and guides the user to create the device's initial calibration inside the
  game for that profile. **Never** insert or edit calibration values
  automatically — the semantics of the values (units, angles, etc.) are
  unknown.
- Known community problem: Windows updates can change device IDs and break
  `joyCalib.yaml` files. The tool may detect GUID mismatches across
  profiles and flag them, but must not auto-fix.

## 3. Goals

- Apply a binding (function + key or button) to N profiles (including
  baseline) with: car list, bulk and individual selection, conflict
  highlighting, overwrite with batch-confirmable warning, diff preview,
  automatic backup and surgical write.
- Copy calibration (`joyCalib.yaml`) from a source profile to
  all/selected. Never edit calibration values.
- Detect key/button conflicts per profile before writing.
- Guarantee safety: simulator closed, timestamped backup, structure
  validation before writing, byte-for-byte preservation of what did not
  change.
- Pleasant, usable UI (visual quality matters).

## 4. Non-goals

- Never edit calibration values (whole-file copy only).
- No telemetry, overlays, or running alongside the simulator.
- No support for other simulators.
- No format conversion between old game versions (detect and warn, at
  most).

## 5. Anti-goals (what makes the project a failure even if it works)

- Corrupting the user's configuration files without a restorable backup.
- Writing bindings to the wrong profile without the user noticing.
- Silently dropping bindings (the same sin as the simulator, which we want
  to avoid).
- Being yet another "slow Electron app" without a reason.

## 6. Constraints

- Runs on Windows (iRacing is Windows-only; files live under
  `C:\Users\...\Documents`).
- The simulator must be closed during any write.
- Writes must be surgical: change only the target record's fields,
  preserving the rest of the file byte-for-byte; validate magic/structure
  before writing.
- No runtime installation required by the end user (single binary).
- Open source (license to be decided during implementation; MIT suggested).
- All code, comments, docs and commit messages in English.

## 7. Stack decisions (made with the author)

- **Backend: Go.** Single binary, local HTTP server on `127.0.0.1` (no
  firewall port; one-time token in the URL to prevent access from other
  tabs/processes), file handling, backup, diff, input capture.
- **Frontend: Svelte** (with Vite), SPA embedded in the binary via
  `go:embed`. Frontend build step accepted in exchange for better
  UI/maintainability.
- **Distribution:** single Windows executable (.exe; exact form — portable
  zip vs installer — to be decided during implementation).
- **Car names in the UI:** friendly-name mapping when possible; a future
  version may show both friendly and technical names.

## 8. Application behavior

### 8.1 General UI
- On launch: start the local server and open the browser automatically.
- Profiles directory configuration: automatic detection of
  `Documents/iRacing/profiles/controls` + persisted manual override.
- Profile list (baseline identified as such + cars) with friendly names.

### 8.2 Binding flow
1. User picks the target function (list of all controls from the file,
   shown with friendly name when mapped) and the input:
   - capture mode (priority): user presses the key or wheel button;
     keyboard keys are captured client-side in the browser (`keydown`);
     wheel buttons are read by the backend via DirectInput/RawInput;
   - dropdown with all known buttons/keys (alternative).
2. User picks the target device when relevant (wheel vs keyboard). A
   function never gets bindings of both kinds at once (author's rule).
3. Car list with bulk ("all") and individual selection, showing per car:
   current binding of the function, conflicts (input already used by
   another function) highlighted.
4. Conflict → overwrite with a warning, confirmable in batch (all
   conflicts at once) and individually.
5. Per-file diff preview before writing.
6. Write: automatic timestamped backup, surgical write, report of what
   changed. Simulator closed is a backend-verified precondition.

### 8.3 Calibration flow
1. Pick source profile.
2. Pick targets (all/selected).
3. Preview (diffs of the `joyCalib.yaml` files), backup, whole-file copy.
4. Values are never edited.

### 8.4 Safety and verification
- Check for the iRacing process before any write (block with a clear
  message).
- Automatic backup (timestamped copy of the original file) before every
  write; backup path visible in the UI.
- Validate `CCFG` magic and basic consistency before writing.
- Preserve byte-for-byte everything outside the target records.

## 9. High-level architecture

```
[ Svelte SPA (embedded via go:embed) ]
             │ HTTP (127.0.0.1 + token)
[ Go backend ]
 ├─ detect/verify: sim process, profiles directory
 ├─ parser: controls.cfg (records, bindings, button bitmask, GUIDs),
 │          joyCalib.yaml
 ├─ writer: surgical editing + validation
 ├─ backup/diff: timestamped snapshots, per-file diff
 ├─ input: key/button capture (DirectInput/RawInput)
 └─ server: single tokenized route, JSON API
```

## 10. Risks and mitigation

| Risk | Mitigation |
|---|---|
| Round-trip not accepted by the game | VALIDATED by experiment (load and save preserve external edits) |
| Binary layout evolves across game versions | Surgical editing; preserve unknown bytes; validate magic before writing |
| Button numbering in capture mode not matching iRacing's | Use DirectInput the same way the simulator does; validate in the prototype with the author's wheel; dropdown as fallback |
| Writing to the wrong profile | Explicit per-car target UI; diff preview before writing |
| Silent conflict | Own conflict detection before writing |
| Data loss | Automatic backup + writes only while sim closed |

## 11. Prototype scope (v1)

0. Milestone: complete binary format mapping (record framing, all types
   present in real files, header recalculation) with round-trip fixtures.
1. Parsing both formats (`controls.cfg`, `joyCalib.yaml`) with validation
   and preservation of unknown structures.
2. Svelte UI: profile list with friendly names, bulk/individual selection.
3. Bulk binding: key capture (keyboard) and button capture (DirectInput) +
   dropdown as alternative.
4. Calibration copy between profiles (whole-file copy, no value editing).
5. Timestamped backup + diff preview + surgical write.
6. Conflict detection with warning and batch-confirmed overwrite.
7. Simulator-closed verification.
8. Missing-device warning on the target profile (guide in-game
   calibration; never auto-edit calibration values).
9. Final round-trip test on the author's simulator.

Out of v1: simultaneous technical+friendly names in the UI, signed
installer, support for other formats.

## 12. Left to the builder

- Exact Go project structure (packages, modules) and Svelte app structure.
- Input capture mechanism (specific DirectInput/RawInput library, e.g.
  go-ole vs alternative) — with the caveat that button numbering must be
  validated against iRacing in the prototype.
- Local route token format and server lifecycle.
- Backup strategy (directory, retention).
- UI details (components, layout) following the usability defined above.
- Open source license (MIT suggested) and .exe packaging.
