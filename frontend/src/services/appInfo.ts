import { GetAppInfo } from "../../wailsjs/go/main/App";

export { toErrorMessage } from "../utils/errors";

export interface AppInfo {
  name: string;
  version: string;
}

export async function fetchAppInfo(): Promise<AppInfo> {
  // Outside the Wails window (e.g. plain browser) the binding is undefined.
  if (!window.go?.main?.App) {
    throw new Error("Not running inside the StudyDesk desktop window.");
  }
  const info = await GetAppInfo();
  return { name: info.name, version: info.version };
}
