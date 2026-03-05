import { create } from 'zustand';

export interface AutoGenStoreConfig {
  intervalMs: number;
  durationMs: number;
  domainPattern: string;
  minPriceUsd: number;
  maxPriceUsd: number;
  sellerShopperId: string;
  autoExtEnabled: boolean;
  autoExtWindowSec: number;
  autoExtSeconds: number;
  endTimeOffsetMinutes: number;
}

interface AutoGeneratorState {
  isRunning: boolean;
  config: AutoGenStoreConfig;
  generatedCount: number;
  recentDomains: string[];
  setRunning: (running: boolean) => void;
  setConfig: (config: Partial<AutoGenStoreConfig>) => void;
  incrementCount: (domain: string) => void;
  reset: () => void;
}

const DEFAULT_CONFIG: AutoGenStoreConfig = {
  intervalMs: 5000,
  durationMs: 60000,
  domainPattern: 'auto-{n}-{r}',
  minPriceUsd: 5,
  maxPriceUsd: 100,
  sellerShopperId: 'shopper-seller-1',
  autoExtEnabled: true,
  autoExtWindowSec: 60,
  autoExtSeconds: 300,
  endTimeOffsetMinutes: 5,
};

export const useAutoGeneratorStore = create<AutoGeneratorState>((set) => ({
  isRunning: false,
  config: DEFAULT_CONFIG,
  generatedCount: 0,
  recentDomains: [],
  setRunning: (isRunning) => set({ isRunning }),
  setConfig: (partial) => set((state) => ({ config: { ...state.config, ...partial } })),
  incrementCount: (domain) =>
    set((state) => ({
      generatedCount: state.generatedCount + 1,
      recentDomains: [domain, ...state.recentDomains].slice(0, 50),
    })),
  reset: () => set({ generatedCount: 0, recentDomains: [], isRunning: false }),
}));
