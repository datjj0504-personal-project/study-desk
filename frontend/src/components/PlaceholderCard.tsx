interface PlaceholderCardProps {
  title: string;
  message: string;
}

export function PlaceholderCard({ title, message }: PlaceholderCardProps) {
  return (
    <section className="rounded-xl border border-edge bg-panel p-5">
      <h2 className="mb-3 text-sm font-medium uppercase tracking-wide text-slate-400">
        {title}
      </h2>
      <div className="flex min-h-32 items-center justify-center rounded-lg border border-dashed border-edge px-4 text-center text-sm text-slate-400">
        {message}
      </div>
    </section>
  );
}
