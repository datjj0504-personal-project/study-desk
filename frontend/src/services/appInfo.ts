import { GetAppInfo } from "../../wailsjs/go/main/App";

export interface AppInfo {
  name: string;
  version: string;
}

/** Converts any thrown value from the Wails bridge into a user-readable message. */
export function toErrorMessage(err: unknown): string {
  if (err instanceof Error) return err.message;
  if (typeof err === "string") return err;
  return "Unexpected error";
}

export async function fetchAppInfo(): Promise<AppInfo> {
  // Outside the Wails window (e.g. plain browser) the binding is undefined.
  if (!window.go?.main?.App) {
    throw new Error("Not running inside the StudyDesk desktop window.");
  }
  const info = await GetAppInfo();
  return { name: info.name, version: info.version };
}
