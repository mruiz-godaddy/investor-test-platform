import { useAutoGeneratorStore } from '../../stores/autoGeneratorStore';

export default function AutoGenStatus() {
  const { isRunning, generatedCount, recentDomains } = useAutoGeneratorStore();

  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4">
      <div className="flex items-center gap-3">
        <div className={`h-3 w-3 rounded-full ${isRunning ? 'bg-green-500 animate-pulse' : 'bg-gray-300'}`} />
        <h3 className="text-sm font-semibold text-gray-900">
          {isRunning ? 'Running' : 'Stopped'}
        </h3>
        <span className="text-sm text-gray-500">Generated: {generatedCount}</span>
      </div>
      {recentDomains.length > 0 && (
        <div className="mt-3 max-h-48 overflow-y-auto">
          <h4 className="text-xs font-medium text-gray-700">Recent Domains</h4>
          <ul className="mt-1 space-y-0.5 font-mono text-xs text-gray-500">
            {recentDomains.map((domain, i) => (
              <li key={i}>{domain}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
