import type { ScenarioResult } from '../../../domain/entities/Scenario';

interface Props {
  result: ScenarioResult;
}

export default function ScenarioResultPanel({ result }: Props) {
  return (
    <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-4">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Scenario Result: {result.scenario}</h3>
      <p className="mt-1 text-base text-gray-500 dark:text-gray-400">{result.description}</p>

      <div className="mt-4 grid grid-cols-3 gap-4">
        <div>
          <h4 className="text-base font-semibold text-gray-700 dark:text-gray-200">Config</h4>
          <dl className="mt-1 space-y-1 text-base">
            <div className="flex justify-between">
              <dt className="text-gray-500 dark:text-gray-400">autoFinalize</dt>
              <dd>{result.config.autoFinalize ? 'true' : 'false'}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500 dark:text-gray-400">transitionDelay</dt>
              <dd>{result.config.statusTransitionDelayMs}ms</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500 dark:text-gray-400">finalizerInterval</dt>
              <dd>{result.config.finalizerIntervalMs}ms</dd>
            </div>
          </dl>
        </div>

        <div>
          <h4 className="text-base font-semibold text-gray-700 dark:text-gray-200">Shoppers ({result.shoppers.length})</h4>
          <ul className="mt-1 space-y-1 text-base text-gray-600 dark:text-gray-300">
            {result.shoppers.map((s) => (
              <li key={s.shopperId}>{s.shopperId} (#{s.memberId})</li>
            ))}
          </ul>
        </div>

        <div>
          <h4 className="text-base font-semibold text-gray-700 dark:text-gray-200">Listings ({result.listings.length})</h4>
          <ul className="mt-1 space-y-1 text-base text-gray-600 dark:text-gray-300">
            {result.listings.map((l) => (
              <li key={l.listingId}>{l.domainName} (#{l.listingId})</li>
            ))}
          </ul>
          <p className="mt-2 text-base text-gray-500 dark:text-gray-400">Bids placed: {result.bidsPlaced}</p>
        </div>
      </div>
    </div>
  );
}
