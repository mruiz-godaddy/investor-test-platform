import { useState } from 'react';
import { Dialog, DialogPanel, DialogTitle, DialogBackdrop } from '@headlessui/react';
import type { AdminListing } from '../../../domain/entities/Listing';
import type { Shopper } from '../../../domain/entities/Shopper';
import { usdToMicros, microsToUsd } from '../../../domain/entities/Price';

interface Props {
  open: boolean;
  onClose: () => void;
  listing: AdminListing | null;
  shoppers: Shopper[];
  onSubmit: (id: number, shopperId: string, bidAmountUsd: number) => void;
}

export default function SniperBidDialog({ open, onClose, listing, shoppers, onSubmit }: Props) {
  const [shopperId, setShopperId] = useState(shoppers[0]?.shopperId ?? '');
  const [bidUsd, setBidUsd] = useState('');

  if (!listing) return null;

  const suggestedBid = microsToUsd(listing.nextBidPriceUsd);

  return (
    <Dialog open={open} onClose={onClose} className="relative z-50">
      <DialogBackdrop className="fixed inset-0 bg-black/30" />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel className="w-full max-w-sm rounded-xl bg-white p-6 shadow-xl">
          <DialogTitle className="text-lg font-semibold text-gray-900">Sniper Bid</DialogTitle>
          <p className="mt-1 text-sm text-gray-500">{listing.domainName}</p>
          <p className="mt-1 text-xs text-gray-400">Suggested: ${suggestedBid.toFixed(2)}</p>
          <div className="mt-4 space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700">Shopper</label>
              <select value={shopperId} onChange={(e) => setShopperId(e.target.value)}
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                {shoppers.map((s) => (
                  <option key={s.shopperId} value={s.shopperId}>{s.shopperId}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Bid Amount ($)</label>
              <input type="number" step="0.01" min="0.01" value={bidUsd} onChange={(e) => setBidUsd(e.target.value)}
                placeholder={suggestedBid.toFixed(2)}
                className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
          </div>
          <div className="mt-6 flex justify-end gap-3">
            <button type="button" onClick={onClose}
              className="rounded-md px-3 py-2 text-sm font-semibold text-gray-900 ring-1 ring-gray-300 hover:bg-gray-50">Cancel</button>
            <button type="button"
              disabled={!bidUsd || Number(bidUsd) <= 0}
              onClick={() => { onSubmit(listing.listingId, shopperId, usdToMicros(Number(bidUsd))); onClose(); }}
              className="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed">Place Bid</button>
          </div>
        </DialogPanel>
      </div>
    </Dialog>
  );
}
