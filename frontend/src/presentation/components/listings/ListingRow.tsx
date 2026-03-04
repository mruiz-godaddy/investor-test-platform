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
}

export default function ListingRow({ listing, onClick, onForceStatus, onExtendTime, onSniperBid }: Props) {
  return (
    <tr className="hover:bg-gray-50 cursor-pointer" onClick={onClick}>
      <td className="px-4 py-3 text-sm font-medium text-gray-900">{listing.domainName}</td>
      <td className="px-4 py-3"><ListingStatusBadge status={listing.listingStatus} /></td>
      <td className="px-4 py-3"><CountdownTimer endTime={listing.endTime} /></td>
      <td className="px-4 py-3 text-sm"><PriceDisplay micros={listing.currentPriceUsd} /></td>
      <td className="px-4 py-3 text-sm">
        {listing.bidsCount > 0
          ? <PriceDisplay micros={Math.max(...listing.bidHistory.map((b) => b.bidAmountUsd))} />
          : <span className="text-gray-400">-</span>
        }
      </td>
      <td className="px-4 py-3 text-sm text-gray-500">{listing.bidsCount}</td>
      <td className="px-4 py-3 text-sm text-gray-500">{listing.biddersCount}</td>
      <td className="px-4 py-3">
        <div className="flex gap-1" onClick={(e) => e.stopPropagation()}>
          <button
            onClick={() => onForceStatus(listing)}
            className="rounded bg-gray-100 px-2 py-1 text-xs text-gray-700 hover:bg-gray-200"
          >
            Status
          </button>
          <button
            onClick={() => onExtendTime(listing)}
            className="rounded bg-gray-100 px-2 py-1 text-xs text-gray-700 hover:bg-gray-200"
          >
            Extend
          </button>
          <button
            onClick={() => onSniperBid(listing)}
            className="rounded bg-gray-100 px-2 py-1 text-xs text-gray-700 hover:bg-gray-200"
          >
            Bid
          </button>
        </div>
      </td>
    </tr>
  );
}
