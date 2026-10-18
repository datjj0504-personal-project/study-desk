import { useAppInfo } from "../hooks/useAppInfo";

export function BackendStatus() {
  const state = useAppInfo();

  if (state.status === "loading") {
    return <p className="text-xs text-slate-500">Connecting to backend...</p>;
  }
  if (state.status === "error") {
    return (
      <p role="alert" className="text-xs text-red-400">
        Backend unavailable: {state.message}
      </p>
    );
  }
  return (
    <p className="text-xs text-slate-500">
      {state.info.name} v{state.info.version} (backend connected)
    </p>
  );
}
