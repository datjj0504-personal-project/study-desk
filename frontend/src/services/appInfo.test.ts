import { describe, expect, it } from "vitest";

import { toErrorMessage } from "../utils/errors";

describe("toErrorMessage", () => {
  it("returns messages from errors and strings", () => {
    expect(toErrorMessage(new Error("backend failed"))).toBe("backend failed");
    expect(toErrorMessage("request failed")).toBe("request failed");
  });

  it("returns a fallback for other thrown values", () => {
    expect(toErrorMessage({ reason: "unknown" })).toBe("Unexpected error");
  });
});
