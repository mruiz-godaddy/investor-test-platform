import { useEffect, useState } from 'react';
import type { CreateListingRequest } from '../../../domain/entities/Listing';
import { usdToMicros } from '../../../domain/entities/Price';
import { buildDomain, randomInt, fetchWordList } from '../../../lib/wordList';
import { BIN_INVENTORY_TYPES } from '../../../lib/constants';

const ASKING_PRICE_OPTIONS = [5, 10, 15, 20, 25, 50, 75, 100];

function randomAskingPriceMicros(): number {
  return usdToMicros(ASKING_PRICE_OPTIONS[randomInt(0, ASKING_PRICE_OPTIONS.length - 1)]);
}

interface Props {
  onCreate: (req: CreateListingRequest) => void;
  onGenerate: () => void;
  isCreating?: boolean;
  isGenerating?: boolean;
  variant?: 'inline' | 'expanded';
}

export default function BinCreateControl({ onCreate, onGenerate, isCreating, isGenerating, variant = 'inline' }: Props) {
  const [typeIndex, setTypeIndex] = useState(0);
  const [minutes, setMinutes] = useState(60);
  const [words, setWords] = useState<string[]>([]);

  useEffect(() => {
    fetchWordList().then(setWords);
  }, []);

  const wordsReady = words.length > 0;
  const selected = BIN_INVENTORY_TYPES[typeIndex];

  const handleCreate = () => {
    const endTime = new Date(Date.now() + Math.max(1, minutes) * 60_000).toISOString();
    onCreate({
      domainName: buildDomain(words),
      sellerShopperId: 'shopper-seller-1',
      askingPriceUsd: randomAskingPriceMicros(),
      endTime,
      auctionTypeId: selected.auctionTypeId,
      listingType: selected.listingType,
      radarVisible: true,
    });
  };

  const typeSelect = (
    <select
      value={typeIndex}
      onChange={(e) => setTypeIndex(Number(e.target.value))}
      className="rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-2 py-1.5 text-sm"
    >
      {BIN_INVENTORY_TYPES.map((t, i) => (
        <option key={`${t.auctionTypeId}-${t.label}`} value={i}>
          {t.label} ({t.auctionTypeId} → {t.itc})
        </option>
      ))}
    </select>
  );

  if (variant === 'expanded') {
    return (
      <div className="flex flex-col items-center gap-4 rounded-xl border-2 border-dashed border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-950 py-12 px-6">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">No BIN domains yet</h3>
        <p className="text-sm text-gray-500 dark:text-gray-400 text-center max-w-md">
          Generate one of each BIN / closeout / OCO type on top of the existing auctions, or
          create a single listing of a chosen inventory type below.
        </p>
        <button
          type="button"
          onClick={onGenerate}
          disabled={isGenerating}
          className="rounded-md bg-indigo-600 px-5 py-2 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
        >
          {isGenerating ? 'Generating...' : 'Generate BIN domains'}
        </button>
        <div className="flex items-end gap-2">
          <div className="flex flex-col gap-1">
            <label className="text-xs font-medium text-gray-500 dark:text-gray-400">Inventory type</label>
            {typeSelect}
          </div>
          <button
            type="button"
            onClick={handleCreate}
            disabled={isCreating || !wordsReady}
            className="rounded-md bg-gray-700 px-3 py-1.5 text-sm font-semibold text-white hover:bg-gray-600 disabled:opacity-50"
          >
            {isCreating ? 'Creating...' : 'Create one'}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex items-end gap-2">
      <div className="flex flex-col gap-1">
        <label className="text-xs font-medium text-gray-500 dark:text-gray-400">Inventory type</label>
        {typeSelect}
      </div>
      <div className="flex flex-col gap-1">
        <label className="text-xs font-medium text-gray-500 dark:text-gray-400">Duration (min)</label>
        <input
          type="number"
          min={1}
          value={minutes}
          onChange={(e) => setMinutes(Math.max(1, Number(e.target.value)))}
          className="w-20 rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-2 py-1.5 text-sm tabular-nums"
        />
      </div>
      <button
        type="button"
        onClick={handleCreate}
        disabled={isCreating || !wordsReady}
        className="rounded-md bg-gray-700 px-3 py-1.5 text-sm font-semibold text-white hover:bg-gray-600 disabled:opacity-50"
      >
        {isCreating ? 'Creating...' : 'Create'}
      </button>
      <button
        type="button"
        onClick={onGenerate}
        disabled={isGenerating}
        className="rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
      >
        {isGenerating ? 'Generating...' : 'Generate all'}
      </button>
    </div>
  );
}
