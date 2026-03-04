export interface ConfigSnapshot {
  autoFinalize: boolean;
  statusTransitionDelayMs: number;
  finalizerIntervalMs: number;
}

export interface ConfigUpdate {
  autoFinalize?: boolean;
  statusTransitionDelayMs?: number;
  finalizerIntervalMs?: number;
}
