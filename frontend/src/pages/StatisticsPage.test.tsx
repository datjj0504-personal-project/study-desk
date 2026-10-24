import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { StatisticsPage } from "./StatisticsPage";

const { getStudyStatistics } = vi.hoisted(() => ({ getStudyStatistics: vi.fn() }));

vi.mock("../services/statistics", () => ({ getStudyStatistics }));

afterEach(() => {
  cleanup();
  vi.resetAllMocks();
});

describe("StatisticsPage", () => {
  it("shows period totals and per-subject duration", async () => {
    getStudyStatistics.mockResolvedValue({
      todaySeconds: 90,
      weekSeconds: 3600,
      allTimeSeconds: 7200,
      bySubject: [
        { subjectId: 1, subjectName: "Physics", durationSeconds: 5400 },
        { subjectName: "Unassigned", durationSeconds: 1800 },
      ],
    });
    render(<StatisticsPage />);

    expect(await screen.findByText("1 min 30 sec")).toBeTruthy();
    expect(screen.getByText("1 hr 0 min")).toBeTruthy();
    expect(screen.getByText("2 hr 0 min")).toBeTruthy();
    expect(screen.getByText("Physics")).toBeTruthy();
    expect(screen.getByText("Unassigned")).toBeTruthy();
  });

  it("shows an empty state and retries when refreshed", async () => {
    getStudyStatistics.mockResolvedValueOnce({
      todaySeconds: 0,
      weekSeconds: 0,
      allTimeSeconds: 0,
      bySubject: [],
    }).mockRejectedValueOnce(new Error("database unavailable"));
    render(<StatisticsPage />);
    expect(await screen.findByText("No completed study sessions yet.")).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));

    await waitFor(() => expect(screen.getByRole("alert").textContent).toContain("database unavailable"));
    expect(getStudyStatistics).toHaveBeenCalledTimes(2);
  });
});
