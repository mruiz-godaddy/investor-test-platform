import { useState } from 'react';
import { Dialog, DialogPanel, DialogTitle, DialogBackdrop } from '@headlessui/react';
import type { AdminListing, ListingStatus } from '../../../domain/entities/Listing';
import { ListingStatus as LS } from '../../../domain/entities/Listing';

interface Props {
  open: boolean;
  onClose: () => void;
  listing: AdminListing | null;
  onSubmit: (id: number, status: ListingStatus) => void;
}

export default function StatusTransitionDialog({ open, onClose, listing, onSubmit }: Props) {
  const [status, setStatus] = useState<ListingStatus>(LS.SOLD);

  if (!listing) return null;

  return (
    <Dialog open={open} onClose={onClose} className="relative z-50">
      <DialogBackdrop className="fixed inset-0 bg-black/30" />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel className="w-full max-w-sm rounded-xl bg-white dark:bg-gray-900 p-6 shadow-xl">
          <DialogTitle className="text-lg font-semibold text-gray-900 dark:text-white">Force Status</DialogTitle>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">{listing.domainName}</p>
          <div className="mt-4 flex gap-2">
            {([LS.OPEN, LS.SOLD, LS.CLOSED] as const).map((s) => (
              <button
                key={s}
                onClick={() => setStatus(s)}
                className={`rounded-md px-3 py-2 text-sm font-medium ${
                  status === s ? 'bg-indigo-600 text-white' : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700'
                }`}
              >
                {s}
              </button>
            ))}
          </div>
          <div className="mt-6 flex justify-end gap-3">
            <button type="button" onClick={onClose}
              className="rounded-md px-3 py-2 text-sm font-semibold text-gray-900 dark:text-white ring-1 ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-800">Cancel</button>
            <button type="button"
              onClick={() => { onSubmit(listing.listingId, status); onClose(); }}
              className="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-500">Apply</button>
          </div>
        </DialogPanel>
      </div>
    </Dialog>
  );
}
