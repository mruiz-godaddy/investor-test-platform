import type { TimeResponse, TimeMode } from '../../domain/entities/ServerTime';
import type { z } from 'zod';
import type { timeResponseSchema } from '../schemas/timeSchema';

type TimeDto = z.infer<typeof timeResponseSchema>;

export function mapTime(dto: TimeDto): TimeResponse {
  return {
    serverTime: dto.serverTime,
    mode: dto.mode as TimeMode,
  };
}
