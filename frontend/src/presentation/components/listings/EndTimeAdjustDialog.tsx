import { useState } from 'react';
import { Dialog, DialogPanel, DialogTitle, DialogBackdrop } from '@headlessui/react';
import type { AdminListing } from '../../../domain/entities/Listing';

interface Props {
  open: boolean;
  onClose: () => void;
  listing: AdminListing | null;
  onSubmit: (id: number, update: { endTime?: string; addSeconds?: number }) => void;
}

export default function EndTimeAdjustDialog({ open, onClose, listing, onSubmit }: Props) {
  const [mode, setMode] = useState<'relative' | 'absolute'>('relative');
  const [addSeconds, setAddSeconds] = useState(300);
  const [endTime, setEndTime] = useState('');

  if (!listing) return null;

  const handleSubmit = () => {
    if (mode === 'relative') {
      onSubmit(listing.listingId, { addSeconds });
    } else {
      onSubmit(listing.listingId, { endTime: new Date(endTime).toISOString() });
    }
    onClose();
  };

  return (
    <Dialog open={open} onClose={onClose} className="relative z-50">
      <DialogBackdrop className="fixed inset-0 bg-black/30" />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel className="w-full max-w-sm rounded-xl bg-white dark:bg-gray-900 p-6 shadow-xl">
          <DialogTitle className="text-lg font-semibold text-gray-900 dark:text-white">Adjust End Time</DialogTitle>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">{listing.domainName}</p>
          <div className="mt-4 flex gap-2">
            <button onClick={() => setMode('relative')}
              className={`rounded-md px-3 py-1 text-sm ${mode === 'relative' ? 'bg-indigo-600 text-white' : 'bg-gray-100 dark:bg-gray-800'}`}>
              + Seconds
            </button>
            <button onClick={() => setMode('absolute')}
              className={`rounded-md px-3 py-1 text-sm ${mode === 'absolute' ? 'bg-indigo-600 text-white' : 'bg-gray-100 dark:bg-gray-800'}`}>
              Absolute
            </button>
          </div>
          <div className="mt-4">
            {mode === 'relative' ? (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200">Add Seconds</label>
                <input type="number" value={addSeconds} onChange={(e) => setAddSeconds(Number(e.target.value))}
                  className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-sm" />
              </div>
            ) : (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-200">New End Time</label>
                <input type="datetime-local" value={endTime} onChange={(e) => setEndTime(e.target.value)}
                  className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-3 py-2 text-sm" />
              </div>
            )}
          </div>
          <div className="mt-6 flex justify-end gap-3">
            <button type="button" onClick={onClose}
              className="rounded-md px-3 py-2 text-sm font-semibold text-gray-900 dark:text-white ring-1 ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-800">Cancel</button>
            <button type="button" onClick={handleSubmit}
              className="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-500">Apply</button>
          </div>
        </DialogPanel>
      </div>
    </Dialog>
  );
}
