import { useState } from 'react';
import { Dialog, DialogPanel, DialogTitle, DialogBackdrop } from '@headlessui/react';
import type { CreateListingRequest } from '../../../domain/entities/Listing';
import type { Shopper } from '../../../domain/entities/Shopper';
import {
  DEFAULT_ASKING_PRICE_MICROS,
  DEFAULT_AUCTION_TYPE_ID,
  DEFAULT_LISTING_TYPE,
  DEFAULT_AUTO_EXT_ENABLED,
  DEFAULT_AUTO_EXT_WINDOW_SEC,
  DEFAULT_AUTO_EXT_SECONDS,
  DEFAULT_LISTING_DURATION_MINUTES,
} from '../../../lib/constants';
import { usdToMicros, microsToUsd } from '../../../domain/entities/Price';

interface Props {
  open: boolean;
  onClose: () => void;
  onSubmit: (req: CreateListingRequest) => void;
  shoppers: Shopper[];
}

export default function CreateListingForm({ open, onClose, onSubmit, shoppers }: Props) {
  const [domainName, setDomainName] = useState('');
  const [sellerShopperId, setSellerShopperId] = useState(shoppers[0]?.shopperId ?? '');
  const [askingPriceUsd, setAskingPriceUsd] = useState(microsToUsd(DEFAULT_ASKING_PRICE_MICROS));
  const [durationMinutes, setDurationMinutes] = useState(DEFAULT_LISTING_DURATION_MINUTES);
  const [auctionTypeId, setAuctionTypeId] = useState(DEFAULT_AUCTION_TYPE_ID);
  const [listingType, setListingType] = useState(DEFAULT_LISTING_TYPE);
  const [autoExtEnabled, setAutoExtEnabled] = useState(DEFAULT_AUTO_EXT_ENABLED);
  const [autoExtWindowSec, setAutoExtWindowSec] = useState(DEFAULT_AUTO_EXT_WINDOW_SEC);
  const [autoExtSeconds, setAutoExtSeconds] = useState(DEFAULT_AUTO_EXT_SECONDS);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const endTime = new Date(Date.now() + durationMinutes * 60_000).toISOString();
    onSubmit({
      domainName,
      sellerShopperId,
      askingPriceUsd: usdToMicros(askingPriceUsd),
      endTime,
      auctionTypeId,
      listingType,
      autoExtEnabled,
      autoExtWindowSec,
      autoExtSeconds,
    });
    setDomainName('');
    onClose();
  };

  return (
    <Dialog open={open} onClose={onClose} className="relative z-50">
      <DialogBackdrop className="fixed inset-0 bg-black/30" />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel className="w-full max-w-lg rounded-xl bg-white dark:bg-gray-900 p-6 shadow-xl">
          <DialogTitle className="text-lg font-semibold text-gray-900 dark:text-white">Create Listing</DialogTitle>
          <form onSubmit={handleSubmit} className="mt-4 space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-200">Domain Name *</label>
              <input required value={domainName} onChange={(e) => setDomainName(e.target.value)}
                className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-sm" placeholder="example.com" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-200">Seller *</label>
              <select value={sellerShopperId} onChange={(e) => setSellerShopperId(e.target.value)}
                className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-sm">
                {shoppers.map((s) => (
                  <option key={s.shopperId} value={s.shopperId}>{s.shopperId}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-200">Asking Price ($)</label>
              <input type="number" step="0.01" value={askingPriceUsd} onChange={(e) => setAskingPriceUsd(Number(e.target.value))}
                className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-sm" />
            </div>
            <div className="grid grid-cols-3 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200">Duration (min)</label>
                <input type="number" value={durationMinutes} onChange={(e) => setDurationMinutes(Number(e.target.value))}
                  className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200">Auction Type</label>
                <input type="number" value={auctionTypeId} onChange={(e) => setAuctionTypeId(Number(e.target.value))}
                  className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200">Listing Type</label>
                <input value={listingType} onChange={(e) => setListingType(e.target.value)}
                  className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-sm" />
              </div>
            </div>
            <div className="grid grid-cols-3 gap-3">
              <div className="flex items-center gap-2">
                <input type="checkbox" checked={autoExtEnabled} onChange={(e) => setAutoExtEnabled(e.target.checked)}
                  className="h-4 w-4 rounded border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white text-indigo-600" />
                <label className="text-sm text-gray-700 dark:text-gray-200">Auto-Extend</label>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200">Window (s)</label>
                <input type="number" value={autoExtWindowSec} onChange={(e) => setAutoExtWindowSec(Number(e.target.value))}
                  className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200">Extension (s)</label>
                <input type="number" value={autoExtSeconds} onChange={(e) => setAutoExtSeconds(Number(e.target.value))}
                  className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-sm" />
              </div>
            </div>
            <div className="flex justify-end gap-3 pt-4">
              <button type="button" onClick={onClose}
                className="rounded-md px-3 py-2 text-sm font-semibold text-gray-900 dark:text-white ring-1 ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-800">Cancel</button>
              <button type="submit"
                className="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-500">Create</button>
            </div>
          </form>
        </DialogPanel>
      </div>
    </Dialog>
  );
}
