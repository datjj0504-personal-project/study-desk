# Windows Installation, Backup, and Release

## Requirements

- Windows 10 or 11 (64-bit)
- Microsoft Edge WebView2 Runtime

Go, Node.js, and the Wails CLI are only required to build StudyDesk from source; end users do not need them to run the prebuilt application.

## Install and run

StudyDesk currently ships as a portable `studydesk.exe`. There is no MSI, setup wizard, automatic update mechanism, or uninstaller.

1. Obtain `studydesk.exe` from a trusted release or build it from source as described below.
2. Copy the executable to a stable folder writable by your Windows user, such as a folder under `%LOCALAPPDATA%`.
3. Double-click `studydesk.exe` to launch StudyDesk.

The first launch creates the local application data directory and SQLite database. If Windows reports that WebView2 is missing, install the Microsoft Edge WebView2 Runtime and relaunch StudyDesk.

To update a portable installation, close StudyDesk and replace the executable with the newer trusted build in the same folder. The database is stored separately and is not removed by replacing the executable; back it up before an update. StudyDesk applies pending database migrations at startup.

To remove the portable application, close it and delete the executable. This does not remove `%APPDATA%\StudyDesk`; the database remains available for a later reinstall or manual backup.

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

## Build a release executable

On a Windows build machine, install the development prerequisites in [setup](../03-development/setup.md), then run from the repository root:

```powershell
Set-Location frontend
npm ci
npm run test
npm run build
Set-Location ..
go test ./...
wails build
```

Wails also invokes the configured frontend install and production build during `wails build`; the explicit checks above make failures easier to diagnose before packaging. The generated Windows executable is:

```text
build\bin\studydesk.exe
```

Test the executable on a Windows machine with WebView2 before distributing it. Current packaging limitations:

- No installer, uninstaller, auto-updater, or code-signing configuration is provided.
- Distribute the executable over a trusted channel and communicate its checksum separately if recipients need integrity verification.
- Users are responsible for backing up their local database; it is not included with the executable.

To calculate a SHA-256 checksum for a built executable:

```powershell
Get-FileHash -Algorithm SHA256 -LiteralPath 'build\bin\studydesk.exe'
```
