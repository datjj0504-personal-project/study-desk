import { useEffect, useState } from "react";
import { fetchAppInfo, toErrorMessage, type AppInfo } from "../services/appInfo";

export type AppInfoState =
  | { status: "loading" }
  | { status: "ready"; info: AppInfo }
  | { status: "error"; message: string };

export function useAppInfo(): AppInfoState {
  const [state, setState] = useState<AppInfoState>({ status: "loading" });

  useEffect(() => {
    let cancelled = false;
    fetchAppInfo()
      .then((info) => !cancelled && setState({ status: "ready", info }))
      .catch((err: unknown) =>
        !cancelled && setState({ status: "error", message: toErrorMessage(err) }),
      );
    return () => {
      cancelled = true;
    };
  }, []);

  return state;
}
