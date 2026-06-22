import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useBinDomainsViewModel } from '../hooks/useBinDomainsViewModel';
import { useListingsViewModel } from '../hooks/useListingsViewModel';
import ListingsTable from '../components/listings/ListingsTable';
import BinCreateControl from '../components/listings/BinCreateControl';
import CartEventsPanel from '../components/listings/CartEventsPanel';
import StatusTransitionDialog from '../components/listings/StatusTransitionDialog';
import EndTimeAdjustDialog from '../components/listings/EndTimeAdjustDialog';
import LoadingSpinner from '../components/shared/LoadingSpinner';
import type { AdminListing } from '../../domain/entities/Listing';

export default function BinDomainsPage() {
  const navigate = useNavigate();
  const { listings, isLoading, totalCount, openCount, soldCount, closedCount } = useBinDomainsViewModel();
  const { createListing, generateBin, isGeneratingBin, isCreating, updateStatus, updateEndTime, toggleRadar } = useListingsViewModel();

  const [statusTarget, setStatusTarget] = useState<AdminListing | null>(null);
  const [extendTarget, setExtendTarget] = useState<AdminListing | null>(null);

  if (isLoading) return <LoadingSpinner />;

  const hasListings = listings.length > 0;

  const cards = [
    { label: 'Total', value: totalCount, color: 'bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-100' },
    { label: 'Open', value: openCount, color: 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200' },
    { label: 'Sold', value: soldCount, color: 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200' },
    { label: 'Closed', value: closedCount, color: 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200' },
  ];

  return (
    <div>
      <h2 className="text-xl font-bold text-gray-900 dark:text-white">BIN Domains</h2>
      <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
        Closeout / Buy-It-Now / OCO inventory, added on top of the auctions. Each type drives a distinct itc code.
      </p>

      {hasListings && (
        <div className="mt-4 grid grid-cols-4 gap-4">
          {cards.map((c) => (
            <div key={c.label} className={`rounded-lg p-4 ${c.color}`}>
              <p className="text-sm font-medium">{c.label}</p>
              <p className="text-2xl font-bold">{c.value}</p>
            </div>
          ))}
        </div>
      )}

      <div className="mt-6">
        {hasListings ? (
          <ListingsTable
            listings={listings}
            onRowClick={(l) => navigate(`/listings/${l.listingId}`)}
            onForceStatus={setStatusTarget}
            onExtendTime={setExtendTarget}
            onSniperBid={() => {}}
            onToggleRadar={(l) => toggleRadar({ id: l.listingId, radarVisible: !l.radarVisible })}
            headerRight={
              <BinCreateControl
                onCreate={createListing}
                onGenerate={() => generateBin({})}
                isCreating={isCreating}
                isGenerating={isGeneratingBin}
              />
            }
          />
        ) : (
          <BinCreateControl
            variant="expanded"
            onCreate={createListing}
            onGenerate={() => generateBin({})}
            isCreating={isCreating}
            isGenerating={isGeneratingBin}
          />
        )}
      </div>

      <CartEventsPanel />

      <StatusTransitionDialog
        open={!!statusTarget}
        onClose={() => setStatusTarget(null)}
        listing={statusTarget}
        onSubmit={(id, status) => updateStatus({ id, status })}
      />
      <EndTimeAdjustDialog
        open={!!extendTarget}
        onClose={() => setExtendTarget(null)}
        listing={extendTarget}
        onSubmit={(id, update) => updateEndTime({ id, update })}
      />
    </div>
  );
}
