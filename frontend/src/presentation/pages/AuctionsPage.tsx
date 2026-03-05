import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useUseCase } from '../hooks/useDiContainer';
import { GetConfigUseCase } from '../../domain/usecases/GetConfigUseCase';
import { useAuctionsViewModel } from '../hooks/useAuctionsViewModel';
import { useListingsViewModel } from '../hooks/useListingsViewModel';
import ListingsTable from '../components/listings/ListingsTable';
import QuickCreateControl from '../components/listings/QuickCreateControl';
import StatusTransitionDialog from '../components/listings/StatusTransitionDialog';
import EndTimeAdjustDialog from '../components/listings/EndTimeAdjustDialog';
import SniperBidDialog from '../components/listings/SniperBidDialog';
import LoadingSpinner from '../components/shared/LoadingSpinner';
import type { AdminListing } from '../../domain/entities/Listing';

export default function AuctionsPage() {
  const navigate = useNavigate();
  const getConfig = useUseCase(GetConfigUseCase);
  const { data: config } = useQuery({
    queryKey: ['config'],
    queryFn: () => getConfig.execute(),
  });
  const { listings, isLoading, totalCount, openCount, soldCount, closedCount } = useAuctionsViewModel();
  const { shoppers, createListing, setupSystem, isSettingUp, isCreating, updateStatus, updateEndTime, placeSniperBid } = useListingsViewModel();

  const [statusTarget, setStatusTarget] = useState<AdminListing | null>(null);
  const [extendTarget, setExtendTarget] = useState<AdminListing | null>(null);
  const [bidTarget, setBidTarget] = useState<AdminListing | null>(null);

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
      <h2 className="text-xl font-bold text-gray-900 dark:text-white">Auctions</h2>

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
            onSniperBid={setBidTarget}
            headerRight={<QuickCreateControl onSubmit={createListing} isPending={isCreating} config={config} />}
          />
        ) : (
          <QuickCreateControl onSetup={setupSystem} isPending={isSettingUp} variant="expanded" />
        )}
      </div>

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
      <SniperBidDialog
        open={!!bidTarget}
        onClose={() => setBidTarget(null)}
        listing={bidTarget}
        shoppers={shoppers}
        onSubmit={(id, shopperId, bidAmountUsd) => placeSniperBid({ id, shopperId, bidAmountUsd })}
      />
    </div>
  );
}
