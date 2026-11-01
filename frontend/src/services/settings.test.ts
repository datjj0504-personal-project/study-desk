import { afterEach, describe, expect, it, vi } from "vitest";
import { themeSettings, timerSettings } from "./settings";

const { bindings } = vi.hoisted(() => ({
  bindings: {
    GetTimerSettings: vi.fn(),
    UpdateTimerSettings: vi.fn(),
    GetTheme: vi.fn(),
    UpdateTheme: vi.fn(),
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

  it("loads and saves the persisted theme preference", async () => {
    bindings.GetTheme.mockResolvedValue("light");
    bindings.UpdateTheme.mockResolvedValue(undefined);

    await expect(themeSettings.get()).resolves.toBe("light");
    await themeSettings.update("system");

    expect(bindings.GetTheme).toHaveBeenCalledOnce();
    expect(bindings.UpdateTheme).toHaveBeenCalledWith("system");
  });
});
