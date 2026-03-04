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
        <h2 className="text-xl font-bold text-gray-900">{listing.domainName}</h2>
        <ListingStatusBadge status={listing.listingStatus} />
        <CountdownTimer endTime={listing.endTime} />
      </div>

      <div className="mt-4 flex gap-2">
        <button onClick={() => setShowStatus(true)} className="rounded-md bg-gray-100 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200">Force Status</button>
        <button onClick={() => setShowExtend(true)} className="rounded-md bg-gray-100 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200">Adjust End Time</button>
        <button onClick={() => setShowBid(true)} className="rounded-md bg-gray-100 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200">Sniper Bid</button>
      </div>

      <div className="mt-6 grid grid-cols-2 gap-6">
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h3 className="text-sm font-semibold text-gray-900">Listing Info</h3>
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
                <dt className="text-gray-500">{String(label)}</dt>
                <dd className="font-medium text-gray-900">{String(value)}</dd>
              </div>
            ))}
          </dl>
        </div>

        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h3 className="text-sm font-semibold text-gray-900">Pricing</h3>
          <dl className="mt-3 space-y-2 text-sm">
            {[
              ['Asking', listing.askingPriceUsd],
              ['Current', listing.currentPriceUsd],
              ['Next Bid', listing.nextBidPriceUsd],
              ['Reserve', listing.reservePriceUsd],
              ['Sale', listing.salePriceUsd],
            ].map(([label, value]) => (
              <div key={String(label)} className="flex justify-between">
                <dt className="text-gray-500">{String(label)}</dt>
                <dd className="font-medium text-gray-900">
                  {value != null ? <PriceDisplay micros={value as number} /> : '-'}
                </dd>
              </div>
            ))}
            <div className="flex justify-between">
              <dt className="text-gray-500">Reserve Met?</dt>
              <dd className="font-medium text-gray-900">{listing.isReserveMet ? 'Yes' : 'No'}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500">Bids / Bidders</dt>
              <dd className="font-medium text-gray-900">{listing.bidsCount} / {listing.biddersCount}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500">Highest Bidder</dt>
              <dd className="font-medium text-gray-900">{listing.highestBidderShopper || '-'}</dd>
            </div>
          </dl>
        </div>
      </div>

      <div className="mt-6">
        <h3 className="text-sm font-semibold text-gray-900">Bid History</h3>
        {listing.bidHistory.length === 0 ? (
          <p className="mt-2 text-sm text-gray-500">No bids yet.</p>
        ) : (
          <div className="mt-2 overflow-x-auto rounded-lg border border-gray-200">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  {['Bid ID', 'Shopper', 'Amount', 'Type', 'Status', 'High Bid', 'Parent', 'Created'].map((h) => (
                    <th key={h} className="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 bg-white">
                {listing.bidHistory.map((bid) => (
                  <tr key={bid.bidId}>
                    <td className="px-3 py-2 text-xs text-gray-500 font-mono">{bid.bidId.slice(0, 8)}</td>
                    <td className="px-3 py-2 text-xs text-gray-900">{bid.shopperId}</td>
                    <td className="px-3 py-2 text-xs"><PriceDisplay micros={bid.bidAmountUsd} /></td>
                    <td className="px-3 py-2">
                      <span className={`inline-flex rounded px-1.5 py-0.5 text-xs font-medium ${
                        bid.bidType === 'PROXY' ? 'bg-purple-100 text-purple-800' : 'bg-gray-100 text-gray-800'
                      }`}>{bid.bidType}</span>
                    </td>
                    <td className="px-3 py-2">
                      <span className={`inline-flex rounded px-1.5 py-0.5 text-xs font-medium ${
                        bid.bidStatus === 'CANCELLED' ? 'bg-red-100 text-red-800' : 'bg-green-100 text-green-800'
                      }`}>{bid.bidStatus}</span>
                    </td>
                    <td className="px-3 py-2 text-xs text-gray-900">{bid.isHighBid ? 'Yes' : '-'}</td>
                    <td className="px-3 py-2 text-xs text-gray-500 font-mono">{bid.parentBidId ? bid.parentBidId.slice(0, 8) : '-'}</td>
                    <td className="px-3 py-2 text-xs text-gray-500">{formatDateTime(bid.createdAt)}</td>
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
