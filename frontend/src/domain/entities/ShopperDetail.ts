import type { Shopper } from './Shopper';
import type { ShopperBid } from './ShopperBid';

export interface ShopperDetail extends Shopper {
  bidHistory: ShopperBid[];
}
