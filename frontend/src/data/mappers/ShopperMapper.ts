import type { Shopper } from '../../domain/entities/Shopper';
import type { z } from 'zod';
import type { shopperSchema } from '../schemas/shopperSchema';

type ShopperDto = z.infer<typeof shopperSchema>;

export function mapShopper(dto: ShopperDto): Shopper {
  return {
    shopperId: dto.shopperId,
    memberId: dto.memberId,
    customerId: dto.customerId,
    displayName: dto.displayName,
  };
}
