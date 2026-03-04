import { z } from 'zod';

export const shopperSchema = z.object({
  shopperId: z.string(),
  memberId: z.number(),
  customerId: z.string(),
  displayName: z.string(),
});
