import { z } from 'zod';

export const configSnapshotSchema = z.object({
  autoFinalize: z.boolean(),
  statusTransitionDelayMs: z.number(),
  finalizerIntervalMs: z.number(),
});
