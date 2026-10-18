# Setup

## Prerequisites

- Windows 10/11 (64-bit) with the WebView2 runtime
- Go 1.21+ (`go version`)
- Node.js LTS and npm (`node --version`)
- Wails CLI v2

Install the Wails CLI (needs explicit approval on your machine):

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor
```

## Project setup

The Wails project is already in the repo (written by hand, without the CLI). After installing the CLI:

```powershell
cd frontend; npm install; cd ..
go mod tidy
wails dev   # generates frontend/wailsjs and any missing build assets
```

If `wails dev` reports missing `build/` assets (icon, manifest), generate them with `wails init` in a temp directory and copy only `build/`. See [PROJECT_STATUS.md](../../ai/PROJECT_STATUS.md).

## Run

```powershell
wails dev
```

See [development workflow](development-workflow.md).
