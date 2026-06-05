import type { AdminListing } from '../../../domain/entities/Listing';
import ListingStatusBadge from './ListingStatusBadge';
import CountdownTimer from './CountdownTimer';
import PriceDisplay from '../shared/PriceDisplay';

interface Props {
  listing: AdminListing;
  onClick: () => void;
  onForceStatus: (listing: AdminListing) => void;
  onExtendTime: (listing: AdminListing) => void;
  onSniperBid: (listing: AdminListing) => void;
  onToggleRadar: (listing: AdminListing) => void;
}

export default function ListingRow({ listing, onClick, onForceStatus, onExtendTime, onSniperBid, onToggleRadar }: Props) {
  return (
    <tr className="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer" onClick={onClick}>
      <td className="px-4 py-3 text-sm font-medium text-gray-900 dark:text-white">{listing.domainName}</td>
      <td className="px-4 py-3"><ListingStatusBadge status={listing.listingStatus} /></td>
      <td className="px-4 py-3"><CountdownTimer endTime={listing.endTime} /></td>
      <td className="px-4 py-3 text-sm"><PriceDisplay micros={listing.currentPriceUsd} /></td>
      <td className="px-4 py-3 text-sm">
        {listing.bidsCount > 0
          ? <PriceDisplay micros={Math.max(...listing.bidHistory.map((b) => b.bidAmountUsd))} />
          : <span className="text-gray-400 dark:text-gray-500">-</span>
        }
      </td>
      <td className="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{listing.bidsCount}</td>
      <td className="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{listing.biddersCount}</td>
      <td className="px-4 py-3">
        <div className="flex gap-1" onClick={(e) => e.stopPropagation()}>
          <button
            onClick={() => onForceStatus(listing)}
            className="rounded bg-gray-100 dark:bg-gray-800 px-2 py-1 text-xs text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700"
          >
            Status
          </button>
          <button
            onClick={() => onExtendTime(listing)}
            className="rounded bg-gray-100 dark:bg-gray-800 px-2 py-1 text-xs text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700"
          >
            Extend
          </button>
          <button
            onClick={() => onSniperBid(listing)}
            className="rounded bg-gray-100 dark:bg-gray-800 px-2 py-1 text-xs text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700"
          >
            Bid
          </button>
          <button
            onClick={() => onToggleRadar(listing)}
            className={`rounded-full px-2 py-1 text-xs font-semibold transition-colors ${
              listing.radarVisible
                ? 'bg-amber-100 dark:bg-amber-900 text-amber-800 dark:text-amber-200 hover:bg-amber-200 dark:hover:bg-amber-800'
                : 'bg-gray-100 dark:bg-gray-800 text-gray-400 dark:text-gray-500 hover:bg-gray-200 dark:hover:bg-gray-700'
            }`}
          >
            Radar
          </button>
        </div>
      </td>
    </tr>
  );
}
