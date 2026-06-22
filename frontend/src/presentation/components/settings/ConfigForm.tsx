import { useState, useEffect } from 'react';
import type { ConfigSnapshot, ConfigUpdate } from '../../../domain/entities/ServerConfig';

interface Props {
  config: ConfigSnapshot | undefined;
  onUpdate: (update: ConfigUpdate) => void;
  isPending?: boolean;
}

export default function ConfigForm({ config, onUpdate, isPending }: Props) {
  const [autoFinalize, setAutoFinalize] = useState(true);
  const [transitionDelay, setTransitionDelay] = useState(0);
  const [finalizerInterval, setFinalizerInterval] = useState(1000);
  const [autoExtWindowSec, setAutoExtWindowSec] = useState(60);
  const [autoExtSeconds, setAutoExtSeconds] = useState(300);
  const [includeBin, setIncludeBin] = useState(false);

  useEffect(() => {
    if (config) {
      setAutoFinalize(config.autoFinalize);
      setTransitionDelay(config.statusTransitionDelayMs);
      setFinalizerInterval(config.finalizerIntervalMs);
      setAutoExtWindowSec(config.autoExtWindowSec);
      setAutoExtSeconds(config.autoExtSeconds);
      setIncludeBin(config.includeBin);
    }
  }, [config]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const update: ConfigUpdate = {};
    if (config && autoFinalize !== config.autoFinalize) update.autoFinalize = autoFinalize;
    if (config && transitionDelay !== config.statusTransitionDelayMs) update.statusTransitionDelayMs = transitionDelay;
    if (config && finalizerInterval !== config.finalizerIntervalMs) update.finalizerIntervalMs = finalizerInterval;
    if (config && autoExtWindowSec !== config.autoExtWindowSec) update.autoExtWindowSec = autoExtWindowSec;
    if (config && autoExtSeconds !== config.autoExtSeconds) update.autoExtSeconds = autoExtSeconds;
    if (config && includeBin !== config.includeBin) update.includeBin = includeBin;
    onUpdate(update);
  };

  return (
    <form onSubmit={handleSubmit} className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-4">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Server Config</h3>
      <div className="mt-3 space-y-3">
        <div className="flex items-center gap-3">
          <input type="checkbox" checked={autoFinalize} onChange={(e) => setAutoFinalize(e.target.checked)}
            className="h-4 w-4 rounded border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white text-indigo-600" />
          <label className="text-lg text-gray-700 dark:text-gray-200">Auto-Finalize</label>
        </div>
        <div className="flex items-center gap-3">
          <input type="checkbox" checked={includeBin} onChange={(e) => setIncludeBin(e.target.checked)}
            className="h-4 w-4 rounded border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white text-indigo-600" />
          <label className="text-lg text-gray-700 dark:text-gray-200">Include BIN domains in app feeds</label>
        </div>
        <div>
          <label className="block text-base font-medium text-gray-700 dark:text-gray-200">Status Transition Delay (ms)</label>
          <input type="number" value={transitionDelay} onChange={(e) => setTransitionDelay(Number(e.target.value))}
            className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-lg" />
        </div>
        <div>
          <label className="block text-base font-medium text-gray-700 dark:text-gray-200">Finalizer Interval (ms)</label>
          <input type="number" value={finalizerInterval} onChange={(e) => setFinalizerInterval(Number(e.target.value))}
            className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-lg" />
        </div>
        <div>
          <label className="block text-base font-medium text-gray-700 dark:text-gray-200">Auto-Extension Window (sec)</label>
          <input type="number" value={autoExtWindowSec} onChange={(e) => setAutoExtWindowSec(Number(e.target.value))}
            className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-lg" />
        </div>
        <div>
          <label className="block text-base font-medium text-gray-700 dark:text-gray-200">Auto-Extension Duration (sec)</label>
          <input type="number" value={autoExtSeconds} onChange={(e) => setAutoExtSeconds(Number(e.target.value))}
            className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-lg" />
        </div>
      </div>
      <button type="submit" disabled={isPending} className="mt-4 rounded-md bg-indigo-600 px-4 py-2 text-lg font-semibold text-white hover:bg-indigo-500 disabled:opacity-50">
        {isPending ? 'Saving...' : 'Update Config'}
      </button>
    </form>
  );
}
