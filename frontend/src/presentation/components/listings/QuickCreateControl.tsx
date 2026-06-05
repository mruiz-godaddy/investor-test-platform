import { useState } from 'react';
import type { CreateListingRequest } from '../../../domain/entities/Listing';
import type { ConfigSnapshot } from '../../../domain/entities/ServerConfig';
import { usdToMicros } from '../../../domain/entities/Price';
import { buildDomain, randomInt, fetchWordList } from '../../../lib/wordList';
import { useEffect } from 'react';

type TimeUnit = 'seconds' | 'minutes' | 'hours' | 'days';

const UNIT_TO_MS: Record<TimeUnit, number> = {
  seconds: 1_000,
  minutes: 60_000,
  hours: 3_600_000,
  days: 86_400_000,
};

const UNIT_MIN: Record<TimeUnit, number> = {
  seconds: 1,
  minutes: 1,
  hours: 1,
  days: 1,
};

const ASKING_PRICE_OPTIONS = [5, 10, 15, 20, 25, 50, 75, 100];

function randomAskingPriceMicros(): number {
  return usdToMicros(ASKING_PRICE_OPTIONS[randomInt(0, ASKING_PRICE_OPTIONS.length - 1)]);
}

export function buildRandomListing(
  domain: string,
  durationMs: number,
  autoExtEnabled: boolean,
  config?: ConfigSnapshot,
  radarVisible?: boolean,
): CreateListingRequest {
  const endTime = new Date(Date.now() + durationMs).toISOString();
  const askingPriceUsd = randomAskingPriceMicros();

  return {
    domainName: domain,
    sellerShopperId: 'shopper-seller-1',
    askingPriceUsd,
    endTime,
    autoExtEnabled,
    autoExtWindowSec: autoExtEnabled ? (config?.autoExtWindowSec ?? 60) : undefined,
    autoExtSeconds: autoExtEnabled ? (config?.autoExtSeconds ?? 300) : undefined,
    radarVisible: radarVisible || undefined,
  };
}

function useWordList(): string[] {
  const [words, setWords] = useState<string[]>([]);
  useEffect(() => {
    fetchWordList().then(setWords);
  }, []);
  return words;
}

// --- Single quick-create (inline header bar) ---

interface InlineProps {
  onSubmit: (req: CreateListingRequest) => void;
  isPending?: boolean;
  variant?: 'inline';
  config?: ConfigSnapshot;
}

// --- System setup (empty state) ---

interface ExpandedProps {
  onSetup: (opts: { durationMinutes: number; appShopperId?: string }) => void;
  isPending?: boolean;
  variant: 'expanded';
}

// App shopper id persists across setups (prefilled until the user changes it).
const APP_SHOPPER_ID_KEY = 'emulatorAppShopperId';

type Props = InlineProps | ExpandedProps;

export default function QuickCreateControl(props: Props) {
  const [amount, setAmount] = useState(5);
  const [unit, setUnit] = useState<TimeUnit>('minutes');
  const [autoExtEnabled, setAutoExtEnabled] = useState(false);
  const [radarVisible, setRadarVisible] = useState(false);
  const [setupMinutes, setSetupMinutes] = useState(5);
  const [appShopperId, setAppShopperId] = useState(() => localStorage.getItem(APP_SHOPPER_ID_KEY) ?? '');
  const words = useWordList();
  const wordsReady = words.length > 0;

  if (props.variant === 'expanded') {
    const { onSetup, isPending } = props;
    const setupMinsValid = Number.isFinite(setupMinutes) && setupMinutes >= 1;
    return (
      <div className="flex flex-col items-center gap-4 rounded-xl border-2 border-dashed border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-950 py-12 px-6">
        <svg className="h-12 w-12 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
        </svg>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">No biddings yet</h3>
        <p className="text-sm text-gray-500 dark:text-gray-400 text-center max-w-md">
          Set up the system with 8 shoppers (3 sellers + 5 buyers) and 26 auctions (A-Z),
          each with a unique domain. All auctions use the duration below (default 5 minutes).
        </p>
        <div className="flex flex-col items-center gap-1">
          <label htmlFor="setup-duration" className="text-xs font-medium text-gray-500 dark:text-gray-400">
            Auction duration (minutes)
          </label>
          <input
            id="setup-duration"
            type="number"
            min={1}
            step={1}
            value={setupMinutes}
            onChange={(e) => setSetupMinutes(Math.max(1, Math.floor(Number(e.target.value))))}
            className="w-28 rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-2 py-1.5 text-sm text-center tabular-nums"
          />
        </div>
        <div className="flex flex-col items-center gap-1">
          <label htmlFor="setup-app-shopper" className="text-xs font-medium text-gray-500 dark:text-gray-400">
            App shopper ID (for Won/Lost) — optional
          </label>
          <input
            id="setup-app-shopper"
            type="text"
            placeholder="e.g. 1251418"
            value={appShopperId}
            onChange={(e) => setAppShopperId(e.target.value)}
            className="w-56 rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-2 py-1.5 text-sm text-center"
          />
        </div>
        <button
          type="button"
          onClick={() => {
            const trimmed = appShopperId.trim();
            // Persist so it prefills next time; the user can still change it.
            if (trimmed) localStorage.setItem(APP_SHOPPER_ID_KEY, trimmed);
            else localStorage.removeItem(APP_SHOPPER_ID_KEY);
            onSetup({ durationMinutes: setupMinsValid ? setupMinutes : 5, appShopperId: trimmed || undefined });
          }}
          disabled={isPending || !setupMinsValid}
          className="rounded-md bg-indigo-600 px-5 py-2 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
        >
          {isPending ? 'Setting up...' : 'Setup System'}
        </button>
        <p className="text-xs text-gray-400 dark:text-gray-500">
          8 shoppers + 26 listings · all end in {setupMinsValid ? setupMinutes : 5} min
          {appShopperId.trim() ? ` · won/lost → ${appShopperId.trim()}` : ''}
        </p>
      </div>
    );
  }

  const { onSubmit, isPending, config: configProp } = props;
  const minVal = UNIT_MIN[unit];

  const handleClick = () => {
    const clamped = Math.max(amount, minVal);
    const durationMs = clamped * UNIT_TO_MS[unit];
    const domain = buildDomain(words);
    onSubmit(buildRandomListing(domain, durationMs, autoExtEnabled, configProp, radarVisible));
  };

  return (
    <div className="flex items-end gap-2">
      <div>
        <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Duration</label>
        <div className="flex gap-1">
          <input
            type="number"
            min={minVal}
            value={amount}
            onChange={(e) => setAmount(Math.max(minVal, Number(e.target.value)))}
            className="w-20 rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-2 py-1.5 text-sm tabular-nums"
          />
          <select
            value={unit}
            onChange={(e) => {
              const next = e.target.value as TimeUnit;
              setUnit(next);
              setAmount(Math.max(UNIT_MIN[next], amount));
            }}
            className="rounded-md border border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white px-2 py-1.5 text-sm"
          >
            <option value="seconds">sec</option>
            <option value="minutes">min</option>
            <option value="hours">hrs</option>
            <option value="days">days</option>
          </select>
        </div>
      </div>
      <div className="flex items-center gap-1.5 pb-0.5">
        <input
          type="checkbox"
          checked={autoExtEnabled}
          onChange={(e) => setAutoExtEnabled(e.target.checked)}
          className="h-3.5 w-3.5 rounded border-gray-300 dark:border-gray-600 dark:bg-gray-800 text-indigo-600"
        />
        <label className="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">Auto Ext</label>
      </div>
      <div className="flex items-center gap-1.5 pb-0.5">
        <input
          type="checkbox"
          checked={radarVisible}
          onChange={(e) => setRadarVisible(e.target.checked)}
          className="h-3.5 w-3.5 rounded border-gray-300 dark:border-gray-600 dark:bg-gray-800 text-indigo-600"
        />
        <label className="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">Radar</label>
      </div>
      <button
        type="button"
        onClick={handleClick}
        disabled={isPending || !wordsReady}
        className="rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
      >
        {isPending ? 'Creating...' : 'Quick Create'}
      </button>
    </div>
  );
}
