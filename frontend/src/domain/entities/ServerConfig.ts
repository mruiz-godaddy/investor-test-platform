export interface ConfigSnapshot {
  autoFinalize: boolean;
  statusTransitionDelayMs: number;
  finalizerIntervalMs: number;
  autoExtWindowSec: number;
  autoExtSeconds: number;
}

export interface ConfigUpdate {
  autoFinalize?: boolean;
  statusTransitionDelayMs?: number;
  finalizerIntervalMs?: number;
  autoExtWindowSec?: number;
  autoExtSeconds?: number;
}
