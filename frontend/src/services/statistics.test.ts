import { afterEach, describe, expect, it, vi } from "vitest";
import { getStudyStatistics } from "./statistics";

const { get } = vi.hoisted(() => ({ get: vi.fn() }));

vi.mock("../../wailsjs/go/main/App", () => ({ GetStudyStatistics: get }));

afterEach(() => vi.resetAllMocks());

describe("study statistics service", () => {
  it("maps statistics and optional subject IDs from the Wails binding", async () => {
    get.mockResolvedValue({
      todaySeconds: 1800,
      weekSeconds: 3600,
      allTimeSeconds: 7200,
      bySubject: [
        { subjectId: 4, subjectName: "Physics", durationSeconds: 5400 },
        { subjectId: null, subjectName: "Unassigned", durationSeconds: 1800 },
      ],
    });

    await expect(getStudyStatistics()).resolves.toEqual({
      todaySeconds: 1800,
      weekSeconds: 3600,
      allTimeSeconds: 7200,
      bySubject: [
        { subjectId: 4, subjectName: "Physics", durationSeconds: 5400 },
        { subjectName: "Unassigned", durationSeconds: 1800 },
      ],
    });
  });

  it("propagates backend errors", async () => {
    get.mockRejectedValue(new Error("statistics unavailable"));

    await expect(getStudyStatistics()).rejects.toThrow("statistics unavailable");
  });
});
