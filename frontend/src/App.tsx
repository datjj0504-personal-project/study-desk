import { useState } from "react";
import { Sidebar } from "./components/Sidebar";
import { ComingSoonPage } from "./pages/ComingSoonPage";
import { DashboardPage } from "./pages/DashboardPage";
import { NAV_ITEMS, type PageId } from "./types/navigation";

function App() {
  const [page, setPage] = useState<PageId>("dashboard");
  const current = NAV_ITEMS.find((item) => item.id === page);

  return (
    <div className="flex h-full">
      <Sidebar active={page} onSelect={setPage} />
      <main className="flex-1 overflow-y-auto p-8">
        {page === "dashboard" ? (
          <DashboardPage />
        ) : (
          <ComingSoonPage title={current?.label ?? "Page"} />
        )}
      </main>
    </div>
  );
}

export default App;
