import { useState } from 'react';
import type { AdminListing, ListingStatus } from '../../../domain/entities/Listing';
import ListingRow from './ListingRow';
import EmptyState from '../shared/EmptyState';

interface Props {
  listings: AdminListing[];
  onRowClick: (listing: AdminListing) => void;
  onForceStatus: (listing: AdminListing) => void;
  onExtendTime: (listing: AdminListing) => void;
  onSniperBid: (listing: AdminListing) => void;
}

type SortKey = 'domainName' | 'endTime' | 'currentPriceUsd' | 'bidsCount';

export default function ListingsTable({ listings, onRowClick, onForceStatus, onExtendTime, onSniperBid }: Props) {
  const [statusFilter, setStatusFilter] = useState<ListingStatus | 'ALL'>('ALL');
  const [sortKey, setSortKey] = useState<SortKey>('endTime');
  const [sortAsc, setSortAsc] = useState(true);

  const filtered = statusFilter === 'ALL'
    ? listings
    : listings.filter((l) => l.listingStatus === statusFilter);

  const sorted = [...filtered].sort((a, b) => {
    const va = a[sortKey];
    const vb = b[sortKey];
    const cmp = typeof va === 'string' ? va.localeCompare(vb as string) : (va as number) - (vb as number);
    return sortAsc ? cmp : -cmp;
  });

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) setSortAsc(!sortAsc);
    else { setSortKey(key); setSortAsc(true); }
  };

  if (listings.length === 0) {
    return <EmptyState title="No listings" description="Create a listing or load a scenario to get started." />;
  }

  const headers: { key: SortKey; label: string }[] = [
    { key: 'domainName', label: 'Domain' },
  ];

  return (
    <div>
      <div className="mb-3 flex gap-2">
        {(['ALL', 'OPEN', 'SOLD', 'CLOSED'] as const).map((s) => (
          <button
            key={s}
            onClick={() => setStatusFilter(s)}
            className={`rounded-md px-3 py-1 text-xs font-medium ${
              statusFilter === s ? 'bg-indigo-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            {s}
          </button>
        ))}
      </div>
      <div className="overflow-x-auto rounded-lg border border-gray-200">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              {[
                { key: 'domainName' as SortKey, label: 'Domain' },
                { key: null, label: 'Status' },
                { key: 'endTime' as SortKey, label: 'Countdown' },
                { key: 'currentPriceUsd' as SortKey, label: 'Price' },
                { key: null, label: 'High Bid' },
                { key: 'bidsCount' as SortKey, label: 'Bids' },
                { key: null, label: 'Bidders' },
                { key: null, label: 'Actions' },
              ].map((h, i) => (
                <th
                  key={i}
                  className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 cursor-pointer select-none"
                  onClick={() => h.key && toggleSort(h.key)}
                >
                  {h.label}
                  {h.key && sortKey === h.key && (sortAsc ? ' \u2191' : ' \u2193')}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 bg-white">
            {sorted.map((listing) => (
              <ListingRow
                key={listing.listingId}
                listing={listing}
                onClick={() => onRowClick(listing)}
                onForceStatus={onForceStatus}
                onExtendTime={onExtendTime}
                onSniperBid={onSniperBid}
              />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
