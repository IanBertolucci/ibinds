# Profile-centric view + bulk binding copy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a profile-centric UI (left sidebar of profiles → per-profile page showing every control and its current binding) and a bulk "copy selected bindings from this profile to other profiles" operation, while keeping the existing bulk-binding wizard (relocated to a global button) and calibration copy (relocated under the profile page).

**Architecture:** Copy is a byte-exact copy of a source record's binding region (type + key/mask + GUID + ProductID + PIDVID), which makes multi-button bindings copy faithfully (a control's entire button mask is transferred, never just the lowest button). Two new endpoints (`/api/plan/copy`, `/api/apply/copy`) mirror the existing binding plan/apply flow. The `controls` DTO gains a `buttons` array so the table can show every button. The Svelte app is restructured around a sidebar + profile page, with the existing wizard extracted into a modal and a new `CopyModal`.

**Tech Stack:** Go 1.22+, Svelte 5 + Vite + Vitest (unchanged).

**Spec:** `docs/superpowers/specs/2026-10-27-profile-view-copy-bindings-design.md`

## Global Constraints

- Windows target; Windows-only code isolated behind `//go:build windows` with stubs; tests run on Linux.
- Simulator closed before any write (409, no file touched).
- Surgical byte-exact edits; validate `CCFG` magic before writing; timestamped backup before every write.
- Never edit calibration values (`joyCalib.yaml` is only copied whole).
- Copy must never silently drop bindings: a button binding is copied **whole** (full mask), and only key/button source records are copyable.
- Go nil slices marshal to JSON `null` → always `make([]T, 0)` for API-response arrays (e.g. `buttons`, `conflicts`, `applied`, `skipped`).
- Conflict detection is namespace-scoped (replay/camera controls excluded) and follows the existing rules.
- Copy targets exclude the source profile (`profiles - source`).
- Go module `ibinds`; frontend build output in `web/dist` committed.
- Unit tests driven by real fixtures in `internal/cfgparse/testdata/`; never replace them with synthetic guesses.
- All code, comments, docs, commit messages in English.

## Review Focus

1. **Copying a multi-button control** (e.g. `TearOffVisor` = Btn 7 + Btn 33) must copy ALL buttons, not just the lowest. Test in Task 1 (`TestCopyBindingMultiButton`) and Task 2 (`TestApplyCopyMultiButtonReproducesRealFixture`).
2. **Copy where the target already binds the input to a different control** → conflict reported, target skipped unless forced. Test in Task 2 (`TestPlanCopyConflict`, `TestPlanCopyMultiButtonConflict`) and Task 3 (`TestApplyCopySkipsConflict`).
3. **Copy where the target profile lacks the source's device** (`joyCalib.yaml` has no matching InstanceGUID) → warning, never edits calibration. Test in Task 2 (`TestPlanCopyWarnsDeviceMissing`).
4. **Simulator running during copy** → 409, no file touched. Test in Task 3 (`TestApplyCopyBlockedWhenSimRunning`).
5. **Copy produces byte-exact reproduction of a real simulator edit** (key→button and button→key transitions already captured as fixtures). Test in Task 2 (`TestApplyCopyKeyReproducesRealFixture`, `TestApplyCopyButtonReproducesRealFixture`).

---

### Task 1: Byte-exact copy writer + multi-button conflict detection in `internal/editor`

**Files:**
- Create: `internal/editor/copy.go`
- Test: `internal/editor/copy_test.go`

**Interfaces:**
- Consumes: `cfgparse.File`, `cfgparse.Record`, `cfgparse.ValidMagic`, `cfgparse.TypeOffset`/`KeyOffset`/`MaskWordsOffset`/`GUIDOffset`/`ProductOffset`/`PIDVIDOffset`/`FixedRegionSize`, `cfgparse.KindKey`/`KindButton`/`KindAxis`/`KindUnknown`, `cfgparse.ErrBadMagic`, `ErrUnknownControl`, `ErrUnsupportedKind`, `Conflict`, `isReplayOrCamera` (existing).
- Produces:
  - `var ErrNotCopyable = errors.New("editor: source record has no copyable binding")`.
  - `editor.CopyBinding(f *cfgparse.File, control string, src cfgparse.Record) ([]byte, cfgparse.Record, error)`:
    - validate `cfgparse.ValidMagic(f.Raw)` → `ErrBadMagic`;
    - find the target record by name → `ErrUnknownControl` if absent;
    - target kind `KindAxis`/`KindUnknown` → `ErrUnsupportedKind`;
    - `src.Kind` not `KindKey`/`KindButton` → `ErrNotCopyable`;
    - copy `src`'s binding region — the fixed region bytes from offset `TypeOffset` (8) to `FixedRegionSize` (68), i.e. type + key/mask + reserved + GUID + ProductID + reserved + PIDVID — onto the target record's fixed region, leaving flags (+0) and category (+4) untouched;
    - reparse and return the new bytes and the target record.
  - `editor.FindCopyConflicts(f *cfgparse.File, target string, src cfgparse.Record) []Conflict` — same namespace scoping as `FindConflicts` (skip `target`, `isReplayOrCamera(target) == isReplayOrCamera(r.Name)`); key source conflicts with another `KindKey` with the same `Key`; button source conflicts when any of `src.Buttons` overlaps another `KindButton` record's `Buttons` on the same `Device.InstanceGUID`. Returns `make([]Conflict, 0)`.

- [ ] **Step 1: Write failing tests** — `internal/editor/copy_test.go` (package `editor`, uses `../cfgparse/testdata` fixtures):

  - `TestCopyBindingKeyReproducesRealFixture`: parse `f4_btn128.cfg`; take `TractionControlArm`'s record from `f4_key_u.cfg` as `src`; `CopyBinding(f, "TractionControlArm", srcKeyU)` → output byte-identical to `f4_key_u.cfg`.
  - `TestCopyBindingButtonReproducesRealFixture`: parse `f4_key_u.cfg`; `src` = `TractionControlArm` from `f4_btn130.cfg`; `CopyBinding` → byte-identical to `f4_btn130.cfg`.
  - `TestCopyBindingMultiButton`: parse `f4_key_u.cfg`; `src` = `TearOffVisor` (buttons `[7,33]`); apply to a fresh `f4_btn128.cfg` whose `TearOffVisor` is the same `[7,33]` (verify both buttons preserved) — assert the target record's reparsed `Buttons == [7,33]` and the mask words are byte-identical to the source's.
  - `TestCopyBindingPreservesFlagsAndCategory`: copy `TractionControlArm` and assert bytes `+0..+7` (flags/category) of the target record are unchanged.
  - `TestCopyBindingUnknownControl`: nonexistent control → `ErrUnknownControl`.
  - `TestCopyBindingRefusesAxisTarget`: `src` = a key record; target `SteerLeft` (axis) → `ErrUnsupportedKind`.
  - `TestCopyBindingRefusesNonCopyableSource`: `src` = an unbound or axis record → `ErrNotCopyable`.
  - `TestFindCopyConflictsMultiButton`: build a file where `TractionControlArm` is set to `Btn 7 + 33` (via `ApplyBinding` twice is not possible — instead copy a `TearOffVisor`-style record, or construct via `CopyBinding` from a record with `[7,33]`); then `FindCopyConflicts(f, "PushToPass", srcTearOff)` where `srcTearOff.Buttons == [7,33]` → 1 conflict with `Control == "TractionControlArm"`.
  - `TestFindCopyConflictsKey`: mirror `TestDetectKeyConflict` but via `FindCopyConflicts` with a key source record.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/editor/ -run 'Copy' -v`
Expected: FAIL (functions do not exist).

- [ ] **Step 3: Implement `internal/editor/copy.go`**

Use `clearBytes` where needed but the copy is a straight `copy(out[fixed+TypeOffset:fixed+FixedRegionSize], srcRegion)`. `FindCopyConflicts` reuses `isReplayOrCamera`; implement a small `buttonsOverlap(a, b []int) bool`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/editor/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/editor/copy.go internal/editor/copy_test.go
git commit -m "feat: byte-exact binding copy and multi-button conflict detection"
```

---

### Task 2: Copy plan/apply logic in `internal/ops`

**Files:**
- Create: `internal/ops/copy.go`
- Test: `internal/ops/copy_test.go`

**Interfaces:**
- Consumes: `editor.CopyBinding`, `editor.FindCopyConflicts`, `cfgparse.Parse`/`File`, `fileops.Diff`/`DiffSummary`, `calib.Parse`, `profileHasDevice` (existing).
- Produces:
  - `ops.PlanCopy(profDir, sourceID, targetID string, controls []string, force bool) (PlanResult, error)` — reuses the existing `PlanResult{Conflicts, Warning, Diff, Force}`. For each control: locate the source record, copy it, find conflicts, collect warnings; apply all copyable controls in sequence to compute the combined diff. `Warning` is the per-control warnings joined with `"; "`.
  - `ops.ApplyCopy(profDir, sourceID, targetID string, controls []string) ([]byte, error)` — returns the target's new `controls.cfg` bytes after applying the copyable controls.
  - Both share unexported `copyBindings(profDir, sourceID, targetID string, controls []string) (orig, new []byte, conflicts []editor.Conflict, warnings []string, err error)`.
  - Unexported `sourceRecord(src *cfgparse.File, control string) (cfgparse.Record, error)`:
    - `KindKey`/`KindButton` → the record;
    - `KindUnbound`/`KindAxis`/`KindUnknown` → `errNotCopyable`;
    - not found → `errControlMissing`.
  - Sentinels: `var errControlMissing = errors.New("ops: control not found in source profile")`, `var errNotCopyable = errors.New("ops: control binding is not copyable")`.

- [ ] **Step 1: Write failing tests** — `internal/ops/copy_test.go` (package `ops`, temp-dir fixtures):

Helper `writeProfile(t, profDir, id, cfgName string, edits map[string]editor.Binding) string` copies `../cfgparse/testdata/<cfgName>` to `profDir/<id>/controls.cfg` (plus `../calib/testdata/f4_joycalib.yaml` as `joyCalib.yaml`), then applies each edit via `editor.ApplyBinding` and rewrites. `readFixture(t, cfgName)` returns raw fixture bytes.

  - `TestSourceRecordKey`: `sourceRecord(parse(f4_key_u.cfg), "TractionControlArm")` → `Kind==KindKey, Key=='U'`.
  - `TestSourceRecordButton`: `sourceRecord(parse(f4_btn130.cfg), "TractionControlArm")` → `Kind==KindButton, Buttons==[130]`.
  - `TestSourceRecordMultiButton`: `sourceRecord(parse(f4_key_u.cfg), "TearOffVisor")` → `Buttons==[7,33]` (copyable).
  - `TestSourceRecordUnbound`: `sourceRecord(parse(f4_key_u.cfg), "Ignition")` → `errNotCopyable`.
  - `TestSourceRecordMissing`: `sourceRecord(parse(f4_key_u.cfg), "Nope")` → `errControlMissing`.
  - `TestPlanCopyKeyDiff`: source `f4_key_u.cfg`, target `f4_btn128.cfg`, `["TractionControlArm"]` → `Conflicts` empty, `Warning == ""`, `Diff != "no changes"`.
  - `TestApplyCopyKeyReproducesRealFixture`: `ApplyCopy(["TractionControlArm"])` (source `f4_key_u.cfg` → target `f4_btn128.cfg`) → bytes byte-identical to `readFixture("f4_key_u.cfg")`.
  - `TestApplyCopyButtonReproducesRealFixture`: source `f4_btn130.cfg`, target `f4_key_u.cfg`, `["TractionControlArm"]` → byte-identical to `readFixture("f4_btn130.cfg")`.
  - `TestApplyCopyMultiButtonReproducesRealFixture`: source `f4_key_u.cfg` (TearOffVisor = `[7,33]`); target = `f4_btn128.cfg` with an edit clearing `TearOffVisor` to Btn 9 (single button). `ApplyCopy(["TearOffVisor"])` → reparse target; `TearOffVisor.Buttons == [7,33]`, and its mask region byte-identical to the source's.
  - `TestPlanCopyConflict`: source = `f4_key_u.cfg` with edit `{"Ignition": {Key:'U', HasKey:true}}`; target = unedited `f4_key_u.cfg` (TractionControlArm on U). `PlanCopy(..., ["Ignition"], false)` → exactly one conflict `Control=="TractionControlArm"`.
  - `TestPlanCopyMultiButtonConflict`: source = `f4_key_u.cfg` (TearOffVisor `[7,33]`); target = `f4_btn128.cfg` with edit `{"PushToPass": {Button:7, HasButton:true, ...MOZA GUID}}`. `PlanCopy(..., ["TearOffVisor"], false)` → conflict `Control=="PushToPass"`.
  - `TestPlanCopyWarnsDeviceMissing`: source `f4_btn130.cfg` with edit `{"Ignition": {Button:130, HasButton:true, DeviceInstanceGUID:"{11111111-2222-3333-4444-555555555555}", ProductGUID:"{0002346E-0000-0000-0000-504944564944}"}}`; target `f4_key_u.cfg`. `PlanCopy(..., ["Ignition"], false)` → `Warning` non-empty.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/ops/ -run 'Copy|SourceRecord' -v`
Expected: FAIL.

- [ ] **Step 3: Implement `internal/ops/copy.go`**

`copyBindings` reads+parses the source `controls.cfg` once (missing → error), reads+parses the target, then iterates `controls` in order: `sourceRecord` (err → append warning and skip); `conflicts = append(conflicts, editor.FindCopyConflicts(curFile, control, rec)...)`; for a button source, if `profileHasDevice(target, rec.Device.InstanceGUID)` is false append the missing-device warning; `newData, _, err := editor.CopyBinding(curFile, control, rec)` (err → warning and skip); reparse `curFile`. `PlanCopy` returns `PlanResult{Conflicts, Warning: strings.Join(warnings, "; "), Diff: fileops.DiffSummary(fileops.Diff(orig, new)), Force: force}`. `ApplyCopy` returns `new`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/ops/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ops/copy.go internal/ops/copy_test.go
git commit -m "feat: bulk binding copy plan/apply in ops"
```

---

### Task 3: Copy endpoints + `buttons` on controls DTO

**Files:**
- Modify: `internal/server/api.go` (add `buttons` to `controlDTO` and populate it in `handleProfileControls`; add `copyReq` DTO, `handlePlanCopy`, `handleApplyCopy`; register routes)
- Test: `internal/server/api_test.go`

**Interfaces:**
- Consumes: `ops.PlanCopy`, `ops.ApplyCopy`, `fileops.Backup`/`AtomicWrite`/`Diff`/`DiffSummary`, `simcheck.Running` (existing).
- Produces:
  - `controlDTO` gains `Buttons []int \`json:"buttons"\``; `handleProfileControls` sets `Buttons: rec.Buttons` for button records and `[]int{}` otherwise (never `null`).
  - `type copyReq struct { SourceProfileID string; Controls []string; ProfileIDs []string; Force bool; ProfileOverrides []profileOverride }` with the same `forceFor` method as `bindingReq`.
  - Routes: `POST /api/plan/copy` → `{"plans":[planItem...]}` (reuses `planItem`); `POST /api/apply/copy` → `{"applied":[appliedItem...], "skipped":[skippedItem...]}`.
  - `handleApplyCopy`: validate source + targets via `validProfileID`; sim check → 409; for each target: `PlanCopy` → conflicts + `!force` → skip `"binding conflict"`; else `ApplyCopy`, `fileops.Backup`, `fileops.AtomicWrite`, diff into `applied`.

- [ ] **Step 1: Write failing tests** — add to `internal/server/api_test.go`:

Helper `writeProfile(t, dir, id, cfgName string, edits map[string]editor.Binding)` (mirror Task 2; import `editor`).

  - `TestControlsListExposesButtons`: GET `/api/profiles/f4/controls` on `f4_btn130.cfg` → `TractionControlArm.buttons == [130]`; on `f4_key_u.cfg` → `TractionControlArm.buttons == []`.
  - `TestPlanCopyReturnsPlans`: profiles `src`=`f4_key_u.cfg`, `dst`=`f4_btn128.cfg`; POST `/api/plan/copy` `{sourceProfileId:"src", controls:["TractionControlArm"], profileIds:["dst"]}` → 200, one plan, empty conflicts.
  - `TestApplyCopyWritesTarget`: same; POST `/api/apply/copy` → `applied` has `dst`; read `dst/controls.cfg` → byte-identical to `../cfgparse/testdata/f4_key_u.cfg`.
  - `TestApplyCopySkipsConflict`: source `src`=`f4_key_u.cfg` with `{"Ignition":{Key:'U',HasKey:true}}`; target `dst`=`f4_key_u.cfg`; apply `["Ignition"]` force `false` → skipped; force `true` → applied.
  - `TestApplyCopyBlockedWhenSimRunning`: `SimCheck` true → 409; `dst/controls.cfg` bytes unchanged.
  - `TestApplyCopyRejectsTraversal`: `profileIds:[".."]` → 400.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/server/ -run 'Copy|Buttons' -v`
Expected: FAIL.

- [ ] **Step 3: Implement endpoints and DTO change**

Follow `handlePlanBinding`/`handleApplyBinding` as the template. `Buttons` uses `make([]int, 0)` for non-button records; `applied`/`skipped` initialized as empty slices.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/server/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/server/api.go internal/server/api_test.go
git commit -m "feat: copy plan/apply endpoints and buttons on controls DTO"
```

---

### Task 4: Frontend API + store helpers

**Files:**
- Modify: `web/src/lib/api.ts`, `web/src/lib/store.ts`
- Test: `web/src/lib/api.test.ts`, `web/src/lib/store.test.ts`

**Interfaces:**
- Consumes: `request`, `token` (existing).
- Produces:
  - `ControlDTO` gains `buttons: number[]`.
  - `export interface CopyReq { sourceProfileId: string; controls: string[]; profileIds: string[]; force: boolean; profileOverrides: { profileId: string; force: boolean }[] }`.
  - `api.planCopy: (req: CopyReq) => request<PlanResp>('POST', '/api/plan/copy', req)`.
  - `api.applyCopy: (req: CopyReq) => request<ApplyResp>('POST', '/api/apply/copy', req)`.
  - `store.copyableControl(control: ControlDTO): boolean` — `kind === 'key' || kind === 'button'`.
  - `store.buildCopyReq(sourceProfileId, controls, profileIds, force, forceOverrides): CopyReq`.
  - `store.currentBinding` updated: for `kind === 'button'` with `buttons.length > 1`, render `Btn 7 + 33 (device)`; otherwise unchanged.

- [ ] **Step 1: Write failing tests**
  - `api.test.ts`: `api.planCopy`/`api.applyCopy` POST the right path and return typed results (mock `fetch`).
  - `store.test.ts`: `copyableControl` true for `{kind:'key'}` and `{kind:'button', buttons:[7,33]}`; false for `{kind:'axis'}` and `{kind:'unbound'}`. `buildCopyReq` maps `forceOverrides` to `profileOverrides`. `currentBinding` for `{kind:'button', buttons:[7,33], deviceName:'MOZA R9 Base'}` → `"Btn 7 + 33 (MOZA R9 Base)"`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd web && npx vitest run src/lib`
Expected: FAIL.

- [ ] **Step 3: Implement**

Add `buttons: number[]` to `ControlDTO`; update the `control()` test helper to include `buttons: []`. Add `CopyReq`, the two `api` methods, the two `store` helpers, and the `currentBinding` multi-button branch.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd web && npx vitest run src/lib`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src/lib/api.ts web/src/lib/store.ts web/src/lib/api.test.ts web/src/lib/store.test.ts
git commit -m "feat: copy API, copyable/store helpers, multi-button display"
```

---

### Task 5: Sidebar + ControlsTable components

**Files:**
- Create: `web/src/components/Sidebar.svelte`, `web/src/components/ControlsTable.svelte`
- Test: `web/src/components/Sidebar.test.ts`, `web/src/components/ControlsTable.test.ts`

**Interfaces:**
- Consumes: `ProfileDTO`, `ControlDTO`, `currentBinding`, `copyableControl` (existing/new).
- Produces:
  - `Sidebar.svelte` props `{ profiles: ProfileDTO[]; activeId: string; onSelect: (id: string) => void }`. One button per profile (baseline badge), `activeId` highlighted; click calls `onSelect(p.id)`.
  - `ControlsTable.svelte` props `{ controls: ControlDTO[]; selected: Record<string, boolean>; search: string; onSearch: (v: string) => void; onToggle: (name: string) => void; onToggleAll: () => void; onEdit: (name: string) => void; onCopySelected: () => void }`.
    - Filters by `search` (case-insensitive substring on `name` and `displayName`).
    - One row per control: checkbox (disabled unless `copyableControl(c)`), display name, `currentBinding(c)`, an "Edit" button (`onEdit(c.name)`).
    - Header "Select all" checkbox and a "Copy selected →" button (`onCopySelected()`).

- [ ] **Step 1: Write failing tests**
  - `Sidebar.test.ts`: mount with 3 profiles; assert 3 items, baseline badge; click `f4` → `onSelect("f4")`.
  - `ControlsTable.test.ts`: mount with `[{name:'A',kind:'key'}, {name:'B',kind:'button',buttons:[131]}, {name:'C',kind:'button',buttons:[7,33]}, {name:'D',kind:'axis'}]`; assert A, B, C checkboxes enabled, D disabled; search filters; "Edit" on A calls `onEdit('A')`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd web && npx vitest run src/components`
Expected: FAIL.

- [ ] **Step 3: Implement the two components**

Follow existing conventions (`$props()`, `$state`, `$derived`, Svelte 5 handlers). Use `currentBinding` and `copyableControl` from `../lib/store`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd web && npx vitest run src/components`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src/components/Sidebar.svelte web/src/components/ControlsTable.svelte web/src/components/Sidebar.test.ts web/src/components/ControlsTable.test.ts
git commit -m "feat: sidebar and controls table components"
```

---

### Task 6: ProfilePage + CopyModal components

**Files:**
- Create: `web/src/components/ProfilePage.svelte`, `web/src/components/CopyModal.svelte`
- Modify: `web/src/components/CalibCopy.svelte` (optional fixed-source prop)
- Test: `web/src/components/ProfilePage.test.ts`, `web/src/components/CopyModal.test.ts`

**Interfaces:**
- Consumes: `ProfileDTO`, `ControlDTO`, `api.planCopy`/`api.applyCopy`, `highlight`, `copyableControl`.
- Produces:
  - `ProfilePage.svelte` props `{ profile: ProfileDTO; profiles: ProfileDTO[]; controls: ControlDTO[]; onEditControl: (name: string) => void }`.
    - Owns tab state (`'controls' | 'calib'`), search, `selected` map, and copy-modal open state.
    - Controls tab: `<ControlsTable>` + a footer count of selected copyable controls.
    - Calib tab: `<CalibCopy source={profile.id} …/>`.
    - "Copy selected →" opens `<CopyModal sourceId={profile.id} controls={selectedNames} profiles={profiles} onClose={...} />`.
  - `CopyModal.svelte` props `{ profiles: ProfileDTO[]; sourceId: string; controls: string[]; onClose: () => void }`.
    - Targets = `profiles.filter(p => p.id !== sourceId)`, default all selected, with a "select all" toggle.
    - "Preview" → `api.planCopy`; per-profile rows (id, `diffSummary`, conflicts via `highlight`, `warning`) with per-profile overwrite + global "overwrite all".
    - "Apply" → `api.applyCopy`; result message; "Close" calls `onClose`.

- [ ] **Step 1: Write failing tests**
  - `ProfilePage.test.ts`: mount; default tab shows the controls table; "Calibração" tab shows `CalibCopy`; selecting a copyable control then "Copy selected →" opens `CopyModal`.
  - `CopyModal.test.ts`: mock `fetch` for `/api/plan/copy`/`/api/apply/copy`; assert the source profile is not listed as a target; "select all" toggles targets; "Preview" calls `planCopy` with `profileIds` excluding the source; "Apply" calls `applyCopy`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd web && npx vitest run src/components`
Expected: FAIL.

- [ ] **Step 3: Implement**

Reuse `highlight`; build the request via `buildCopyReq`. Render the result inline (not a nested modal).

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd web && npx vitest run src/components`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src/components/ProfilePage.svelte web/src/components/CopyModal.svelte web/src/components/ProfilePage.test.ts web/src/components/CopyModal.test.ts web/src/components/CalibCopy.svelte
git commit -m "feat: profile page and copy modal"
```

---

### Task 7: BindingModal + App restructure

**Files:**
- Create: `web/src/components/BindingModal.svelte`
- Modify: `web/src/App.svelte`, `web/src/App.test.ts`
- Test: `web/src/components/BindingModal.test.ts` (new), `web/src/App.test.ts` (rewritten)

**Interfaces:**
- Consumes: `BindingWizard`, `ProfileList`, `PreviewModal`, `api.planBinding`/`applyBinding`, `loadControlsByProfile`.
- Produces:
  - `BindingModal.svelte` props `{ profiles: ProfileDTO[]; controls: ControlDTO[]; presetControl: string; onClose: () => void }`.
    - Extracts the existing binding flow state (controlName, binding, selected targets, plan, force, forceOverrides, controlsByProfile) from `App.svelte`.
    - `controlName` initialized to `presetControl`; selecting a control loads `controlsByProfile` via `loadControlsByProfile`.
    - Renders `BindingWizard`, `ProfileList`, `PreviewModal` in a modal with a close button.
  - `App.svelte` new structure:
    - Loads `profiles`; `Sidebar` (left) with `activeId`; top bar with "Novo binding" button.
    - `activeControls = api.fetchControls(activeId)` (re-fetch on change); `controlNames` = first `hasControls` profile's controls (once).
    - Main: `<ProfilePage profile={activeProfile} profiles={profiles} controls={activeControls} onEditControl={openBindingModal} />`.
    - `bindingModal: { preset: string } | null`; "Novo binding" → `preset:''`; `onEditControl(name)` → `preset:name`.
    - Renders `<BindingModal>` when open (`profiles`, `controls=controlNames`, `presetControl=preset`).
    - `Settings` rendered once.

- [ ] **Step 1: Write failing tests**
  - `BindingModal.test.ts`: mock `fetch`; mount with `presetControl:'Ignition'` → control select shows `Ignition`; plan/apply call `planBinding`/`applyBinding` with `control:'Ignition'`.
  - `App.test.ts` rewrite: sidebar lists all profiles; clicking `f4` loads its controls; "Novo binding" opens the wizard; "Edit" on a control opens the wizard preset; "Calibração" tab shows calibration copy with targets excluding the active profile.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd web && npx vitest run`
Expected: FAIL.

- [ ] **Step 3: Implement**

Extract the binding state machine into `BindingModal.svelte` verbatim from `App.svelte`. Rewrite `App.svelte` as the sidebar + profile-page shell. `main.ts` unchanged.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd web && npx vitest run`
Expected: PASS.

- [ ] **Step 5: Build the SPA and embed**

Run: `cd web && npm run build`
Expected: `web/dist` updated, no errors.

- [ ] **Step 6: Commit**

```bash
git add web/src/App.svelte web/src/App.test.ts web/src/components/BindingModal.svelte web/src/components/BindingModal.test.ts web/dist
git commit -m "feat: profile-centric shell with binding modal"
```

---

### Task 8: Full verification

- [ ] **Step 1: Backend tests + vet + Windows build**

Run: `go test ./... && go vet ./... && GOOS=windows GOARCH=amd64 go build ./...`
Expected: PASS / no vet output / build succeeds.

- [ ] **Step 2: Frontend tests + build**

Run: `cd web && npx vitest run && npm run build`
Expected: PASS; `web/dist` built.

- [ ] **Step 3: Commit any dist drift**

```bash
git add web/dist && git commit -m "chore: rebuild web dist" || true
```

- [ ] **Step 4: Final note for the author**

Record in the session: manual validation on Windows (copy a real binding between profiles with the simulator closed) remains a follow-up, matching the existing DirectInput/round-trip validation precedent.
