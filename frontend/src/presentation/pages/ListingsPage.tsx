import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useListingsViewModel } from '../hooks/useListingsViewModel';
import ListingsTable from '../components/listings/ListingsTable';
import CreateListingForm from '../components/listings/CreateListingForm';
import StatusTransitionDialog from '../components/listings/StatusTransitionDialog';
import EndTimeAdjustDialog from '../components/listings/EndTimeAdjustDialog';
import SniperBidDialog from '../components/listings/SniperBidDialog';
import LoadingSpinner from '../components/shared/LoadingSpinner';
import type { AdminListing } from '../../domain/entities/Listing';

export default function ListingsPage() {
  const navigate = useNavigate();
  const { listings, shoppers, isLoading, createListing, updateStatus, updateEndTime, placeSniperBid } = useListingsViewModel();

  const [showCreate, setShowCreate] = useState(false);
  const [statusTarget, setStatusTarget] = useState<AdminListing | null>(null);
  const [extendTarget, setExtendTarget] = useState<AdminListing | null>(null);
  const [bidTarget, setBidTarget] = useState<AdminListing | null>(null);

  if (isLoading) return <LoadingSpinner />;

  return (
    <div>
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold text-gray-900">Listings</h2>
        <button
          onClick={() => setShowCreate(true)}
          className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
        >
          Create Listing
        </button>
      </div>
      <div className="mt-4">
        <ListingsTable
          listings={listings}
          onRowClick={(l) => navigate(`/listings/${l.listingId}`)}
          onForceStatus={setStatusTarget}
          onExtendTime={setExtendTarget}
          onSniperBid={setBidTarget}
        />
      </div>

      <CreateListingForm open={showCreate} onClose={() => setShowCreate(false)} onSubmit={createListing} shoppers={shoppers} />
      <StatusTransitionDialog open={!!statusTarget} onClose={() => setStatusTarget(null)} listing={statusTarget}
        onSubmit={(id, status) => updateStatus({ id, status })} />
      <EndTimeAdjustDialog open={!!extendTarget} onClose={() => setExtendTarget(null)} listing={extendTarget}
        onSubmit={(id, update) => updateEndTime({ id, update })} />
      <SniperBidDialog open={!!bidTarget} onClose={() => setBidTarget(null)} listing={bidTarget} shoppers={shoppers}
        onSubmit={(id, shopperId, bidAmountUsd) => placeSniperBid({ id, shopperId, bidAmountUsd })} />
    </div>
  );
}
