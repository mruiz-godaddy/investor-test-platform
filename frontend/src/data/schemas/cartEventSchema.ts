import { z } from 'zod';

export const cartEventSchema = z.object({
  eventId: z.number(),
  domainName: z.string(),
  listingId: z.number(),
  inventoryType: z.number(),
  itcCode: z.string(),
  itcInventory: z.string(),
  area: z.string(),
  requestPrice: z.number(),
  createdAt: z.string(),
});

export const cartEventArraySchema = z.array(cartEventSchema);
