import { useRef, useState } from 'react';

interface Props {
  onWipe: () => void;
  onExport: () => void;
  onImport: (file: File) => void;
  isWiping: boolean;
  isExporting: boolean;
  isImporting: boolean;
}

export default function DatabasePanel({ onWipe, onExport, onImport, isWiping, isExporting, isImporting }: Props) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [showWipeConfirm, setShowWipeConfirm] = useState(false);

  const handleImportClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      onImport(file);
      e.target.value = '';
    }
  };

  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4">
      <h3 className="text-sm font-semibold text-gray-900">Database</h3>
      <p className="mt-1 text-xs text-gray-500">Wipe, export, or import the database state.</p>

      <div className="mt-4 space-y-3">
        {/* Wipe DB */}
        {!showWipeConfirm ? (
          <button
            onClick={() => setShowWipeConfirm(true)}
            disabled={isWiping}
            className="w-full rounded-md bg-red-600 px-4 py-2 text-sm font-semibold text-white hover:bg-red-500 disabled:opacity-50"
          >
            {isWiping ? 'Wiping...' : 'Wipe Database'}
          </button>
        ) : (
          <div className="rounded-md border border-red-200 bg-red-50 p-3">
            <p className="text-xs font-medium text-red-800">This will permanently delete all data. Are you sure?</p>
            <div className="mt-2 flex gap-2">
              <button
                onClick={() => { onWipe(); setShowWipeConfirm(false); }}
                disabled={isWiping}
                className="rounded-md bg-red-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-red-500 disabled:opacity-50"
              >
                {isWiping ? 'Wiping...' : 'Yes, Wipe'}
              </button>
              <button
                onClick={() => setShowWipeConfirm(false)}
                className="rounded-md bg-white px-3 py-1.5 text-xs font-semibold text-gray-700 ring-1 ring-gray-300 hover:bg-gray-50"
              >
                Cancel
              </button>
            </div>
          </div>
        )}

        {/* Export */}
        <button
          onClick={onExport}
          disabled={isExporting}
          className="w-full rounded-md bg-emerald-600 px-4 py-2 text-sm font-semibold text-white hover:bg-emerald-500 disabled:opacity-50"
        >
          {isExporting ? 'Exporting...' : 'Export Database'}
        </button>

        {/* Import */}
        <button
          onClick={handleImportClick}
          disabled={isImporting}
          className="w-full rounded-md bg-blue-600 px-4 py-2 text-sm font-semibold text-white hover:bg-blue-500 disabled:opacity-50"
        >
          {isImporting ? 'Importing...' : 'Import Database'}
        </button>
        <input
          ref={fileInputRef}
          type="file"
          accept=".json"
          onChange={handleFileChange}
          className="hidden"
        />
      </div>
    </div>
  );
}
