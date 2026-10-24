import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { SettingsPage } from "./SettingsPage";

const { update } = vi.hoisted(() => ({
  update: vi.fn(),
}));

vi.mock("../services/settings", () => ({
  timerSettings: { update },
}));

const defaultSettings = {
  focusMinutes: 25,
  shortBreakMinutes: 5,
  longBreakMinutes: 15,
  longBreakEvery: 4,
  notificationsEnabled: false,
};

afterEach(() => {
  cleanup();
  vi.resetAllMocks();
});

describe("SettingsPage", () => {
  it("saves configurable Pomodoro durations and reports the update", async () => {
    const onSettingsChange = vi.fn();
    update.mockResolvedValue(undefined);
    render(<SettingsPage settings={defaultSettings} onSettingsChange={onSettingsChange} />);

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
    render(<SettingsPage settings={defaultSettings} onSettingsChange={vi.fn()} />);

    fireEvent.click(screen.getByRole("button", { name: "Save durations" }));

    expect((await screen.findByRole("alert")).textContent).toContain("timer is running");
  });
});
