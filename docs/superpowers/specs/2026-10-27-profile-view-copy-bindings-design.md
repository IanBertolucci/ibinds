# iBinds — Profile-centric view + bulk binding copy — Design

Date: 2026-10-27
Status: approved

## 1. Intent and why

After recalibrating his wheel, the author's left D-pad inputs were renumbered
(Btn 30 → 131–134), silently dropping every binding that used them. He
re-bound them in one profile inside the simulator and wanted to propagate just
those bindings to his other profiles. The current app cannot do that: it only
applies a brand-new binding across profiles, or copies a whole `joyCalib.yaml`.

We want the app to become profile-centric: a left sidebar lists every profile;
clicking one opens that profile's page showing every control and what it is
bound to, with the ability to select several controls and copy their bindings
to other profiles in bulk.

## 2. Goals

- Sidebar navigation: every profile (baseline + cars) as a list item; the
  profile page is the default home screen.
- Profile page with two tabs:
  - **Controles** — table of every control with its current binding
    (key/btn/axis/none + device name), a search box, multi-select checkboxes,
    a "copiar selecionados" action, and a per-row "editar" that opens the
    binding wizard preloaded with that control.
  - **Calibração** — "copy this profile's calibration to X or all" (relocated
    existing flow, source fixed to the active profile).
- Bulk binding copy: select several controls on the active profile, choose
  target profiles, preview per-profile conflicts/diff, apply with backup.
- The existing "apply a new binding in bulk" wizard stays, relocated to a
  global "Novo binding" button.

## 3. Non-goals

- No automatic device-ID remapping (migrating Btn 30 → 131 is manual: bind
  once, then copy).
- No in-place single-control binding edit beyond the preloaded wizard (the
  wizard remains the only way to set a binding).
- No editing of calibration values (copy whole file only).
- No support for other simulators.

## 4. Anti-goals (what makes this a failure even if it works)

- Silently dropping bindings during a copy (e.g. copying only the lowest
  button of a multi-button control).
- Writing to the wrong profile without the user noticing.
- Corrupting files without a restorable backup.

## 5. Copy semantics

- Copying a control means: take the **source** profile's current binding for
  that control and write it to the **same control** in each target profile.
- Copyable source bindings are **key** and **button** bindings.
  - A **button** binding is copied **whole**: all buttons in the mask (e.g.
    `TearOffVisor` = Btn 7 + Btn 33 copies both, never just the lowest).
  - Unbound, axis, and unknown controls are not copyable.
- The copy is a byte-exact copy of the source record's **binding region**
  (type + key/mask + GUID + ProductID + PIDVID, offsets `+8..+67` of the fixed
  region), not a re-derivation through the single-button writer. This is what
  makes multi-button copy faithful.
- Conflict detection reuses the same rules as the existing
  `editor.FindConflicts` (key equality; button set overlap on the same device
  GUID; replay/camera namespace scoping), evaluated against each target
  profile's current file.
- If the copied binding is a button whose device is not listed in the target
  profile's `joyCalib.yaml`, reuse the existing missing-device warning; never
  edit calibration.
- Targets exclude the source profile (`profiles - source`), mirroring
  calibration copy.

## 6. Decisions made (as the author chose them)

- Sidebar is for profiles only.
- The binding wizard and calibration copy are reached from the profile page;
  the wizard is also a global "Novo binding" button.
- "Edit a single control in place" = open the wizard preloaded with that
  control (reuse, not a new editor).
- Conflict overwrite: per-profile and batch confirmation, reusing the existing
  force/`profileOverrides` mechanism.

## 7. Left to the builder

- Exact component split and Svelte structure (follow the existing components).
- Exact route/request shape for the copy plan/apply (mirror the existing
  binding/calibration endpoints).
- How the editor's byte-exact binding-region copy and its multi-button
  conflict detection are structured.

## 8. Constraints (unchanged from the original design)

- Windows target; simulator closed before any write; surgical byte-exact
  edits; timestamped backup before every write; `CCFG` magic validated before
  writing; local 127.0.0.1 server + token; single binary; all code/docs/commit
  messages in English.
