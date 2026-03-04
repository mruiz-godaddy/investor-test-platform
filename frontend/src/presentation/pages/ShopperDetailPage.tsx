import { useParams, Link } from 'react-router-dom';
import { useShopperDetailViewModel } from '../hooks/useShopperDetailViewModel';
import PriceDisplay from '../components/shared/PriceDisplay';
import LoadingSpinner from '../components/shared/LoadingSpinner';
import { formatDateTime } from '../../lib/formatters';
import type { ShopperBid } from '../../domain/entities/ShopperBid';

function BidTable({ title, bids, colorClass }: { title: string; bids: ShopperBid[]; colorClass: string }) {
  if (bids.length === 0) {
    return (
      <div className="mt-4">
        <h3 className="text-sm font-semibold text-gray-900 dark:text-white">{title} <span className={`ml-1 inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${colorClass}`}>0</span></h3>
        <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">No bids.</p>
      </div>
    );
  }

  return (
    <div className="mt-4">
      <h3 className="text-sm font-semibold text-gray-900 dark:text-white">{title} <span className={`ml-1 inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${colorClass}`}>{bids.length}</span></h3>
      <div className="mt-2 overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-950">
            <tr>
              {['Domain', 'Amount', 'Type', 'Status', 'Created', ''].map((h) => (
                <th key={h} className="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-700 bg-white dark:bg-gray-900">
            {bids.map((bid) => (
              <tr key={bid.bidId}>
                <td className="px-3 py-2 text-xs text-gray-900 dark:text-white font-medium">{bid.domainName || '-'}</td>
                <td className="px-3 py-2 text-xs"><PriceDisplay micros={bid.bidAmountUsd} /></td>
                <td className="px-3 py-2">
                  <span className={`inline-flex rounded px-1.5 py-0.5 text-xs font-medium ${
                    bid.bidType === 'PROXY' ? 'bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200' : 'bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-100'
                  }`}>{bid.bidType}</span>
                </td>
                <td className="px-3 py-2">
                  {(() => {
                    let label: string = bid.bidStatus;
                    let cls = 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200';
                    if (bid.bidStatus === 'CANCELLED') {
                      label = 'CANCELLED';
                      cls = 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200';
                    } else if (bid.listingStatus === 'SOLD' && bid.highestBidderShopper === bid.shopperId && bid.isHighBid) {
                      label = 'WON';
                      cls = 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200';
                    } else if (bid.listingStatus === 'SOLD' || bid.listingStatus === 'CLOSED') {
                      label = 'LOST';
                      cls = 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200';
                    }
                    return <span className={`inline-flex rounded px-1.5 py-0.5 text-xs font-medium ${cls}`}>{label}</span>;
                  })()}
                </td>
                <td className="px-3 py-2 text-xs text-gray-500 dark:text-gray-400">{formatDateTime(bid.createdAt)}</td>
                <td className="px-3 py-2 text-xs">
                  <Link to={`/listings/${bid.listingId}`} className="text-indigo-600 hover:text-indigo-500">View Listing</Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default function ShopperDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { shopper, isLoading, activeBids, wonBids, lostBids } =
    useShopperDetailViewModel(id!);

  if (isLoading || !shopper) return <LoadingSpinner />;

  return (
    <div>
      <Link to="/shoppers" className="text-sm text-indigo-600 hover:text-indigo-500">&larr; Back to Shoppers</Link>

      <h2 className="mt-4 text-xl font-bold text-gray-900 dark:text-white">{shopper.displayName || shopper.shopperId}</h2>

      <div className="mt-4 grid grid-cols-2 gap-6">
        <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-4">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-white">Shopper Info</h3>
          <dl className="mt-3 space-y-2 text-sm">
            {[
              ['Shopper ID', shopper.shopperId],
              ['Member ID', shopper.memberId],
              ['Customer ID', shopper.customerId],
              ['Display Name', shopper.displayName || '-'],
            ].map(([label, value]) => (
              <div key={String(label)} className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">{String(label)}</dt>
                <dd className="font-medium text-gray-900 dark:text-white">{String(value)}</dd>
              </div>
            ))}
          </dl>
        </div>

        <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-4">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-white">Bid Summary</h3>
          <dl className="mt-3 space-y-2 text-sm">
            {[
              ['Total Bids', shopper.bidHistory.length],
              ['Active', activeBids.length],
              ['Won', wonBids.length],
              ['Lost', lostBids.length],
            ].map(([label, value]) => (
              <div key={String(label)} className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">{String(label)}</dt>
                <dd className="font-medium text-gray-900 dark:text-white">{String(value)}</dd>
              </div>
            ))}
          </dl>
        </div>
      </div>

      <BidTable title="Active Bids" bids={activeBids} colorClass="bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200" />
      <BidTable title="Won Bids" bids={wonBids} colorClass="bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200" />
      <BidTable title="Lost Bids" bids={lostBids} colorClass="bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200" />
    </div>
  );
}
