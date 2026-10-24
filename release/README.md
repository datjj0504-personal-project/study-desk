# Release artifacts

Run `.\build-tools\release-windows.ps1` from the repository root on a Windows build machine with Go, Node.js/npm, Wails CLI v2, Inno Setup 6, and internet access for the Microsoft WebView2 Evergreen bootstrapper.

The script creates or updates only the generated artifacts for the current app version and preserves unrelated files in this directory:

- `StudyDesk-Setup-<appVersion>.exe` — Inno Setup installer.
- `StudyDesk-Windows-x64-<appVersion>.zip` — ZIP containing that installer only.

The single version source is `appVersion` in `app.go`. Do not commit generated `.exe` or `.zip` files. See [release instructions](../docs/05-release/install-backup.md).
