import { PlaceholderCard } from "../components/PlaceholderCard";

export function DashboardPage() {
  return (
    <div>
      <h1 className="text-2xl font-semibold">Welcome to StudyDesk</h1>
      <p className="mt-1 text-slate-400">
        Your personal study companion. Features below are not implemented yet.
      </p>
      <div className="mt-6 grid grid-cols-1 gap-4 xl:grid-cols-3">
        <PlaceholderCard
          title="Recent Notes"
          message="Notes are not available yet."
        />
        <PlaceholderCard
          title="Focus Timer"
          message="The focus timer is not available yet."
        />
        <PlaceholderCard
          title="Study Statistics"
          message="No statistics yet. Statistics will appear once study sessions can be recorded."
        />
      </div>
    </div>
  );
}
