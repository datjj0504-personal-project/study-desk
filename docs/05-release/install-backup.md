# Windows Installation, Backup, and Release

## Requirements

- Windows 10 or 11 (64-bit)
- Microsoft Edge WebView2 Runtime

Go, Node.js, and the Wails CLI are only required to build StudyDesk from source; end users do not need them to run the prebuilt application.

## Install, update, and uninstall

Download `StudyDesk-Windows-x64-<version>.zip` from a trusted release and extract it. The archive contains `StudyDesk-Setup-<version>.exe`; run the installer and approve the Windows administrator prompt. The installer targets 64-bit Windows 10/11, installs under Program Files, creates a Start Menu shortcut, and offers an optional Desktop shortcut.

The installer bundles Microsoft's small [Evergreen WebView2 bootstrapper](https://learn.microsoft.com/en-us/microsoft-edge/webview2/concepts/evergreen-vs-fixed-version) and runs it silently. The bootstrapper installs or updates the shared WebView2 Runtime from Microsoft servers, so an internet connection is required if it must download the runtime. Windows 11 and many eligible Windows 10 machines already have the runtime, but the installer does not assume that.

The first launch creates the local application data directory and SQLite database. The installer registers StudyDesk with Windows Apps/Programs for uninstall. Uninstall removes the app and shortcuts but leaves `%APPDATA%\StudyDesk` intact. To update, run the newer setup executable; the installer upgrades the same app installation and database migrations run when the app starts. Back up the data before updating.

There is no auto-updater. The installer is not code-signed; Windows SmartScreen or other security software may display an untrusted-publisher warning.

## User data location

StudyDesk stores its database at:

```text
%APPDATA%\StudyDesk\studydesk.db
```

On a typical Windows installation, `%APPDATA%` expands to `C:\Users\<user>\AppData\Roaming`. The database contains notes, subjects, study sessions, timer preferences, and theme preference. The app does not sync this data to a server.

## Back up data

Close StudyDesk normally before copying the data. SQLite may use write-ahead log sidecar files while the app is open; copying only the `.db` file during use can produce an incomplete backup. Copy the complete `StudyDesk` directory after the application has exited.

In PowerShell:

```powershell
$dataDirectory = Join-Path $env:APPDATA 'StudyDesk'
$backupRoot = Join-Path $HOME 'StudyDeskBackups'
$backupFile = Join-Path $backupRoot ("StudyDesk-" + (Get-Date -Format 'yyyyMMdd-HHmmss') + ".zip")

if (Get-Process -Name studydesk -ErrorAction SilentlyContinue) {
    throw 'Close StudyDesk before creating a backup.'
}
if (-not (Test-Path -LiteralPath (Join-Path $dataDirectory 'studydesk.db'))) {
    throw "StudyDesk database not found at $dataDirectory"
}

New-Item -ItemType Directory -Path $backupRoot -Force | Out-Null
Compress-Archive -Path (Join-Path $dataDirectory '*') -DestinationPath $backupFile
```

The backup contains private study notes and history. Store it somewhere access-controlled; use an encrypted drive or archive if it leaves your device. Keep more than one backup if the data is important.

## Restore data

Restoring replaces the current StudyDesk data with the chosen backup. Close the app first. The following example moves the current data directory aside as a recovery copy, then extracts a backup archive into a fresh data directory. Set `$backupFile` to the chosen ZIP file.

```powershell
$dataDirectory = Join-Path $env:APPDATA 'StudyDesk'
$backupFile = 'D:\Backups\StudyDeskBackups\StudyDesk-20261024-120000.zip'
$recoveryDirectory = Join-Path $env:APPDATA ("StudyDesk-before-restore-" + (Get-Date -Format 'yyyyMMdd-HHmmss'))

if (Get-Process -Name studydesk -ErrorAction SilentlyContinue) {
    throw 'Close StudyDesk before restoring a backup.'
}
if (-not (Test-Path -LiteralPath $backupFile)) {
    throw "Backup archive not found at $backupFile"
}

if (Test-Path -LiteralPath $dataDirectory) {
    Move-Item -LiteralPath $dataDirectory -Destination $recoveryDirectory
}
New-Item -ItemType Directory -Path $dataDirectory | Out-Null
Expand-Archive -LiteralPath $backupFile -DestinationPath $dataDirectory
```

Launch StudyDesk after restoration and check that your notes and settings are present. Keep the recovery directory until the restored data has been verified.

## Build an installer and release ZIP

On a Windows build machine, install:

- Go **1.27.2 or later**, as required by `go.mod`.
- Node.js LTS and npm.
- Wails CLI **v2.17.0** (`go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0`).
- Inno Setup 6 ([download](https://jrsoftware.org/isinfo.php)), with `ISCC.exe` on `PATH` or installed in its default `Program Files` folder.
- Internet access to download the signed WebView2 bootstrapper during release and the Evergreen Runtime during installation.

Then run from the repository root:

```powershell
.\build-tools\release-windows.ps1
```

The script runs Go and frontend tests/builds, runs `wails build -platform windows/amd64`, downloads the WebView2 Evergreen bootstrapper from Microsoft and verifies its Authenticode signature, compiles [installer.iss](../../build/windows/installer.iss) with Inno Setup, and verifies the ZIP contains the generated installer. It reads the app version from the single `appVersion` constant in `app.go`; update that constant for a release rather than maintaining a second version.

```text
release\
  README.md
  StudyDesk-Setup-<appVersion>.exe
  StudyDesk-Windows-x64-<appVersion>.zip
```

The script stages intermediate files in a unique temporary directory. It only replaces the generated installer and ZIP with names matching the current app version and preserves other files already in `release/`. `build\bin\studydesk.exe` is generated by Wails and is not included as a separate portable artifact in the ZIP.

## Branding and release verification

No StudyDesk-specific logo/icon is currently provided. The existing `build\appicon.png` is the Wails logo; it is deliberately not used as installer branding. The installer uses Inno Setup's default setup icon, and the Wails executable may still show its generated default application icon.

The current environment has not been verified with a clean-machine install/uninstall test. Before publishing a release, install from the ZIP on a clean Windows 10/11 machine and verify WebView2 setup, app launch, Start Menu/Desktop shortcuts, Windows uninstall registration, and that uninstall preserves `%APPDATA%\StudyDesk\studydesk.db`.

- No auto-updater or code-signing configuration is provided.
- The installer configuration and release ZIP script are present, but actual installer compilation and clean-machine install/uninstall verification remain outstanding until Inno Setup and a test machine are available.
- The installer is not signed. Distribute it over a trusted channel and communicate its checksum separately if recipients need integrity verification.
- Users are responsible for backing up their local database; it is not included with the installer or ZIP.

To calculate a SHA-256 checksum for the installer or ZIP:

```powershell
Get-FileHash -Algorithm SHA256 -LiteralPath 'release\StudyDesk-Setup-<version>.exe'
Get-FileHash -Algorithm SHA256 -LiteralPath 'release\StudyDesk-Windows-x64-<version>.zip'
```
