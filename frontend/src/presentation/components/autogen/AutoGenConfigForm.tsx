import { useAutoGeneratorStore } from '../../stores/autoGeneratorStore';

export default function AutoGenConfigForm() {
  const { config, setConfig, isRunning } = useAutoGeneratorStore();

  return (
    <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-4">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Generator Config</h3>
      <div className="mt-3 grid grid-cols-2 gap-3">
        <div>
          <label className="block text-base font-medium text-gray-700 dark:text-gray-200">Interval (ms)</label>
          <input type="number" value={config.intervalMs} disabled={isRunning}
            onChange={(e) => setConfig({ intervalMs: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-lg disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-base font-medium text-gray-700 dark:text-gray-200">Duration (ms)</label>
          <input type="number" value={config.durationMs} disabled={isRunning}
            onChange={(e) => setConfig({ durationMs: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-lg disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-base font-medium text-gray-700 dark:text-gray-200">Domain Pattern</label>
          <input value={config.domainPattern} disabled={isRunning}
            onChange={(e) => setConfig({ domainPattern: e.target.value })}
            className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-lg disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-base font-medium text-gray-700 dark:text-gray-200">Seller Shopper ID</label>
          <input value={config.sellerShopperId} disabled={isRunning}
            onChange={(e) => setConfig({ sellerShopperId: e.target.value })}
            className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-lg disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-base font-medium text-gray-700 dark:text-gray-200">Min Price ($)</label>
          <input type="number" step="0.01" value={config.minPriceUsd} disabled={isRunning}
            onChange={(e) => setConfig({ minPriceUsd: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-lg disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-base font-medium text-gray-700 dark:text-gray-200">Max Price ($)</label>
          <input type="number" step="0.01" value={config.maxPriceUsd} disabled={isRunning}
            onChange={(e) => setConfig({ maxPriceUsd: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-lg disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-base font-medium text-gray-700 dark:text-gray-200">End Time Offset (min)</label>
          <input type="number" value={config.endTimeOffsetMinutes} disabled={isRunning}
            onChange={(e) => setConfig({ endTimeOffsetMinutes: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-lg disabled:bg-gray-100" />
        </div>
        <div className="flex items-center gap-2">
          <input type="checkbox" checked={config.autoExtEnabled} disabled={isRunning}
            onChange={(e) => setConfig({ autoExtEnabled: e.target.checked })}
            className="h-4 w-4 rounded border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white text-indigo-600" />
          <label className="text-base text-gray-700 dark:text-gray-200">Auto-Extend</label>
        </div>
        <div>
          <label className="block text-base font-medium text-gray-700 dark:text-gray-200">Ext Window (s)</label>
          <input type="number" value={config.autoExtWindowSec} disabled={isRunning}
            onChange={(e) => setConfig({ autoExtWindowSec: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-lg disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-base font-medium text-gray-700 dark:text-gray-200">Ext Seconds</label>
          <input type="number" value={config.autoExtSeconds} disabled={isRunning}
            onChange={(e) => setConfig({ autoExtSeconds: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-lg disabled:bg-gray-100" />
        </div>
      </div>
    </div>
  );
}
