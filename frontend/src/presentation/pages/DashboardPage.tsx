import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDashboardViewModel } from '../hooks/useDashboardViewModel';
import { useListingsViewModel } from '../hooks/useListingsViewModel';
import ListingsTable from '../components/listings/ListingsTable';
import QuickCreateControl from '../components/listings/QuickCreateControl';
import StatusTransitionDialog from '../components/listings/StatusTransitionDialog';
import EndTimeAdjustDialog from '../components/listings/EndTimeAdjustDialog';
import SniperBidDialog from '../components/listings/SniperBidDialog';
import LoadingSpinner from '../components/shared/LoadingSpinner';
import type { AdminListing } from '../../domain/entities/Listing';

export default function DashboardPage() {
  const navigate = useNavigate();
  const { listings, isLoading, totalCount, openCount, soldCount, closedCount } = useDashboardViewModel();
  const { shoppers, createListing, setupSystem, isSettingUp, isCreating, updateStatus, updateEndTime, placeSniperBid } = useListingsViewModel();

  const [statusTarget, setStatusTarget] = useState<AdminListing | null>(null);
  const [extendTarget, setExtendTarget] = useState<AdminListing | null>(null);
  const [bidTarget, setBidTarget] = useState<AdminListing | null>(null);

  if (isLoading) return <LoadingSpinner />;

  const hasListings = listings.length > 0;

  const cards = [
    { label: 'Total', value: totalCount, color: 'bg-gray-100 text-gray-800' },
    { label: 'Open', value: openCount, color: 'bg-green-100 text-green-800' },
    { label: 'Sold', value: soldCount, color: 'bg-blue-100 text-blue-800' },
    { label: 'Closed', value: closedCount, color: 'bg-red-100 text-red-800' },
  ];

  return (
    <div>
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold text-gray-900">Dashboard</h2>
        {hasListings && (
          <QuickCreateControl onSubmit={createListing} isPending={isCreating} />
        )}
      </div>

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
