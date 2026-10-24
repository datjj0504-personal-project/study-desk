import { ListStudySessions } from "../../wailsjs/go/main/App";
import { study as generatedStudy } from "../../wailsjs/go/models";

export type SessionKind = "pomodoro" | "stopwatch" | "manual";

export interface StudySession {
  id: number;
  subjectId?: number;
  kind: SessionKind;
  startedAt: string;
  endedAt: string;
  durationSeconds: number;
  note: string;
}

function isSessionKind(kind: string): kind is SessionKind {
  return kind === "pomodoro" || kind === "stopwatch" || kind === "manual";
}

function toStudySession(session: generatedStudy.Session): StudySession {
  if (!isSessionKind(session.kind)) {
    throw new Error(`Unsupported study session kind: ${session.kind}`);
  }
  return {
    id: session.id,
    subjectId: session.subjectId ?? undefined,
    kind: session.kind,
    startedAt: String(session.startedAt),
    endedAt: String(session.endedAt),
    durationSeconds: session.durationSeconds,
    note: session.note,
  };
}

export async function listStudySessions(): Promise<StudySession[]> {
  return (await ListStudySessions()).map(toStudySession);
}
