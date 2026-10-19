import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { BackendStatus } from "./BackendStatus";

const { mockFetchAppInfo } = vi.hoisted(() => ({
  mockFetchAppInfo: vi.fn<() => Promise<{ name: string; version: string }>>(),
}));

vi.mock("../services/appInfo", () => ({
  fetchAppInfo: mockFetchAppInfo,
  toErrorMessage: (error: unknown) =>
    error instanceof Error ? error.message : String(error),
}));

afterEach(() => {
  cleanup();
  vi.resetAllMocks();
});

describe("BackendStatus", () => {
  it("shows application details after the backend connects", async () => {
    mockFetchAppInfo.mockResolvedValue({ name: "StudyDesk", version: "0.1.0" });

    render(<BackendStatus />);

    expect(await screen.findByText("StudyDesk v0.1.0 (backend connected)")).toBeTruthy();
  });

  it("shows an alert when the backend request fails", async () => {
    mockFetchAppInfo.mockRejectedValue(new Error("connection failed"));

    render(<BackendStatus />);

    expect((await screen.findByRole("alert")).textContent).toContain(
      "Backend unavailable: connection failed",
    );
  });
});
