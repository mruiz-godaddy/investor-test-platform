import { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useListingDetailViewModel } from '../hooks/useListingDetailViewModel';
import ListingStatusBadge from '../components/listings/ListingStatusBadge';
import CountdownTimer from '../components/listings/CountdownTimer';
import PriceDisplay from '../components/shared/PriceDisplay';
import StatusTransitionDialog from '../components/listings/StatusTransitionDialog';
import EndTimeAdjustDialog from '../components/listings/EndTimeAdjustDialog';
import SniperBidDialog from '../components/listings/SniperBidDialog';
import LoadingSpinner from '../components/shared/LoadingSpinner';
import { formatDateTime } from '../../lib/formatters';

export default function ListingDetailPage() {
  const { id } = useParams<{ id: string }>();
  const listingId = Number(id);
  const { listing, shoppers, isLoading, updateStatus, updateEndTime, placeSniperBid } =
    useListingDetailViewModel(listingId);

  const [showStatus, setShowStatus] = useState(false);
  const [showExtend, setShowExtend] = useState(false);
  const [showBid, setShowBid] = useState(false);

  if (isLoading || !listing) return <LoadingSpinner />;

  return (
    <div>
      <Link to="/listings" className="text-sm text-indigo-600 hover:text-indigo-500">&larr; Back to Listings</Link>

      <div className="mt-4 flex items-center gap-4">
        <h2 className="text-xl font-bold text-gray-900 dark:text-white">{listing.domainName}</h2>
        <ListingStatusBadge status={listing.listingStatus} />
        <CountdownTimer endTime={listing.endTime} />
      </div>

      <div className="mt-4 flex gap-2">
        <button onClick={() => setShowStatus(true)} className="rounded-md bg-gray-100 dark:bg-gray-800 px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700">Force Status</button>
        <button onClick={() => setShowExtend(true)} className="rounded-md bg-gray-100 dark:bg-gray-800 px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700">Adjust End Time</button>
        <button onClick={() => setShowBid(true)} className="rounded-md bg-gray-100 dark:bg-gray-800 px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700">Sniper Bid</button>
      </div>

      <div className="mt-6 grid grid-cols-2 gap-6">
        <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-4">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-white">Listing Info</h3>
          <dl className="mt-3 space-y-2 text-sm">
            {[
              ['Listing ID', listing.listingId],
              ['Type', listing.listingType],
              ['Auction Type', listing.auctionTypeId],
              ['Seller', listing.sellerShopperId],
              ['Start Time', formatDateTime(listing.startTime)],
              ['End Time', formatDateTime(listing.endTime)],
              ['Created', formatDateTime(listing.createdAt)],
              ['Auto-Extend', listing.autoExtEnabled ? `Yes (${listing.autoExtWindowSec}s / +${listing.autoExtSeconds}s)` : 'No'],
              ['Extended?', listing.isAutoExtended ? 'Yes' : 'No'],
            ].map(([label, value]) => (
              <div key={String(label)} className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">{String(label)}</dt>
                <dd className="font-medium text-gray-900 dark:text-white">{String(value)}</dd>
              </div>
            ))}
          </dl>
        </div>

        <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-4">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-white">Pricing</h3>
          <dl className="mt-3 space-y-2 text-sm">
            {[
              ['Asking', listing.askingPriceUsd],
              ['Current', listing.currentPriceUsd],
              ['Next Bid', listing.nextBidPriceUsd],
              ['Reserve', listing.reservePriceUsd],
              ['Sale', listing.salePriceUsd],
            ].map(([label, value]) => (
              <div key={String(label)} className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">{String(label)}</dt>
                <dd className="font-medium text-gray-900 dark:text-white">
                  {value != null ? <PriceDisplay micros={value as number} /> : '-'}
                </dd>
              </div>
            ))}
            <div className="flex justify-between">
              <dt className="text-gray-500 dark:text-gray-400">Reserve Met?</dt>
              <dd className="font-medium text-gray-900 dark:text-white">{listing.isReserveMet ? 'Yes' : 'No'}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500 dark:text-gray-400">Bids / Bidders</dt>
              <dd className="font-medium text-gray-900 dark:text-white">{listing.bidsCount} / {listing.biddersCount}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500 dark:text-gray-400">Highest Bidder</dt>
              <dd className="font-medium text-gray-900 dark:text-white">{listing.highestBidderShopper || '-'}</dd>
            </div>
          </dl>
        </div>
      </div>

      <div className="mt-6">
        <h3 className="text-sm font-semibold text-gray-900 dark:text-white">Bid History</h3>
        {listing.bidHistory.length === 0 ? (
          <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">No bids yet.</p>
        ) : (
          <div className="mt-2 overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-950">
                <tr>
                  {['Bid ID', 'Shopper', 'Amount', 'Type', 'Status', 'High Bid', 'Parent', 'Created'].map((h) => (
                    <th key={h} className="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700 bg-white dark:bg-gray-900">
                {listing.bidHistory.map((bid) => (
                  <tr key={bid.bidId}>
                    <td className="px-3 py-2 text-xs text-gray-500 dark:text-gray-400 font-mono">{bid.bidId.slice(0, 8)}</td>
                    <td className="px-3 py-2 text-xs text-gray-900 dark:text-white">{bid.shopperId}</td>
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
                        } else if (listing.listingStatus === 'SOLD' && bid.shopperId === listing.highestBidderShopper && bid.isHighBid) {
                          label = 'WON';
                          cls = 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200';
                        } else if (listing.listingStatus === 'SOLD' || listing.listingStatus === 'CLOSED') {
                          label = 'LOST';
                          cls = 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200';
                        }
                        return <span className={`inline-flex rounded px-1.5 py-0.5 text-xs font-medium ${cls}`}>{label}</span>;
                      })()}
                    </td>
                    <td className="px-3 py-2 text-xs text-gray-900 dark:text-white">{bid.isHighBid ? 'Yes' : '-'}</td>
                    <td className="px-3 py-2 text-xs text-gray-500 dark:text-gray-400 font-mono">{bid.parentBidId ? bid.parentBidId.slice(0, 8) : '-'}</td>
                    <td className="px-3 py-2 text-xs text-gray-500 dark:text-gray-400">{formatDateTime(bid.createdAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <StatusTransitionDialog open={showStatus} onClose={() => setShowStatus(false)} listing={listing}
        onSubmit={(_, status) => updateStatus(status)} />
      <EndTimeAdjustDialog open={showExtend} onClose={() => setShowExtend(false)} listing={listing}
        onSubmit={(_, update) => updateEndTime(update)} />
      <SniperBidDialog open={showBid} onClose={() => setShowBid(false)} listing={listing} shoppers={shoppers}
        onSubmit={(_, shopperId, bidAmountUsd) => placeSniperBid({ shopperId, bidAmountUsd })} />
    </div>
  );
}
