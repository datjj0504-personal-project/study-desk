# StudyDesk

A local-first Windows desktop study companion built with Go, Wails, React, TypeScript, Tailwind CSS and SQLite.

> Status: functional Windows desktop build available. Current release limitations and project progress: [ai/PROJECT_STATUS.md](ai/PROJECT_STATUS.md).

## Features

- Notes with subjects, search, and automatic saving
- Configurable Pomodoro focus and break timer with completion alerts
- Stopwatch
- Study session history and statistics
- Locally persisted notes, sessions, timer preferences, and appearance settings

Details: [functional requirements](docs/01-requirements/functional-requirements.md).

## Technology Stack

| Layer | Technology |
| --- | --- |
| Backend | Go |
| Desktop shell | Wails |
| Frontend | React + TypeScript |
| Styling | Tailwind CSS |
| Database | SQLite (local file) |

## Running the Windows build

The application requires Windows 10/11 and the Microsoft Edge WebView2 Runtime. No Go or Node.js installation is needed to run the prebuilt executable.

Download or build `studydesk.exe`, copy it to a folder writable by your Windows user, and launch it. StudyDesk currently ships as an executable, not an installer; see [installation, backup, and release instructions](docs/05-release/install-backup.md).

## Development prerequisites

- Windows 10/11 with WebView2 Runtime
- Go (1.21+)
- Node.js (LTS) and npm
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

Verify the toolchain with `wails doctor`. Full instructions: [development setup](docs/03-development/setup.md).

## Development

```powershell
cd frontend; npm install; cd ..
wails dev
```

Workflow and conventions: [development workflow](docs/03-development/development-workflow.md), [coding conventions](docs/03-development/coding-conventions.md).

## Testing

```powershell
go test ./...
cd frontend; npm run test; npm run build
```

Frontend tests use Vitest and React Testing Library for service utilities and UI components.

Strategy: [testing strategy](docs/04-testing/testing-strategy.md).

## Build

```powershell
wails build
```

The Windows executable is written to `build/bin/studydesk.exe`. Build and release details: [installation, backup, and release instructions](docs/05-release/install-backup.md).

## Documentation

- [Project overview](docs/00-project-context/project-overview.md)
- [Architecture](docs/02-design/architecture.md)
- [Folder structure](docs/02-design/folder-structure.md)
- [Database design (proposal)](docs/02-design/database-design.md)
- [AI agent context](ai/README.md)
- [Project status](ai/PROJECT_STATUS.md)
