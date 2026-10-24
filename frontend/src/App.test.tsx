import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import App from "./App";

const { getTheme, getTimerSettings } = vi.hoisted(() => ({
  getTheme: vi.fn(),
  getTimerSettings: vi.fn(),
}));

vi.mock("./services/settings", () => ({
  themeSettings: { get: getTheme },
  timerSettings: { get: getTimerSettings },
}));
vi.mock("./components/Sidebar", () => ({ Sidebar: () => null }));
vi.mock("./components/PomodoroNotificationMonitor", () => ({
  PomodoroNotificationMonitor: () => null,
}));
vi.mock("./pages/DashboardPage", () => ({ DashboardPage: () => null }));
vi.mock("./pages/NotesPage", () => ({ NotesPage: () => null }));
vi.mock("./pages/SessionHistoryPage", () => ({ SessionHistoryPage: () => null }));
vi.mock("./pages/StatisticsPage", () => ({ StatisticsPage: () => null }));
vi.mock("./pages/SettingsPage", () => ({ SettingsPage: () => null }));
vi.mock("./pages/TimerPage", () => ({ TimerPage: () => null }));
vi.mock("./pages/ComingSoonPage", () => ({ ComingSoonPage: () => null }));

afterEach(() => {
  cleanup();
  vi.resetAllMocks();
});

describe("App preference loading", () => {
  it("keeps independent startup errors visible and retries each preference", async () => {
    getTimerSettings
      .mockRejectedValueOnce(new Error("timer storage unavailable"))
      .mockResolvedValueOnce({
        focusMinutes: 25,
        shortBreakMinutes: 5,
        longBreakMinutes: 15,
        longBreakEvery: 4,
        notificationsEnabled: false,
      });
    getTheme
      .mockRejectedValueOnce(new Error("theme storage unavailable"))
      .mockResolvedValueOnce("dark");

    render(<App />);

    await waitFor(() => expect(screen.getAllByRole("alert")).toHaveLength(2));
    expect(screen.getByText(/timer storage unavailable/)).toBeTruthy();
    expect(screen.getByText(/theme storage unavailable/)).toBeTruthy();

    fireEvent.click(screen.getAllByRole("button", { name: "Retry" })[0]);
    await waitFor(() => expect(getTimerSettings).toHaveBeenCalledTimes(2));
    await waitFor(() => expect(screen.getAllByRole("alert")).toHaveLength(1));
    expect(screen.queryByText(/timer storage unavailable/)).toBeNull();
    expect(screen.getByText(/theme storage unavailable/)).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    await waitFor(() => expect(getTheme).toHaveBeenCalledTimes(2));
    await waitFor(() => expect(screen.queryByRole("alert")).toBeNull());
  });
});
