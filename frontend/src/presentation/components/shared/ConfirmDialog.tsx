import { Dialog, DialogPanel, DialogTitle, DialogBackdrop } from '@headlessui/react';

interface Props {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  description?: string;
  confirmLabel?: string;
  confirmVariant?: 'danger' | 'primary';
  children?: React.ReactNode;
}

export default function ConfirmDialog({
  open,
  onClose,
  onConfirm,
  title,
  description,
  confirmLabel = 'Confirm',
  confirmVariant = 'primary',
  children,
}: Props) {
  const btnClass =
    confirmVariant === 'danger'
      ? 'bg-red-600 hover:bg-red-500 text-white'
      : 'bg-indigo-600 hover:bg-indigo-500 text-white';

  return (
    <Dialog open={open} onClose={onClose} className="relative z-50">
      <DialogBackdrop className="fixed inset-0 bg-black/30" />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel className="w-full max-w-md rounded-xl bg-white dark:bg-gray-900 p-6 shadow-xl">
          <DialogTitle className="text-lg font-semibold text-gray-900 dark:text-white">{title}</DialogTitle>
          {description && <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">{description}</p>}
          {children && <div className="mt-4">{children}</div>}
          <div className="mt-6 flex justify-end gap-3">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md px-3 py-2 text-sm font-semibold text-gray-900 dark:text-white ring-1 ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-800"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={onConfirm}
              className={`rounded-md px-3 py-2 text-sm font-semibold ${btnClass}`}
            >
              {confirmLabel}
            </button>
          </div>
        </DialogPanel>
      </div>
    </Dialog>
  );
}
