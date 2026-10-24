import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { SessionHistoryPage } from "./SessionHistoryPage";

const { listStudySessions, listSubjects } = vi.hoisted(() => ({
  listStudySessions: vi.fn(),
  listSubjects: vi.fn(),
}));

vi.mock("../services/sessions", () => ({ listStudySessions }));
vi.mock("../services/notes", () => ({ listSubjects }));

afterEach(() => {
  cleanup();
  vi.resetAllMocks();
});

const sessions = [
  {
    id: 10,
    subjectId: 1,
    kind: "pomodoro" as const,
    startedAt: new Date(2026, 9, 24, 9, 0).toISOString(),
    endedAt: new Date(2026, 9, 24, 9, 25).toISOString(),
    durationSeconds: 1500,
    note: "Read about mechanics",
  },
  {
    id: 9,
    subjectId: null,
    kind: "stopwatch" as const,
    startedAt: new Date(2026, 9, 24, 8, 0).toISOString(),
    endedAt: new Date(2026, 9, 24, 8, 10).toISOString(),
    durationSeconds: 600,
    note: "",
  },
  {
    id: 8,
    subjectId: 2,
    kind: "manual" as const,
    startedAt: new Date(2026, 9, 23, 13, 0).toISOString(),
    endedAt: new Date(2026, 9, 23, 13, 30).toISOString(),
    durationSeconds: 1800,
    note: "",
  },
];

describe("SessionHistoryPage", () => {
  it("lists sessions grouped by local date with subject, duration, and note", async () => {
    listStudySessions.mockResolvedValue(sessions);
    listSubjects.mockResolvedValue([{ id: 1, name: "Physics" }, { id: 2, name: "Biology" }]);
    render(<SessionHistoryPage />);

    expect(await screen.findByText("Read about mechanics")).toBeTruthy();
    const items = screen.getAllByRole("listitem");
    expect(items).toHaveLength(3);
    expect(items[0].textContent).toContain("Pomodoro");
    expect(items[1].textContent).toContain("Stopwatch");
    expect(screen.getByText("25 min")).toBeTruthy();
    expect(screen.getByText("10 min")).toBeTruthy();
    expect(items[0].textContent).toContain("Physics");
    expect(screen.getAllByRole("region")).toHaveLength(2);
  });

  it("filters by date, session type, and subject and can clear filters", async () => {
    listStudySessions.mockResolvedValue(sessions);
    listSubjects.mockResolvedValue([{ id: 1, name: "Physics" }, { id: 2, name: "Biology" }]);
    render(<SessionHistoryPage />);
    await screen.findByText("Read about mechanics");

    fireEvent.change(screen.getByLabelText("From"), { target: { value: "2026-10-24" } });
    fireEvent.change(screen.getByLabelText("Session type"), { target: { value: "pomodoro" } });
    fireEvent.change(screen.getByLabelText("Subject"), { target: { value: "1" } });

    let items = screen.getAllByRole("listitem");
    expect(items).toHaveLength(1);
    expect(items[0].textContent).toContain("Pomodoro");
    expect(items[0].textContent).toContain("Read about mechanics");

    fireEvent.click(screen.getByRole("button", { name: "Clear filters" }));
    items = screen.getAllByRole("listitem");
    expect(items).toHaveLength(3);
    expect(items[1].textContent).toContain("Stopwatch");
    expect(items[2].textContent).toContain("Manual");
  });

  it("shows empty and error states without hiding backend failures", async () => {
    listStudySessions.mockResolvedValueOnce([]);
    listSubjects.mockResolvedValueOnce([]);
    const { unmount } = render(<SessionHistoryPage />);
    expect(await screen.findByText("No study sessions have been recorded yet.")).toBeTruthy();
    unmount();

    listStudySessions.mockRejectedValueOnce(new Error("database unavailable"));
    listSubjects.mockResolvedValueOnce([]);
    render(<SessionHistoryPage />);
    await waitFor(() => expect(screen.getByRole("alert").textContent).toContain("database unavailable"));
  });
});
