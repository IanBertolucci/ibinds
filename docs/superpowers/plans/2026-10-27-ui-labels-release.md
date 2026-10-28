# iBinds UI, Labels, and Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make iBinds visually modern and practical, replace technical control names with the in-game names from `/home/ian/binds_do_iracing.txt`, and ship a trustworthy GitHub Releases process with a prebuilt Windows binary plus source.

**Architecture:** Keep the existing Go binary + embedded Svelte SPA. The backend only changes for friendly labels and version metadata. The frontend is restyled component by component on top of shared design tokens. Releases are built by GitHub Actions from tags and published with checksums and build provenance.

**Tech Stack:** Go 1.22+, Svelte 5, Vite 8, TypeScript, Vitest, GitHub Actions.

**Spec:** This plan is the approved design from the 2026-10-27 session.

## Global Constraints

- Do not edit `joyCalib.yaml`; calibration is copied whole, never modified.
- Every write stays surgical, backed up, and blocked while iRacing is running.
- All code, comments, docs, and commit messages are in English.
- No commits unless the user explicitly asks.
- Labels are display-only; never change parser/editor identifiers or file bytes.
- Any control absent from `/home/ian/binds_do_iracing.txt` keeps its technical ID.
- Grouped setup controls use the action suffix: `↑` for `Inc`, `↓` for `Dec`; `Level` uses the base name.
- The visual direction is light, compact, functional, and distinct from the discontinued dark three-column editor.
- Do not add a paid code-signing dependency in this phase.
- Existing API behavior for binding/copy/calibration must be preserved.

## Review Focus

- A friendly label must never be applied to the wrong internal control; the fixture-backed existence test and the ambiguous mapping table are the guard.
- A mapped label must remain searchable by both its friendly name and technical ID.
- The profile/sidebar/settings layout must not overlap or cause horizontal scrolling at 1280-1920 px.
- Copy/edit/calibration flows must keep working after the DOM restyle, including multi-button bindings and conflict overrides.
- Release artifacts must be built from the same tag/source and must include the checksum and source zip.

---

## Task 1: Control labels mapping

**Files:**
- Create: `internal/labels/control_labels.go`
- Modify: `internal/labels/labels.go`
- Test: `internal/labels/labels_test.go`

**Interfaces:**
- Produces: `Control(id string) string` keeps its signature and returns the curated label or the id.
- Produces: `controlLabels map[string]string` in `internal/labels/control_labels.go`.

- [ ] **Step 1: Write failing tests** in `internal/labels/labels_test.go`
  - `TestControlLabelsExistInFixtures`: parse every `../cfgparse/testdata/*.cfg` file, collect record names, and assert every key in `controlLabels` is present in the union.
  - `TestControlGroupedLabels`: assert `Control("BrakeBiasInc") == "Brake Bias Set ↑"`, `Control("BrakeBiasDec") == "Brake Bias Set ↓"`, `Control("BrakeBiasLevel") == "Brake Bias Set"`.
  - `TestControlQuickChat`: assert `AutoChat1 == "Pitting In"`, `AutoChat5 == "Sorry!"`, `AutoChat7 == "AutoChat7"`.
  - `TestControlFallback`: update the existing test to use `NotARealControl` only; `TractionControlArm` is now mapped.
  - `TestControlLabelsUnique`: assert no two mapped IDs produce the same label.
- [ ] **Step 2: Run the tests and confirm failure**
  - Run: `go test ./internal/labels -count=1`
  - Expected: FAIL because the new mappings are absent.
- [ ] **Step 3: Create `control_labels.go`**
  - Move `controlLabels` out of `labels.go`.
  - Direct matches from `binds_do_iracing.txt` (e.g. `PitSpeedLimiter`, `TearOffVisor`, `PushToPass`, `ToggleWindshieldWipers`).
  - Group rules: `*Inc` → `Base ↑`, `*Dec` → `Base ↓`, `*Level` → `Base`; apply to brake bias, ABS, traction control 1-4, differential, anti-roll, wing, qualifying tape, engine, boost, launch RPM, fuel mixture/cut, power steering, dash pages, hybrid MGU-K, shocks, and weight jackers.
  - Setup toggles: `ABSToggle`, `TractionControlToggle`, `FuelNoCutToggle`, `RFBrakeAttachedToggle`, `InLapToggle`, `FCYToggle`, `HysBoostHold`, `HysRegenHold`, `HysNoBoostToggle`.
  - Cockpit/driver: `Throttle2` → `Second Throttle`, `Brake2` → `Second Brake`, `Clutch2` → `Second Clutch`, `ShiftUp2`/`ShiftDown2` → `Second Upshift`/`Second Downshift`, `Reset` → `Enter / Exit / Tow Car`.
  - Interface/black box/spotter/audio/chat/camera/replay mappings from the txt, including `BlackBoxShowNext` vs `BlackBoxNext`.
  - Quick chat: `AutoChat1..6`; `AutoChat7..15` stay unchanged (blank names in the txt).
  - Proposed uncertain defaults: `DCamVanishYInc/Dec` → `Shift Horizon Up/Down`; `ToggleSystemMeterDisplay` → `FPS / Network Display`; `ToggleUIVisible` → `Cycle UI Visibility Mode`; `ToggleirsdkDiskLogging` → `Toggle Telemetry Logging`; `IrsdkDriverMarker` → `Mark Event in Telemetry`; `LFEVolumeUp/Down` → `Wheel LFE Louder/Quieter`; `TrueForceVolumeUp/Down` → `BassShaker LFE Louder/Quieter`; `WeightJackerLeft/Right*` → `Left/Right Spring Set`; `RpyIncreaseRate` → `Slow Motion`.
- [ ] **Step 4: Run the tests and confirm pass**
  - Run: `go test ./internal/labels -count=1`
  - Expected: PASS.
- [ ] **Step 5: Run the backend suite**
  - Run: `go test ./... -count=1`
  - Expected: PASS.

**Acceptance:** The backend returns friendly labels; unmapped IDs are unchanged; no file-format code changed.

---

## Task 2: UI foundation and shell

**Files:**
- Create: `web/src/app.css`
- Create: `web/src/components/Icon.svelte`
- Modify: `web/src/main.ts`
- Modify: `web/index.html`
- Modify: `web/src/App.svelte`
- Modify: `web/src/components/Sidebar.svelte`
- Modify: `web/src/components/Settings.svelte`
- Test: `web/src/App.test.ts`, `web/src/components/Sidebar.test.ts`
- Create: `web/src/components/Settings.test.ts`

**Interfaces:**
- Consumes: existing `ProfileDTO`, `ControlDTO`, `api` modules.
- Produces: design tokens in `app.css` (`--bg`, `--surface`, `--text`, `--muted`, `--border`, `--accent`, `--accent-soft`, `--danger`, `--warning`, `--success`, `--radius`, `--shadow`).
- Produces: `Icon` component with `name` prop (`search`, `settings`, `edit`, `copy`, `close`, `check`, `chevron`, `device`, `refresh`).

- [ ] **Step 1: Update shell tests first**
  - `App.test.ts`: assert the top bar shows `iBinds`, the profile rail lists Baseline/F4/F5, selecting F4 shows its controls, and the settings gear opens a dialog containing the profiles-directory field.
  - `Sidebar.test.ts`: assert profile search filters the list and the active profile is marked.
- [ ] **Step 2: Run tests and confirm failure**
  - Run: `cd web && npx vitest run src/App.test.ts src/components/Sidebar.test.ts`
  - Expected: FAIL on the new dialog/search expectations.
- [ ] **Step 3: Add tokens and icons**
  - `app.css`: system font stack, neutral light palette, teal accent, 8px spacing scale, 10px radius, subtle border/shadow, focus-visible ring, reduced-motion media query, base button/input/dialog styles.
  - Import `./app.css` from `main.ts`.
  - `Icon.svelte`: inline SVGs, `aria-hidden="true"`, size via `currentColor`.
- [ ] **Step 4: Rebuild the shell**
  - `App.svelte`: top bar, profile rail, main content, settings dialog state; keep `activeProfile = $derived(...)` and existing API loading.
  - `Sidebar.svelte`: search input, baseline badge, active state, empty state.
  - `Settings.svelte`: render inside a native `<dialog>` opened by the gear; keep save behavior and the profiles-directory field.
- [ ] **Step 5: Run tests and confirm pass**
  - Run: `cd web && npx vitest run src/App.test.ts src/components/Sidebar.test.ts src/components/Settings.test.ts`
  - Expected: PASS.
- [ ] **Step 6: Build check**
  - Run: `cd web && npm run build`
  - Expected: PASS with no Svelte warnings.

---

## Task 3: Controls table and profile page

**Files:**
- Modify: `web/src/components/ControlsTable.svelte`
- Modify: `web/src/components/ProfilePage.svelte`
- Test: `web/src/components/ControlsTable.test.ts`, `web/src/components/ProfilePage.test.ts`

**Interfaces:**
- Consumes: `ControlDTO`, `selected`, `search`, callbacks from `ProfilePage`.
- Produces: selection bar with `selectedCount` and copy callback; filter chips `All`, `Bound`, `Unbound`, `Copyable`.

- [ ] **Step 1: Write failing tests**
  - Search by friendly name and technical ID both filter rows.
  - Filter chips narrow the table.
  - Selecting rows shows the bottom selection bar and the selected count.
  - Multi-button binding renders `7 + 33` as chips.
  - Unbound rows show `None`.
- [ ] **Step 2: Run tests and confirm failure**
  - Run: `cd web && npx vitest run src/components/ControlsTable.test.ts src/components/ProfilePage.test.ts`
  - Expected: FAIL on the new UI expectations.
- [ ] **Step 3: Implement the table**
  - Toolbar: search input, filter chips, result count.
  - Sticky header, compact rows, hover/focus states, selected-row tint.
  - Binding cell: monospaced value, device name muted, multi-button chips.
  - Row action: Edit button.
  - Bottom selection bar with count and `Copy to profiles →`.
  - Empty states for no controls and no matches.
- [ ] **Step 4: Implement profile header**
  - Car name, control/bound counts, segmented Controls/Calibration tabs, selection reset when switching profile.
- [ ] **Step 5: Run tests and confirm pass**
  - Run: `cd web && npx vitest run src/components/ControlsTable.test.ts src/components/ProfilePage.test.ts`
  - Expected: PASS.
- [ ] **Step 6: Build check**
  - Run: `cd web && npm run build`
  - Expected: PASS with no warnings.

---

## Task 4: Edit, copy, and calibration flows

**Files:**
- Modify: `web/src/components/BindingModal.svelte`
- Modify: `web/src/components/BindingWizard.svelte`
- Modify: `web/src/components/ProfileList.svelte`
- Modify: `web/src/components/CopyModal.svelte`
- Modify: `web/src/components/PreviewModal.svelte`
- Modify: `web/src/components/CalibCopy.svelte`
- Test: `web/src/components/BindingModal.test.ts`, `web/src/components/CopyModal.test.ts`
- Create: `web/src/components/CalibCopy.test.ts`

**Interfaces:**
- Consumes: existing `api.planBinding`, `api.applyBinding`, `api.planCopy`, `api.applyCopy`, `api.planCalib`, `api.applyCalib`.
- Produces: no API contract changes.

- [ ] **Step 1: Write failing tests**
  - Edit opens the wizard preset with the selected control.
  - Copy modal previews conflicts and applies with/without force.
  - Calibration preview shows per-target plans before applying.
- [ ] **Step 2: Run tests and confirm failure**
  - Run: `cd web && npx vitest run src/components/BindingModal.test.ts src/components/CopyModal.test.ts src/components/CalibCopy.test.ts`
  - Expected: FAIL on the new flow expectations.
- [ ] **Step 3: Restyle edit flow**
  - Native dialog with header, device cards, capture key/button states, clear primary action, error banner, close/escape.
  - Profile list with per-profile current binding and checkbox.
- [ ] **Step 4: Restyle copy flow**
  - Target cards, preview plans, conflict/warning rows, per-profile force, global force, sticky footer.
- [ ] **Step 5: Improve calibration flow**
  - Source fixed to the current profile; target checkboxes; select all; preview using `api.planCalib`; apply after preview.
- [ ] **Step 6: Run tests and confirm pass**
  - Run: `cd web && npx vitest run src/components/BindingModal.test.ts src/components/CopyModal.test.ts src/components/CalibCopy.test.ts`
  - Expected: PASS.

---

## Task 5: Visual polish and full frontend verification

**Files:**
- Modify: all touched Svelte components for responsive/a11y polish.
- Test: full `web/src/**/*.test.ts`.

- [ ] **Step 1: Responsive and accessibility pass**
  - 1280/1440/1920 checks; sidebar collapses below ~1100 px; no horizontal overflow.
  - Focus-visible on all interactive elements; dialogs have labels and close on Escape; tables use real `<table>` semantics.
  - Reduced motion respected.
- [ ] **Step 2: Run the full frontend suite**
  - Run: `cd web && npx vitest run`
  - Expected: PASS.
- [ ] **Step 3: Build**
  - Run: `cd web && npm run build`
  - Expected: PASS with no warnings.

---

## Task 6: Versioning and CI

**Files:**
- Modify: `cmd/ibinds/main.go`
- Create: `.github/workflows/ci.yml`
- Test: `cmd/ibinds` build + existing Go suite.

**Interfaces:**
- Produces: `var version = "dev"`, `var commit = "none"`, `var buildDate = "unknown"` in `package main`.
- Produces: `--version` flag prints `ibinds <version> (<commit>, <buildDate>)`.

- [ ] **Step 1: Add version metadata**
  - `main.go`: version vars, `--version` flag, log the version at startup.
- [ ] **Step 2: Add CI workflow**
  - Go job: `go test ./...`, `go vet ./...`, `GOOS=windows GOARCH=amd64 go build ./...`.
  - Web job: `npm ci --legacy-peer-deps`, `npx vitest run`, `npm run build`.
  - Trigger on push and pull request.
- [ ] **Step 3: Verify**
  - Run: `go test ./... && go vet ./... && GOOS=windows GOARCH=amd64 go build ./...`
  - Run: `GOOS=windows GOARCH=amd64 go build -o /tmp/ibinds.exe ./cmd/ibinds && /tmp/ibinds.exe --version`
  - Expected: version line printed.

---

## Task 7: Release workflow and documentation

**Files:**
- Create: `.github/workflows/release.yml`
- Modify: `README.md`
- Create: `docs/RELEASING.md`
- Modify: `build.ps1`
- Modify: `build.sh` (optional version support)

- [ ] **Step 1: Add release workflow**
  - Trigger on tags `v*` and `workflow_dispatch`.
  - Build and test web + Go.
  - Build `ibinds.exe` with `-trimpath -ldflags "-s -w -X main.version=<tag> -X main.commit=<sha> -X main.buildDate=<date>"`.
  - Generate `SHA256SUMS.txt`.
  - Attach build provenance with `actions/attest-build-provenance`.
  - Publish `ibinds.exe` and `SHA256SUMS.txt` with GitHub release notes.
  - Required permissions: `contents: write`, `id-token: write`, `attestations: write`.
- [ ] **Step 2: Update build scripts**
  - `build.ps1`: `$ErrorActionPreference = "Stop"`, `npm ci --legacy-peer-deps`, fail on non-zero, optional `-Version` parameter passed to ldflags.
  - `build.sh`: optional `VERSION` env passed to ldflags.
- [ ] **Step 3: Update README**
  - Install from Releases, verify SHA256 in PowerShell, run.
  - Build from source with Go 1.22+ and Node 20+, including `powershell -ExecutionPolicy Bypass -File .\build.ps1`.
  - SmartScreen explanation; never tell users to disable it.
  - Link to `Source code (zip)` and checksums.
- [ ] **Step 4: Add `docs/RELEASING.md`**
  - Semver rules, tag command, workflow verification, manual Windows smoke test, release notes template.
- [ ] **Step 5: Verify locally**
  - Run: `./build.sh`
  - Run: `powershell.exe -NoProfile -ExecutionPolicy Bypass -File build.ps1` (from WSL) or document if unavailable.
  - Expected: `ibinds.exe` built; `--version` works; no warnings.

---

## Task 8: Release verification (manual, requires user approval to publish)

- [ ] **Step 1: Check release assets**
  - Release page contains `ibinds.exe`, `SHA256SUMS.txt`, and GitHub's source zip.
- [ ] **Step 2: Verify on Windows**
  - Download, compare SHA256, run `ibinds.exe --version`, open the UI, test copy/edit with iRacing closed.
- [ ] **Step 3: Verify source build**
  - Download the source zip, install Go/Node, run `build.ps1`, start `ibinds.exe`.
- [ ] **Step 4: Tag/push**
  - Only after explicit user approval; publishing is an external side effect.

---

## Verification Summary

```bash
go test ./... && go vet ./... && GOOS=windows GOARCH=amd64 go build ./...
cd web && npx vitest run && npm run build
./build.sh
```

## Out of Scope

- Paid code signing, MSIX, Microsoft Store, winget/Scoop manifests.
- Auto-update.
- PWA/helper architecture.
- Dark theme.
- Changing binding/copy/calibration behavior or file formats.

---

## Post-implementation amendments (2026-10-27)

- Quick chat now maps `AutoChat8` (`:(`) and `AutoChat12..15` (the chat command
  strings) in addition to 1-6. Slots 7 and 9-11 are blank in the source list and
  keep their technical ids.
- `LFEVolume*` maps to BassShaker LFE and `TrueForceVolume*` to Wheel LFE,
  based on Logitech TrueForce being wheel haptics. This remains the most
  uncertain mapping in the table.
- Dialogs use a `role="dialog"` container plus a reusable focus-trap action
  instead of native `<dialog>`/`showModal()`; this keeps the happy-dom tests
  stable while providing Tab trapping, focus restore, and scoped Escape.
- Select-all in the controls table toggles only visible copyable rows; hidden
  selections are reported in the footer.
- Calibration targets reset to all profiles except the source when the active
  profile changes.
- The release workflow validates that the checked-out ref is a tag and uses the
  checked-out HEAD SHA for `main.commit`.
- Settings save failures render as errors, not success messages.
- No commits were made; all changes remain in the working tree.

## Post-review fix: binding column alignment (2026-10-27)

The binding cell used `display: flex` directly on a `<td>`, which removes it
from table-cell layout and misaligns the Binding column in some browsers. The
cell now keeps its default table-cell display and wraps the flex content in an
inner `.binding-content` div. Covered by the structural test
`binding cell keeps the table cell and wraps flex content inside it`.
