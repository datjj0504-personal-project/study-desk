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

The foundation, local storage, notes, timer/stopwatch, study tracking, preferences, and Windows executable milestones are implemented. See [project status](../../ai/PROJECT_STATUS.md) for task-level progress.

The Windows executable and end-user backup/restore instructions are available. An Inno Setup configuration and release ZIP script are present, but installer compilation and clean-machine install/uninstall verification are still pending. There is no auto-updater or code signing. See [Windows installation, backup, and release](../05-release/install-backup.md).
