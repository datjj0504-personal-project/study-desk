import { PlaceholderCard } from "../components/PlaceholderCard";

interface ComingSoonPageProps {
  title: string;
}

export function ComingSoonPage({ title }: ComingSoonPageProps) {
  return (
    <div>
      <h1 className="text-2xl font-semibold">{title}</h1>
      <div className="mt-6 max-w-xl">
        <PlaceholderCard title={title} message="Not implemented yet." />
      </div>
    </div>
  );
}
