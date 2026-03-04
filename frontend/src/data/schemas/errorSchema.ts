import { z } from 'zod';

export const adminErrorSchema = z.object({
  error: z.string(),
});

export const appErrorSchema = z.object({
  code: z.string(),
  message: z.string(),
});
