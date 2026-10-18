# Project Overview

## Vision

StudyDesk is a personal, offline study companion: one small desktop app where a student keeps notes, runs focus sessions, and sees how much they actually study.

## Target Users

Students and self-learners on Windows 10/11 who want a simple, private tool without accounts or cloud services.

## Scope

In scope (v1): notes, Pomodoro timer, stopwatch, study session tracking, basic statistics, local persistence.

Out of scope (v1): cloud sync, accounts, collaboration, mobile apps, a separate backend server, Docker.

See [functional requirements](../01-requirements/functional-requirements.md) and [non-functional requirements](../01-requirements/non-functional-requirements.md). Architecture: [architecture](../02-design/architecture.md).

## Milestones

1. **M0 Foundation**: repository structure, docs, Wails project generated (current).
2. **M1 Storage**: SQLite setup and migrations, settings.
3. **M2 Notes**: CRUD and search.
4. **M3 Timer and stopwatch**: Pomodoro and stopwatch.
5. **M4 Sessions**: record sessions from timers and manual entry.
6. **M5 Statistics**: daily/weekly totals.
7. **M6 Release**: Windows build and installer.
