#!/usr/bin/env bash
# Cross-compile iBinds from Linux/WSL: builds the Svelte UI, then the Windows
# binary. Mirrors build.ps1 but needs no Windows shell.
set -euo pipefail
cd "$(dirname "$0")"

GO_BIN="${GO_BIN:-}"
if [ -z "$GO_BIN" ]; then
  if command -v go >/dev/null 2>&1; then
    GO_BIN="$(command -v go)"
  elif [ -x "$HOME/.local/go/bin/go" ]; then
    GO_BIN="$HOME/.local/go/bin/go"
  else
    echo "error: go not found; install Go 1.22+ or set GO_BIN" >&2
    exit 1
  fi
fi

if ! command -v npm >/dev/null 2>&1; then
  echo "error: npm not found; install Node.js 20+" >&2
  exit 1
fi

echo "==> Building web UI"
pushd web >/dev/null
if [ ! -d node_modules ]; then
  npm install --legacy-peer-deps
fi
npm run build
popd >/dev/null

echo "==> Cross-compiling ibinds.exe for Windows"
OUT="${OUT:-ibinds.exe}"
GOOS=windows GOARCH=amd64 "$GO_BIN" build -ldflags "-s -w" -o "$OUT" ./cmd/ibinds

# If the user keeps a copy of the exe in their Windows home (e.g. the one they
# double-click), refresh it too so no manual copy is needed.
if [ "$OUT" = "ibinds.exe" ] && command -v cmd.exe >/dev/null 2>&1 && command -v wslpath >/dev/null 2>&1; then
  win_user="$(cmd.exe /c 'echo %USERPROFILE%' 2>/dev/null | tr -d '\r')"
  if [ -n "$win_user" ]; then
    installed="$(wslpath -u "$win_user")/ibinds.exe"
    if [ -f "$installed" ]; then
      if ! cp "$OUT" "$installed" 2>/dev/null; then
        # Windows locks a running exe; close it and retry.
        if command -v powershell.exe >/dev/null 2>&1 &&
          powershell.exe -NoProfile -Command \
            "Get-Process ibinds -ErrorAction SilentlyContinue | Stop-Process -Force" >/dev/null 2>&1; then
          sleep 1
          cp "$OUT" "$installed"
          echo "==> Stopped running ibinds.exe and updated Windows copy: $installed"
        else
          echo "warning: could not update $installed; close ibinds.exe and rerun ./build.sh" >&2
        fi
      else
        echo "==> Updated Windows copy: $installed"
      fi
    fi
  fi
fi

echo "==> Done: $(cd "$(dirname "$OUT")" && pwd)/$(basename "$OUT")"
