export interface ConfigSnapshot {
  autoFinalize: boolean;
  statusTransitionDelayMs: number;
  finalizerIntervalMs: number;
  autoExtWindowSec: number;
  autoExtSeconds: number;
  includeBin: boolean;
}

export interface ConfigUpdate {
  autoFinalize?: boolean;
  statusTransitionDelayMs?: number;
  finalizerIntervalMs?: number;
  autoExtWindowSec?: number;
  autoExtSeconds?: number;
  includeBin?: boolean;
}
