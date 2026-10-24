import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { SettingsPage } from "./SettingsPage";

const { update, updateTheme } = vi.hoisted(() => ({
  update: vi.fn(),
  updateTheme: vi.fn(),
}));

vi.mock("../services/settings", () => ({
  timerSettings: { update },
  themeSettings: { update: updateTheme },
}));

const defaultSettings = {
  focusMinutes: 25,
  shortBreakMinutes: 5,
  longBreakMinutes: 15,
  longBreakEvery: 4,
  notificationsEnabled: false,
};
const defaultTheme = "dark" as const;

afterEach(() => {
  cleanup();
  vi.resetAllMocks();
});

describe("SettingsPage", () => {
  it("saves configurable Pomodoro durations and reports the update", async () => {
    const onSettingsChange = vi.fn();
    update.mockResolvedValue(undefined);
    render(
      <SettingsPage
        settings={defaultSettings}
        onSettingsChange={onSettingsChange}
        theme={defaultTheme}
        onThemeChange={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByLabelText("Focus duration (minutes)"), {
      target: { value: "40" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save durations" }));

    const expected = { ...defaultSettings, focusMinutes: 40 };
    await waitFor(() => expect(update).toHaveBeenCalledWith(expected));
    expect(onSettingsChange).toHaveBeenCalledWith(expected);
    expect(await screen.findByText("Settings saved.")).toBeTruthy();
  });

  it("shows backend errors instead of reporting a successful save", async () => {
    update.mockRejectedValue(new Error("timer is running"));
    render(
      <SettingsPage
        settings={defaultSettings}
        onSettingsChange={vi.fn()}
        theme={defaultTheme}
        onThemeChange={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Save durations" }));

    expect((await screen.findByRole("alert")).textContent).toContain("timer is running");
  });

  it("persists the selected theme and applies it through the app callback", async () => {
    const onThemeChange = vi.fn();
    updateTheme.mockResolvedValue(undefined);
    render(
      <SettingsPage
        settings={defaultSettings}
        onSettingsChange={vi.fn()}
        theme={defaultTheme}
        onThemeChange={onThemeChange}
      />,
    );

    fireEvent.change(screen.getByLabelText("Theme"), { target: { value: "light" } });
    fireEvent.click(screen.getByRole("button", { name: "Save theme" }));

    await waitFor(() => expect(updateTheme).toHaveBeenCalledWith("light"));
    expect(onThemeChange).toHaveBeenCalledWith("light");
  });

  it("reports theme persistence errors", async () => {
    updateTheme.mockRejectedValue(new Error("theme unavailable"));
    render(
      <SettingsPage
        settings={defaultSettings}
        onSettingsChange={vi.fn()}
        theme={defaultTheme}
        onThemeChange={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByLabelText("Theme"), { target: { value: "light" } });
    fireEvent.click(screen.getByRole("button", { name: "Save theme" }));

    expect((await screen.findByRole("alert")).textContent).toContain("theme unavailable");
  });
});
