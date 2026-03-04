import type { ScenarioMetadata } from '../../../domain/entities/Scenario';

interface Props {
  scenario: ScenarioMetadata;
  onLoad: () => void;
  isLoading: boolean;
}

export default function ScenarioCard({ scenario, onLoad, isLoading }: Props) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4">
      <h3 className="text-sm font-semibold text-gray-900">{scenario.name}</h3>
      <p className="mt-1 text-xs text-gray-500">{scenario.description}</p>
      <div className="mt-2 flex flex-wrap gap-1">
        {scenario.tags.map((tag) => (
          <span key={tag} className="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600">{tag}</span>
        ))}
      </div>
      <button
        onClick={onLoad}
        disabled={isLoading}
        className="mt-3 w-full rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
      >
        {isLoading ? 'Loading...' : 'Load'}
      </button>
    </div>
  );
}
