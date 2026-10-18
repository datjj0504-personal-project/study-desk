export type PageId =
  | "dashboard"
  | "notes"
  | "timer"
  | "stopwatch"
  | "statistics"
  | "settings";

export interface NavItem {
  id: PageId;
  label: string;
}

export const NAV_ITEMS: NavItem[] = [
  { id: "dashboard", label: "Dashboard" },
  { id: "notes", label: "Notes" },
  { id: "timer", label: "Focus Timer" },
  { id: "stopwatch", label: "Stopwatch" },
  { id: "statistics", label: "Study Statistics" },
  { id: "settings", label: "Settings" },
];
