import { z } from 'zod';

export const timeResponseSchema = z.object({
  serverTime: z.string(),
  mode: z.enum(['realtime', 'offset', 'frozen']),
});
