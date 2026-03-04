import { useAutoGeneratorStore } from '../../stores/autoGeneratorStore';

export default function AutoGenConfigForm() {
  const { config, setConfig, isRunning } = useAutoGeneratorStore();

  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4">
      <h3 className="text-sm font-semibold text-gray-900">Generator Config</h3>
      <div className="mt-3 grid grid-cols-2 gap-3">
        <div>
          <label className="block text-xs font-medium text-gray-700">Interval (ms)</label>
          <input type="number" value={config.intervalMs} disabled={isRunning}
            onChange={(e) => setConfig({ intervalMs: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">Duration (ms)</label>
          <input type="number" value={config.durationMs} disabled={isRunning}
            onChange={(e) => setConfig({ durationMs: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">Domain Pattern</label>
          <input value={config.domainPattern} disabled={isRunning}
            onChange={(e) => setConfig({ domainPattern: e.target.value })}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">Seller Shopper ID</label>
          <input value={config.sellerShopperId} disabled={isRunning}
            onChange={(e) => setConfig({ sellerShopperId: e.target.value })}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">Min Price ($)</label>
          <input type="number" step="0.01" value={config.minPriceUsd} disabled={isRunning}
            onChange={(e) => setConfig({ minPriceUsd: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">Max Price ($)</label>
          <input type="number" step="0.01" value={config.maxPriceUsd} disabled={isRunning}
            onChange={(e) => setConfig({ maxPriceUsd: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">End Time Offset (min)</label>
          <input type="number" value={config.endTimeOffsetMinutes} disabled={isRunning}
            onChange={(e) => setConfig({ endTimeOffsetMinutes: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">Reserve Price ($)</label>
          <input type="number" step="0.01" value={config.reservePriceUsd} disabled={isRunning}
            onChange={(e) => setConfig({ reservePriceUsd: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm disabled:bg-gray-100" />
        </div>
        <div className="flex items-center gap-2">
          <input type="checkbox" checked={config.autoExtEnabled} disabled={isRunning}
            onChange={(e) => setConfig({ autoExtEnabled: e.target.checked })}
            className="h-4 w-4 rounded border-gray-300 text-indigo-600" />
          <label className="text-xs text-gray-700">Auto-Extend</label>
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">Ext Window (s)</label>
          <input type="number" value={config.autoExtWindowSec} disabled={isRunning}
            onChange={(e) => setConfig({ autoExtWindowSec: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm disabled:bg-gray-100" />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">Ext Seconds</label>
          <input type="number" value={config.autoExtSeconds} disabled={isRunning}
            onChange={(e) => setConfig({ autoExtSeconds: Number(e.target.value) })}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm disabled:bg-gray-100" />
        </div>
      </div>
    </div>
  );
}
