import type { ConfigSnapshot } from '../../domain/entities/ServerConfig';
import type { z } from 'zod';
import type { configSnapshotSchema } from '../schemas/configSchema';

type ConfigDto = z.infer<typeof configSnapshotSchema>;

export function mapConfig(dto: ConfigDto): ConfigSnapshot {
  return {
    autoFinalize: dto.autoFinalize,
    statusTransitionDelayMs: dto.statusTransitionDelayMs,
    finalizerIntervalMs: dto.finalizerIntervalMs,
    autoExtWindowSec: dto.autoExtWindowSec,
    autoExtSeconds: dto.autoExtSeconds,
    includeBin: dto.includeBin,
  };
}
