import { useAutoGeneratorViewModel } from '../hooks/useAutoGeneratorViewModel';
import AutoGenConfigForm from '../components/autogen/AutoGenConfigForm';
import AutoGenStatus from '../components/autogen/AutoGenStatus';

export default function AutoGeneratorPage() {
  const { isRunning, start, stop, reset } = useAutoGeneratorViewModel();

  return (
    <div>
      <h2 className="text-xl font-bold text-gray-900 dark:text-white">Auto-Generator</h2>
      <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">Automatically create listings at a configurable interval.</p>
      <div className="mt-4 flex gap-2">
        {isRunning ? (
          <button onClick={stop} className="rounded-md bg-red-600 px-4 py-2 text-sm font-semibold text-white hover:bg-red-500">
            Stop
          </button>
        ) : (
          <button onClick={start} className="rounded-md bg-green-600 px-4 py-2 text-sm font-semibold text-white hover:bg-green-500">
            Start
          </button>
        )}
        <button onClick={reset} disabled={isRunning}
          className="rounded-md bg-gray-100 dark:bg-gray-800 px-4 py-2 text-sm font-semibold text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50">
          Reset Counter
        </button>
      </div>
      <div className="mt-4 grid grid-cols-1 gap-6 lg:grid-cols-2">
        <AutoGenConfigForm />
        <AutoGenStatus />
      </div>
    </div>
  );
}
