# Releasing iBinds

iBinds uses semantic versioning (`vMAJOR.MINOR.PATCH`) and publishes Windows
binaries from GitHub Actions. The version comes from the git tag; no version
file needs editing.

## Pre-release checklist

- [ ] `go test ./...` passes.
- [ ] `go vet ./...` passes.
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` passes.
- [ ] `cd web && npx vitest run` passes.
- [ ] `cd web && npm run build` passes with no Svelte warnings.
- [ ] `./build.sh` (Linux/WSL) or `.\build.ps1` (Windows) produces
      `ibinds.exe`.
- [ ] `ibinds.exe --version` prints the expected version.
- [ ] On Windows, with iRacing closed, test: open the UI, change the default
      device in Settings, edit one binding, copy one binding, and copy
      calibration. Confirm the default device is preselected and the backups
      exist.
- [ ] Search for a category (for example `cockpit`) and confirm the controls
      list filters to that section.
- [ ] Update `README.md` if flags, prerequisites, or behavior changed.

## Tag and publish

```bash
git tag -a v0.1.0 -m "iBinds v0.1.0"
git push origin v0.1.0
```

The tag push starts `.github/workflows/release.yml`. The workflow:

1. Checks out the tagged source.
2. Runs backend tests/vet and frontend tests/build.
3. Cross-compiles `ibinds.exe` with `main.version`, `main.commit`, and
   `main.buildDate`.
4. Writes `SHA256SUMS.txt`.
5. Creates a GitHub build-provenance attestation for `ibinds.exe`.
6. Publishes `ibinds.exe` and `SHA256SUMS.txt` to the GitHub release.

GitHub automatically adds `Source code (zip)` and `Source code (tar.gz)` to the
release assets.

## Manual workflow run

Use **Actions → Release → Run workflow** and enter an existing tag. The tag must
already exist, because the workflow checks out that ref.

## Verify the release

On a clean Windows machine:

```powershell
Get-FileHash .\ibinds.exe -Algorithm SHA256
Get-Content .\SHA256SUMS.txt
.\ibinds.exe --version
```

Then run the Windows smoke test from the pre-release checklist.

If you have the GitHub CLI:

```powershell
gh attestation verify .\ibinds.exe --repo IanBertolucci/ibinds
```

## If the workflow fails

- Re-run the failed job from the Actions page.
- If the release was partially created, delete the release (not necessarily the
  tag) and re-run before publishing again.
- If provenance attestation fails, confirm the workflow has `id-token: write`
  and `attestations: write` permissions.
- If the Windows build fails, run the same Go build locally with
  `GOOS=windows GOARCH=amd64`.

## Code signing

The project does not currently pay for code signing. SmartScreen may warn on
first run. Do not tell users to disable SmartScreen; point them to the SHA-256
checksum, the build-provenance attestation, and the source archive. If they
choose to run the binary, use **More info → Run anyway** only after the checksum
matches.

Future options, only after an explicit decision:

- Azure Artifact Signing or an EV certificate for Authenticode.
- winget or Scoop manifests after signing.
- MSIX/Microsoft Store packaging.

## Release notes template

The release workflow publishes this structure automatically. If you edit the
release manually, keep these sections:

```markdown
## Install (Windows)

1. Download `ibinds.exe`.
2. Verify the SHA-256 checksum against `SHA256SUMS.txt`.
3. Run `ibinds.exe`. If the browser does not open, use the URL printed in the
   terminal.

## Verify the download

Get-FileHash .\ibinds.exe -Algorithm SHA256
Get-Content .\SHA256SUMS.txt
gh attestation verify .\ibinds.exe --repo IanBertolucci/ibinds

## Build from source

Download the source zip, install Go 1.27.2+ and Node.js 20.19+ (or 22.12+),
then run:

powershell -NoProfile -ExecutionPolicy Bypass -File .\build.ps1

## Notes

List behavior changes, known limitations, and anything users should test.
```
