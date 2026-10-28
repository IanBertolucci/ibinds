# FAQ

Common questions about iBinds. See also [Known limitations](KNOWN_LIMITATIONS.md)
and the [roadmap](ROADMAP.md).

## General

### What is iBinds?

iBinds is a local Windows tool that edits iRacing control bindings in bulk and
copies calibration between control profiles without opening the simulator. It
reads the existing `controls.cfg` and `joyCalib.yaml` files and applies surgical
edits.

### Is iBinds affiliated with iRacing?

No. iBinds is an independent open-source project. It is not endorsed by or
affiliated with iRacing.

### Is it free?

Yes. iBinds is released under the MIT license.

### Does it send my data anywhere?

No. It runs a local HTTP server on `127.0.0.1` and does not send telemetry or
profile data to the internet.

### Does it modify my iRacing installation?

No. It only modifies profile files under the iRacing profiles folder and its own
settings/log files under `%APPDATA%\ibinds`.

## Safety

### Can it corrupt my files?

iBinds validates the file format, writes atomically, and creates a timestamped
backup before every write. It also blocks writes while the simulator is running.
That said, the format is reverse-engineered: keep a backup of your profiles
folder before first use, especially before a new iRacing update.

### What happens if iRacing changes the format?

The parser is validated against real fixtures, but a future iRacing update may
break it or produce incorrect results. That is the main compatibility risk.
Keep backups and report a fixture so the format can be validated. A stricter
compatibility guard is on the roadmap.

### How do I restore a backup?

Close iRacing, find the `.bak-YYYYMMDD-HHMMSS` file next to the original, and copy
it over the original. Consider saving the current file first. There is no in-app
restore UI yet; it is on the roadmap.

### Can I run it while iRacing is open?

Applying changes is blocked while the simulator is detected. Close iRacing before
applying. If your build uses a nonstandard executable name, close it manually.

### Can I run two instances at the same time?

Avoid it. There is no file lock yet; two instances can race and overwrite each
other's changes.

## Bindings

### Why can't I assign F1, numpad, Pause, or punctuation?

The current wizard only assigns A-Z and 0-9. Existing special-key bindings are
displayed by name when known, but assigning them is on the roadmap.

### Why did my Shift/Ctrl/Alt modifier disappear?

Modifier keys are not exposed in the editor, and editing a key binding may not
preserve an existing modifier. Avoid editing modifier bindings until modifier
support lands.

### Can I create a multi-button binding?

You can copy an existing multi-button binding, but the wizard currently creates
and edits only one button per control. Multi-button creation/editing is planned.

### Can I edit steering, throttle, or brake axis bindings?

No. Axis records are read-only in iBinds.

### How do baseline and per-car profiles work?

iRacing stores a baseline profile plus optional per-car profiles. iBinds edits
the files you select; it does not simulate iRacing's inheritance. Changing a
baseline file does not automatically rewrite car profiles.

### Why do some profiles show technical names?

Friendly car-name coverage is still small (for example, `f4`). Other profiles
show their folder IDs. Control labels are also a snapshot of the in-game list;
unknown controls fall back to technical IDs.

### How does conflict detection work?

It warns when a key or button is already used by another control in the same
namespace. Replay and camera controls are treated separately from driving
controls. Warnings are informational; review the preview before applying.

## Calibration

### Does iBinds edit force feedback or calibration values?

No. It copies `joyCalib.yaml` unchanged. Editing calibration or force feedback
values is a future research item, not a current feature.

### Is copying calibration between profiles safe?

Only when the profiles use the same devices. iBinds copies the whole file and
does not yet compare device sets, so a profile with different hardware can lose
or overwrite calibration. Review the preview and keep backups.

### Why copy the whole file instead of individual axes?

iRacing stores per-device, per-axis calibration. Copying the whole file preserves
the values exactly; merging individual axes is not implemented yet.

### Can I edit force feedback?

Not yet. This is tracked as a future item and would require a new design and
safety model.

## Hardware

### Will it work with my wheel?

Hardware validation is still limited. Test capture with your device before
relying on it, and report the device, button count, and results so it can be
added to the tested-with matrix.

### What are buttons 128-159?

iRacing maps POV hats to virtual buttons `128..159`, eight directions per hat.
iBinds handles that mapping, but not every device or hat configuration has been
validated.

### Does it support button boxes and multiple devices?

Yes, devices are identified by GUID. Capture and binding copy support multiple
devices, but calibration copy has the device-set limitation described above.

## Windows and installation

### Is it Windows-only?

Yes. iRacing is Windows-only; the Linux/WSL workflow is for building and
testing.

### Why does SmartScreen warn me?

The binary is not code-signed. Verify the SHA-256 checksum (and build provenance
when available), or build from source. Do not disable SmartScreen.

### Does it work with OneDrive-redirected Documents?

Yes. On Windows, iBinds resolves the real Documents folder through the Windows
Known Folder API.

### Where are settings and logs?

Under `%APPDATA%\ibinds` on Windows. On Linux, they follow
`$XDG_CONFIG_HOME/ibinds`.

### Is there an installer?

No. iBinds is a portable `.exe`; there is no installer or auto-updater.

## Privacy and security

### Why does the launch URL contain a token?

The token authenticates the local web UI to the local API. It is generated per
run by default. Do not share the URL or screenshots that show the address bar.

### Is the local server exposed to my network?

No. It binds to `127.0.0.1` only.

### Is iBinds a cheat?

No. It does not inject into the game, read game memory, or automate driving. It
edits user configuration files while the simulator is closed. Review iRacing's
terms if you have concerns.

## Development and community

### Why not use the iRacing SDK?

The SDK does not expose control bindings or `joyCalib.yaml` calibration; those
live in local configuration files.

### Is iBinds built with AI?

Yes, with human review and tests. The source and test suite are public; evaluate
them yourself.

### How do I report a bug or send a fixture?

Open an issue with your iRacing version, wheel/device details, steps to reproduce,
and the relevant error output. If you attach `controls.cfg` or `joyCalib.yaml`,
remove personal information first; device names and GUIDs may be present.

### Can I contribute?

Yes. The project is MIT-licensed. Issues, hardware reports, fixtures, and pull
requests are welcome; include tests for behavior changes.
