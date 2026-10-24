import { GetStudyStatistics } from "../../wailsjs/go/main/App";

export interface SubjectStatistics {
  subjectId?: number;
  subjectName: string;
  durationSeconds: number;
}

export interface StudyStatistics {
  todaySeconds: number;
  weekSeconds: number;
  allTimeSeconds: number;
  bySubject: SubjectStatistics[];
}

export async function getStudyStatistics(): Promise<StudyStatistics> {
  const statistics = await GetStudyStatistics();
  return {
    todaySeconds: statistics.todaySeconds,
    weekSeconds: statistics.weekSeconds,
    allTimeSeconds: statistics.allTimeSeconds,
    bySubject: statistics.bySubject.map((subject) => ({
      subjectId: subject.subjectId ?? undefined,
      subjectName: subject.subjectName,
      durationSeconds: subject.durationSeconds,
    })),
  };
}
