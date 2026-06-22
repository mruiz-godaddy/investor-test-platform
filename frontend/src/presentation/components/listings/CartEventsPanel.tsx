import { useCartEventsViewModel } from '../../hooks/useCartEventsViewModel';
import { formatMicrosUsd, formatDateTime } from '../../../lib/formatters';

export default function CartEventsPanel() {
  const { events, clear, isClearing } = useCartEventsViewModel();

  return (
    <div className="mt-8 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900">
      <div className="flex items-center justify-between border-b border-gray-200 dark:border-gray-700 px-4 py-3">
        <div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Captured ITC codes</h3>
          <p className="text-xs text-gray-500 dark:text-gray-400">
            X-Itc-Code sent by the app on each add-to-cart. Verifies the itc string per inventory type.
          </p>
        </div>
        <button
          type="button"
          onClick={() => clear()}
          disabled={isClearing || events.length === 0}
          className="rounded-md bg-gray-100 dark:bg-gray-800 px-3 py-1.5 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50"
        >
          {isClearing ? 'Clearing...' : 'Clear'}
        </button>
      </div>

      {events.length === 0 ? (
        <p className="px-4 py-6 text-sm text-gray-500 dark:text-gray-400">
          No add-to-cart events captured yet. Add a BIN/closeout/OCO domain to the cart from the app.
        </p>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-800 text-left text-xs uppercase text-gray-500 dark:text-gray-400">
              <tr>
                <th className="px-4 py-2">Domain</th>
                <th className="px-4 py-2">Inv. type</th>
                <th className="px-4 py-2">ITC string</th>
                <th className="px-4 py-2">Area</th>
                <th className="px-4 py-2">Price</th>
                <th className="px-4 py-2">Time</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
              {events.map((e) => (
                <tr key={e.eventId} className="text-gray-800 dark:text-gray-100">
                  <td className="px-4 py-2 font-medium">{e.domainName}</td>
                  <td className="px-4 py-2 tabular-nums">{e.inventoryType}</td>
                  <td className="px-4 py-2">
                    <span className="rounded bg-indigo-100 dark:bg-indigo-900 px-2 py-0.5 font-mono text-xs text-indigo-800 dark:text-indigo-200">
                      {e.itcInventory || '(none)'}
                    </span>
                  </td>
                  <td className="px-4 py-2">{e.area}</td>
                  <td className="px-4 py-2 tabular-nums">{formatMicrosUsd(e.requestPrice)}</td>
                  <td className="px-4 py-2 whitespace-nowrap text-gray-500 dark:text-gray-400">{formatDateTime(e.createdAt)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
