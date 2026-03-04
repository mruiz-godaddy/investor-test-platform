import { z } from 'zod';
import { configSnapshotSchema } from './configSchema';

export const scenarioResultSchema = z.object({
  scenario: z.string(),
  description: z.string(),
  config: configSnapshotSchema,
  shoppers: z.array(z.object({ shopperId: z.string(), memberId: z.number() })),
  listings: z.array(z.object({
    listingId: z.number(),
    domainName: z.string(),
    endTime: z.string(),
  }).passthrough()),
  bidsPlaced: z.number(),
});
