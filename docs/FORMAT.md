# iRacing `controls.cfg` binary format

Validated against real fixtures (`internal/cfgparse/testdata/`) captured from
the author's iRacing profiles (F4 car, Oct 2026 game version). The fixtures
are the source of truth; this document records what the bytes actually say.

## Fixtures

| File | `TractionControlArm` binding | Notes |
|---|---|---|
| `f4_btn130.cfg` | Btn 130 (MOZA R9 Base) | mask word `04 00 00 00` |
| `f4_btn128.cfg` | Btn 128 (MOZA R9 Base) | mask word `01 00 00 00` (only byte differing from btn130 is at `0xbe0`) |
| `f4_key_u.cfg` | key `U` | type `04`, key byte `55` at `0xbd0` |
| `f4_btn130.cfg` vs `f4_btn128.cfg` vs `f4_key_u.cfg` | — | all three are 30584 bytes; they differ only in the `TractionControlArm` record |

`f4_key_u.cfg` is byte-identical to the author's current live
`f4/controls.cfg` (the profile was left at key `U` after re-snapshotting).
The plan's original "current state" description (`TractionControlArm='I'`)
no longer matches the live file; `TractionControlArm='U'`, `LowFuelAccept='N'`
and `Ignition` (unbound) are all still exercised by `f4_key_u.cfg`.

## File layout

| Offset | Size | Meaning |
|---|---|---|
| `0x0` | 4 | magic `GFCC` (`47 46 43 43`, i.e. `"CCFG"` stored little-endian) |
| `0x4` | 4 | `14 00 00 00` (version, observed `0x14`) |
| `0x8` | 4 | size field = `len(file) - 12` |
| `0x9f` | 4 | chunk magic `LRTC` (`4c 52 54 43`, i.e. `"CTRL"` reversed) |
| `0xa3` | 8 | chunk header (`08 00 00 00` + `cd 76 00 00` = `0x76cd`), not yet fully decoded |
| `0xab` | — | records begin (12 bytes after the `LRTC` magic) |

## Record framing

Each record is a NUL-terminated printable name followed by a fixed 68-byte
region:

```
record := name NUL fixed[68]
```

Consecutive records are packed tightly (the next name begins immediately
after the previous 68-byte region). Walking `name+NUL+68` from `0xab`
reproduces exactly 369 records ending exactly at EOF. The fixed region is
type-dependent; offsets below are relative to its start (`PayloadStart`).

## Fixed region (68 bytes)

| Offset | Size | Meaning |
|---|---|---|
| `+0` | 4 | flags (observed `1` or `0`; not required for editing) |
| `+4` | 4 | category (`6` = digital, `25` = steering axis, `31` = analog, `57` = level; not required for editing) |
| `+8` | 4 | **binding type**: `0` = unbound, `1` = axis, `2` = button, `4` = key |
| `+12` | 20 | union: key byte (`+12`, for key) **or** five 32-bit button mask words `+12..+31` (buttons 0..159) |
| `+32` | 4 | reserved (observed `0`) |
| `+36` | 16 | InstanceGUID (button/axis records) |
| `+52` | 4 | ProductID (button records) |
| `+56` | 6 | reserved (observed `0`) |
| `+62` | 6 | `PIDVID` literal (`50 49 44 56 49 44`) for button records |

### Binding types

`+8` is a little-endian uint32:

- `0` → unbound
- `1` → axis (steering/throttle/brake/clutch — read-only in v1)
- `2` → button
- `4` → key (ASCII byte at `+12`)

### Button bitmask

Five 32-bit words at `+12..+31` cover buttons `0..159`; button N is bit
`N % 32` of word `N / 32` (word `0` at `+12`, word `4` at `+28` = `0xbe0`).
`Btn 130` = word `4`, value `4`; `Btn 128` = word `4`, value `1`.

> The spec's "word 4 at `0xbe0`" was a partial observation from the author's
> Btn-128/130 experiment. The full mask is five words, and some real records
> bind multiple buttons (e.g. `TearOffVisor` = Btn 7 + Btn 33). The v1 parser
> reports the lowest set button in `Record.Button`; all mask bytes are
> preserved via `Opaque`.

### GUIDs

InstanceGUID (`+36`, 16 bytes) uses the standard Windows GUID layout
(`Data1` LE uint32, `Data2`/`Data3` LE uint16, `Data4` 8 bytes), rendered as
`{8-4-4-4-12}` uppercase hex. MOZA R9 Base:

- bytes `d0 05 18 99 b6 2e f1 11 80 0a 44 45 53 54 00 00`
- string `{991805D0-2EB6-11F1-800A-444553540000}`

ProductGUID is not stored as a GUID; the file holds the 4-byte ProductID
(`+52`, LE) followed by the `PIDVID` literal. It is rendered as
`{<ProductID:08X>-0000-0000-0000-504944564944}` (the tail `504944564944`
is `"PIDVID"` hex). MOZA: ProductID `0x0002346E` → `{0002346E-0000-0000-0000-504944564944}`.

## Key/button edits (writer contract)

To set a binding on a record (see `internal/editor`):

- key: type `=4`, key byte `+12` = ASCII, mask words `+12..+31` = 0
- button: type `=2`, key byte `+12` = 0, mask word `(button/32)` = `1 << (button%32)` (other words 0), GUID at `+36`, ProductID at `+52`, `PIDVID` at `+62`

The simulator's real `btn130 → btn128` transition changed only the mask word
(`4` → `1`); its `key U` transition wrote type `4`, key byte `55`, and zeroed
the mask/GUID/ProductID/`PIDVID` regions.

## Device capture (DirectInput)

`internal/inputcap` captures button presses via DirectInput 8 (`dinput8.dll`).
The COM vtables (`IDirectInput8`, `IDirectInputDevice8`) are hand-written
(go-ole has no DirectInput bindings); they compile for `GOOS=windows` but were
not exercised on the author's machine at the time of writing — see the manual
validation note in the plan's Task 9.

- The device `InstanceGUID` string is rendered from `DIDEVICEINSTANCE.guidInstance`.
- The `ProductGUID` string is derived from `DIDEVCAPS.dwProductId` as
  `{<pid:08X>-0000-0000-0000-504944564944}` (the trailing `504944564944` is
  `"PIDVID"` hex). This matches the real `joyCalib.yaml` ProductGUID
  (e.g. MOZA `dwProductId` `0x0002346E` → `{0002346E-0000-0000-0000-504944564944}`).
  The plan's `{0000<pid:%04X>-…}` wording was imprecise; the full 32-bit
  ProductID is used.
- **Offline fallback:** when DirectInput is unavailable (non-Windows, or no
  device), the server's `GET /api/input/devices` falls back to listing the
  devices from the profiles' `joyCalib.yaml` files (`ButtonCount` unknown/0),
  and capture mode degrades to the device/button dropdown. This decision is
  recorded deliberately — the DirectInput capture path still requires manual
  validation on Windows (press a wheel button and check the reported index
  against iRacing's "Dev N Btn M").

