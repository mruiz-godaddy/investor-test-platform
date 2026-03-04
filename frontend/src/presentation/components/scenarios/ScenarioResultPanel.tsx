import type { ScenarioResult } from '../../../domain/entities/Scenario';

interface Props {
  result: ScenarioResult;
}

export default function ScenarioResultPanel({ result }: Props) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4">
      <h3 className="text-sm font-semibold text-gray-900">Scenario Result: {result.scenario}</h3>
      <p className="mt-1 text-xs text-gray-500">{result.description}</p>

      <div className="mt-4 grid grid-cols-3 gap-4">
        <div>
          <h4 className="text-xs font-semibold text-gray-700">Config</h4>
          <dl className="mt-1 space-y-1 text-xs">
            <div className="flex justify-between">
              <dt className="text-gray-500">autoFinalize</dt>
              <dd>{result.config.autoFinalize ? 'true' : 'false'}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500">transitionDelay</dt>
              <dd>{result.config.statusTransitionDelayMs}ms</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500">finalizerInterval</dt>
              <dd>{result.config.finalizerIntervalMs}ms</dd>
            </div>
          </dl>
        </div>

        <div>
          <h4 className="text-xs font-semibold text-gray-700">Shoppers ({result.shoppers.length})</h4>
          <ul className="mt-1 space-y-1 text-xs text-gray-600">
            {result.shoppers.map((s) => (
              <li key={s.shopperId}>{s.shopperId} (#{s.memberId})</li>
            ))}
          </ul>
        </div>

        <div>
          <h4 className="text-xs font-semibold text-gray-700">Listings ({result.listings.length})</h4>
          <ul className="mt-1 space-y-1 text-xs text-gray-600">
            {result.listings.map((l) => (
              <li key={l.listingId}>{l.domainName} (#{l.listingId})</li>
            ))}
          </ul>
          <p className="mt-2 text-xs text-gray-500">Bids placed: {result.bidsPlaced}</p>
        </div>
      </div>
    </div>
  );
}
