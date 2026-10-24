import { afterEach, describe, expect, it, vi } from "vitest";
import { timerSettings } from "./settings";

const { bindings } = vi.hoisted(() => ({
  bindings: {
    GetTimerSettings: vi.fn(),
    UpdateTimerSettings: vi.fn(),
  },
}));

vi.mock("../../wailsjs/go/main/App", () => bindings);

afterEach(() => vi.resetAllMocks());

describe("timer settings service", () => {
  it("loads and updates persisted timer settings", async () => {
    const value = {
      focusMinutes: 45,
      shortBreakMinutes: 7,
      longBreakMinutes: 18,
      longBreakEvery: 3,
      notificationsEnabled: true,
    };
    bindings.GetTimerSettings.mockResolvedValue(value);
    bindings.UpdateTimerSettings.mockResolvedValue(undefined);

    await expect(timerSettings.get()).resolves.toEqual(value);
    await timerSettings.update(value);

    expect(bindings.GetTimerSettings).toHaveBeenCalledOnce();
    expect(bindings.UpdateTimerSettings).toHaveBeenCalledWith(value);
  });

  it("propagates backend errors", async () => {
    bindings.UpdateTimerSettings.mockRejectedValue(new Error("settings unavailable"));

    await expect(timerSettings.update({
      focusMinutes: 25,
      shortBreakMinutes: 5,
      longBreakMinutes: 15,
      longBreakEvery: 4,
      notificationsEnabled: false,
    })).rejects.toThrow("settings unavailable");
  });
});
