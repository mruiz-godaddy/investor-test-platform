import { useState, type ReactNode } from 'react';
import type { AdminListing, ListingStatus } from '../../../domain/entities/Listing';
import ListingRow from './ListingRow';
import EmptyState from '../shared/EmptyState';
import Pagination from '../shared/Pagination';

const PAGE_SIZE = 15;

interface Props {
  listings: AdminListing[];
  onRowClick: (listing: AdminListing) => void;
  onForceStatus: (listing: AdminListing) => void;
  onExtendTime: (listing: AdminListing) => void;
  onSniperBid: (listing: AdminListing) => void;
  onToggleRadar: (listing: AdminListing) => void;
  headerRight?: ReactNode;
}

type SortKey = 'domainName' | 'endTime' | 'currentPriceUsd' | 'bidsCount';

export default function ListingsTable({ listings, onRowClick, onForceStatus, onExtendTime, onSniperBid, onToggleRadar, headerRight }: Props) {
  const [statusFilter, setStatusFilter] = useState<ListingStatus | 'ALL'>('ALL');
  const [sortKey, setSortKey] = useState<SortKey>('endTime');
  const [sortAsc, setSortAsc] = useState(true);
  const [page, setPage] = useState(1);

  const filtered = statusFilter === 'ALL'
    ? listings
    : listings.filter((l) => l.listingStatus === statusFilter);

  const sorted = [...filtered].sort((a, b) => {
    const va = a[sortKey];
    const vb = b[sortKey];
    const cmp = typeof va === 'string' ? va.localeCompare(vb as string) : (va as number) - (vb as number);
    return sortAsc ? cmp : -cmp;
  });

  const totalPages = Math.max(1, Math.ceil(sorted.length / PAGE_SIZE));
  const safePage = Math.min(page, totalPages);
  const paginated = sorted.slice((safePage - 1) * PAGE_SIZE, safePage * PAGE_SIZE);

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
      <div className="mb-3 flex items-end justify-between">
        <div className="flex gap-2">
          {([
            { key: 'ALL', label: 'ALL' },
            { key: 'OPEN', label: 'OPEN' },
            { key: 'SOLD', label: 'SOLD' },
            { key: 'CLOSED', label: 'CLOSED' },
          ] as const).map((s) => (
            <button
              key={s.key}
              onClick={() => { setStatusFilter(s.key); setPage(1); }}
              className={`rounded-md px-3 py-1 text-xs font-medium ${
                statusFilter === s.key ? 'bg-indigo-600 text-white' : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700'
              }`}
            >
              {s.label}
            </button>
          ))}
        </div>
        {headerRight}
      </div>
      <div className="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-950">
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
                  className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400 cursor-pointer select-none"
                  onClick={() => h.key && toggleSort(h.key)}
                >
                  {h.label}
                  {h.key && sortKey === h.key && (sortAsc ? ' \u2191' : ' \u2193')}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-700 bg-white dark:bg-gray-900">
            {paginated.map((listing) => (
              <ListingRow
                key={listing.listingId}
                listing={listing}
                onClick={() => onRowClick(listing)}
                onForceStatus={onForceStatus}
                onExtendTime={onExtendTime}
                onSniperBid={onSniperBid}
                onToggleRadar={onToggleRadar}
              />
            ))}
          </tbody>
        </table>
      </div>
      <Pagination
        currentPage={safePage}
        totalPages={totalPages}
        onPageChange={setPage}
        totalItems={sorted.length}
        pageSize={PAGE_SIZE}
      />
    </div>
  );
}
