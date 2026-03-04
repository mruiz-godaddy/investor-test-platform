import { useAutoGeneratorStore } from '../../stores/autoGeneratorStore';

export default function AutoGenStatus() {
  const { isRunning, generatedCount, recentDomains } = useAutoGeneratorStore();

  return (
    <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-4">
      <div className="flex items-center gap-3">
        <div className={`h-3 w-3 rounded-full ${isRunning ? 'bg-green-500 animate-pulse' : 'bg-gray-300'}`} />
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
          {isRunning ? 'Running' : 'Stopped'}
        </h3>
        <span className="text-lg text-gray-500 dark:text-gray-400">Generated: {generatedCount}</span>
      </div>
      {recentDomains.length > 0 && (
        <div className="mt-3 max-h-48 overflow-y-auto">
          <h4 className="text-base font-medium text-gray-700 dark:text-gray-200">Recent Domains</h4>
          <ul className="mt-1 space-y-0.5 font-mono text-base text-gray-500 dark:text-gray-400">
            {recentDomains.map((domain, i) => (
              <li key={i}>{domain}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
