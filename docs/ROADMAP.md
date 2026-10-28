# iBinds Roadmap

This is a living, user-facing roadmap. It is not a release schedule or a
commitment: priorities may change based on feedback, hardware testing, and
iRacing updates. Security-sensitive details are tracked privately and are not
listed here.

See also: [Known limitations](KNOWN_LIMITATIONS.md) and [FAQ](FAQ.md).

## How to read this

- **Now** — actively planned; next up.
- **Next** — queued after the Now items.
- **Later** — under consideration; not scheduled.
- **Exploring** — research only; may never ship.
- **Done** — recently completed.

## Now

### Data safety

- **Device-aware calibration copy**: compare the source and target device sets
  and warn before replacing calibration for a profile with different devices.
- **Stronger write safety**: re-check the simulator immediately before each
  write and prevent two iBinds instances from writing at the same time.
- **Preserve modifiers**: keep Shift/Ctrl/Alt modifiers when editing a key and
  include them in conflict detection.
- **Backup restore and retention**: make backups discoverable and restorable
  from the app, and stop them from accumulating forever.
- **Format compatibility**: fail safely on unknown `controls.cfg` layouts, add
  fixtures for new iRacing versions, and publish a compatibility matrix.
- **Hardware validation**: validate DirectInput button and POV mapping on real
  wheels and button boxes; publish a "tested with" matrix.

### Editor and profiles

- **Special keys**: assign F1-F12, numpad, punctuation, Pause, Print Screen,
  and other supported keys instead of only A-Z/0-9.
- **Keyboard layouts**: validate QWERTY, AZERTY, and QWERTZ behavior.
- **Multi-button editing**: create and edit controls that use more than one
  button, not only copy existing multi-button bindings.
- **Baseline detection**: identify the baseline profile robustly and document
  how iRacing uses baseline versus per-car profiles.
- **Car and control labels**: expand friendly car names and add a process for
  reporting missing or outdated control labels.

### Trust and launch readiness

- **Local server hardening**: improve session/token handling and local security
  headers while keeping the app local-only.
- **Windows CI**: run tests and builds on Windows, not only cross-compile from
  Linux.
- **Supply chain**: pin CI dependencies, add vulnerability scanning, and publish
  an SBOM with releases.
- **Launch docs**: FAQ, known limitations, a short demo, a non-affiliation
  disclaimer, and a beta-tester matrix.

## Next

- **Code signing**: evaluate free or low-cost signing options to reduce
  SmartScreen and antivirus friction.
- **Restore/undo UX**: expose backups in the UI with a diff preview and a
  confirmation step.
- **Public roadmap maintenance**: keep this page current and mark items as
  planned, in progress, or done.
- **Community labels**: make it easy to contribute car and control names.
- **Documentation**: troubleshooting for OneDrive-redirected Documents,
  permissions, SmartScreen, and backups.

## Later

- **Axis-binding editing**: research whether iRacing's axis records can be
  edited safely. This may remain unsupported if the semantics cannot be proven.
- **Installer or package manager**: evaluate winget or Scoop after signing.
- **Native webview research**: compare a native shell with the current embedded
  SPA plus local server.
- **Localization**: evaluate community translations after the core docs are
  stable.

## Exploring

- **Force feedback editing**: future research only. iBinds currently copies
  `joyCalib.yaml` whole and never edits calibration values. Editing force
  feedback or calibration values requires a new design and safety model.
- **Automation/CLI workflows**: not planned until the file-safety work is
  complete.

## Done

- Public README with motivation and audience sections.
- Public known limitations, FAQ, and roadmap.
- Internal planning/spec artifacts kept out of the public repository.
