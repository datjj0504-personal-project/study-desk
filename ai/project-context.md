# Project Context

## Purpose and Users

StudyDesk is a local-first Windows 10/11 desktop app for students and self-learners to manage notes, run focus sessions, and see study statistics.

## Core Features

Notes, Pomodoro timer, stopwatch, study session tracking, basic statistics, local persistence.

## Technology Stack

Go, Wails, React + TypeScript, Tailwind CSS, SQLite.

## Status

Desktop shell exists (hand-written Wails project, since the Wails CLI is not installed). No features are implemented. See [PROJECT_STATUS.md](PROJECT_STATUS.md) for details and the next task.

## Constraints

- Works offline, no separate backend server, no Docker.
- Windows 10/11 only.
- No secrets or real user data in the repo.
- Do not hand-write generated Wails output (`frontend/wailsjs`, `build/bin`).

## Key Documents

- [Functional requirements](../docs/01-requirements/functional-requirements.md)
- [Non-functional requirements](../docs/01-requirements/non-functional-requirements.md)
- [Architecture](../docs/02-design/architecture.md)
- [Database design (proposal)](../docs/02-design/database-design.md)
- [Setup](../docs/03-development/setup.md)
