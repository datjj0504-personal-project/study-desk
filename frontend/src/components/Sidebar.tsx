import { NAV_ITEMS, type PageId } from "../types/navigation";

interface SidebarProps {
  active: PageId;
  onSelect: (page: PageId) => void;
}

export function Sidebar({ active, onSelect }: SidebarProps) {
  return (
    <aside className="flex w-60 shrink-0 flex-col border-r border-edge bg-panel/60 p-4">
      <div className="mb-6 px-2 text-xl font-semibold tracking-tight">
        Study<span className="text-accent">Desk</span>
      </div>
      <nav className="flex flex-col gap-1" aria-label="Main navigation">
        {NAV_ITEMS.map((item) => {
          const isActive = item.id === active;
          return (
            <button
              key={item.id}
              type="button"
              aria-current={isActive ? "page" : undefined}
              onClick={() => onSelect(item.id)}
              className={`rounded-md px-3 py-2 text-left text-sm transition-colors ${
                isActive
                  ? "bg-accent text-white"
                  : "text-slate-300 hover:bg-edge/60"
              }`}
            >
              {item.label}
            </button>
          );
        })}
      </nav>
    </aside>
  );
}
