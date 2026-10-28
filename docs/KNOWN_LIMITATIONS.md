# Known Limitations

iBinds is intentionally conservative. This page lists what it does not do yet
and what to watch out for. See the [roadmap](ROADMAP.md) for planned work and
the [FAQ](FAQ.md) for common questions.

## Platform and requirements

- **Windows only.** iRacing is Windows-only; the Linux/WSL path is for building
  and testing, not for running the app against a live installation.
- **iRacing must be closed** before applying changes. Writes are blocked while
  the simulator is detected.
- **Profiles must already exist.** Start iRacing once and save a control profile
  before using iBinds.
- **Portable binary.** There is no installer or auto-updater.
- **Unsigned binary.** Windows SmartScreen or antivirus tools may warn. Verify
  the SHA-256 checksum and, when available, the build provenance before running.

## Editing bindings

- **Keyboard assignment is limited to A-Z and 0-9.** Existing special keys
  (F1-F12, numpad, Pause, Print Screen, punctuation) are displayed by name where
  mapped, but cannot yet be assigned through iBinds.
- **Modifier keys are not exposed.** Shift/Ctrl/Alt are not editable in the
  wizard, and editing a key binding may not preserve an existing modifier.
  Treat modifier bindings with care.
- **Multi-button editing is limited.** Existing multi-button bindings can be
  copied, but the wizard creates and edits a single button. There is no
  add/remove-button UI yet.
- **Axis controls are not editable.** Steering, throttle, brake, and other axis
  bindings are read-only.
- **Keyboard layouts are not validated.** AZERTY, QWERTZ, and other layouts may
  behave differently from QWERTY.
- **Conflict detection is namespace-scoped.** Replay and camera controls
  (`Cam*`, `DCam*`, `Rpy*`, `NextRpyCam*`, `PrevRpyCam*`) are treated as a
  separate namespace from driving controls.
- **Conflict warnings are informational.** Review the preview before applying;
  iRacing resolves conflicts on load.

## Calibration

- **Calibration is copied as a whole file.** iBinds never edits calibration
  values; it copies `joyCalib.yaml` unchanged.
- **Device sets are not compared yet.** Copying calibration to a profile with
  different devices can overwrite or remove calibration for a device that is
  not present in the source profile. Review the diff and keep backups.
- **Different hardware is not interchangeable.** Copying calibration between
  different wheel bases, pedals, or button boxes may not produce meaningful
  results.
- **Force feedback and calibration values cannot be edited.** This is a
  deliberate safety choice, not an oversight.

## Profiles and baseline

- **iBinds does not simulate inheritance.** It edits the selected profile files
  directly; it does not model iRacing's baseline-to-per-car inheritance. Changing
  a baseline file does not automatically rewrite car profiles.
- **Baseline detection uses the conventional `Baseline` folder name.** A
  differently named baseline folder may not be marked as baseline.
- **Friendly car names are incomplete.** The current map covers only a small set
  of cars (for example, `f4`); other profiles show their technical folder IDs.
- **Control labels and categories are a snapshot.** They are based on a captured
  in-game list and may lag new iRacing content. Unknown controls fall back to
  technical IDs.
- **Calibration-only profiles have no bindings.** A profile with only
  `joyCalib.yaml` is listed, but it has no `controls.cfg` to edit.

## Hardware and capture

- **DirectInput button mapping has not been validated on all devices.** The
  conversion from a DirectInput button index to iRacing's "Dev N Btn M" may
  differ by device. Test capture with your hardware before relying on it.
- **POV hats are mapped to buttons 128-159.** Support for every device, hat
  count, or diagonal direction is not guaranteed.
- **High button counts may behave differently.** Devices with more than 128
  buttons, multiple POVs, or unusual HID descriptors should be tested and
  reported.
- **Non-Windows enumeration is limited.** Device button counts may be unknown
  when the app is run outside Windows; the app is intended for Windows.

## Writes, backups, and recovery

- **Backups are automatic but manual to restore.** Every write creates a
  timestamped `.bak-YYYYMMDD-HHMMSS` file next to the original. There is no
  in-app restore/undo UI yet: close iRacing and copy the backup over the original.
- **Backups are not pruned.** They can accumulate over time.
- **Concurrent instances are not locked.** Avoid running two iBinds instances
  against the same profiles folder at the same time.
- **The simulator check uses the standard executable name.** If your build uses
  a different executable name, close iRacing manually before applying changes.
- **The format is reverse-engineered.** A future iRacing update may change
  `controls.cfg`; keep backups and report new fixtures if parsing fails.

## Local server and privacy

- **The app is local-only.** It runs an HTTP server on `127.0.0.1` with a
  per-run session token by default and does not send data to the internet.
- **Do not share the launch URL.** It contains the session token. Avoid showing
  the address bar in screenshots or recordings; close the app to end the session.
- **Settings and logs are local.** They live under `%APPDATA%\ibinds` on
  Windows; logs stay on your machine.

## Legal and AI

- **iBinds is not affiliated with iRacing.** Review iRacing's terms. The tool
  edits user configuration files with the simulator closed and does not inject
  into the game, read game memory, or automate driving.
- **AI assistance is disclosed.** The project is built with AI assistance and
  human review/tests. Evaluate the source and test results yourself.
- **No warranty.** iBinds is released under the MIT license; use it at your own
  risk and keep backups.
