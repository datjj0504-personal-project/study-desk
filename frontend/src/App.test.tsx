import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import App from "./App";

const { getTheme, getTimerSettings, updateTheme, updateTimerSettings } = vi.hoisted(() => ({
  getTheme: vi.fn(),
  getTimerSettings: vi.fn(),
  updateTheme: vi.fn(),
  updateTimerSettings: vi.fn(),
}));

vi.mock("./services/settings", () => ({
  themeSettings: { get: getTheme, update: updateTheme },
  timerSettings: { get: getTimerSettings, update: updateTimerSettings },
}));
vi.mock("./components/PomodoroNotificationMonitor", () => ({
  PomodoroNotificationMonitor: () => null,
}));
vi.mock("./pages/DashboardPage", () => ({ DashboardPage: () => null }));
vi.mock("./pages/NotesPage", () => ({ NotesPage: () => null }));
vi.mock("./pages/SessionHistoryPage", () => ({ SessionHistoryPage: () => null }));
vi.mock("./pages/StatisticsPage", () => ({ StatisticsPage: () => null }));
vi.mock("./pages/TimerPage", () => ({ TimerPage: () => null }));
vi.mock("./pages/ComingSoonPage", () => ({ ComingSoonPage: () => null }));

afterEach(() => {
  cleanup();
  delete document.documentElement.dataset.theme;
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

  it("loads the persisted theme and applies a saved theme from Settings", async () => {
    getTimerSettings.mockResolvedValue({
      focusMinutes: 25,
      shortBreakMinutes: 5,
      longBreakMinutes: 15,
      longBreakEvery: 4,
      notificationsEnabled: false,
    });
    getTheme.mockResolvedValue("dark");
    updateTheme.mockResolvedValue(undefined);
    updateTimerSettings.mockResolvedValue(undefined);

    render(<App />);

    await waitFor(() => expect(document.documentElement.dataset.theme).toBe("dark"));
    fireEvent.click(screen.getByRole("button", { name: "Settings" }));
    expect(await screen.findByRole("heading", { name: "Settings" })).toBeTruthy();

    fireEvent.change(screen.getByLabelText("Theme"), { target: { value: "light" } });
    fireEvent.click(screen.getByRole("button", { name: "Save theme" }));

    await waitFor(() => expect(updateTheme).toHaveBeenCalledWith("light"));
    await waitFor(() => expect(document.documentElement.dataset.theme).toBe("light"));
    expect(await screen.findByText("Settings saved.")).toBeTruthy();
  });
});
