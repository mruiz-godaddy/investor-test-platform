import type { ShopperDetail } from '../../domain/entities/ShopperDetail';
import type { ShopperBid } from '../../domain/entities/ShopperBid';
import type { BidType, BidStatus } from '../../domain/entities/Bid';
import type { z } from 'zod';
import type { shopperDetailSchema, shopperBidSchema } from '../schemas/shopperDetailSchema';

type ShopperDetailDto = z.infer<typeof shopperDetailSchema>;
type ShopperBidDto = z.infer<typeof shopperBidSchema>;

export function mapShopperBid(dto: ShopperBidDto): ShopperBid {
  return {
    bidId: dto.bidId,
    listingId: dto.listingId,
    shopperId: dto.shopperId,
    bidAmountUsd: dto.bidAmountUsd,
    bidType: dto.bidType as BidType,
    bidStatus: dto.bidStatus as BidStatus,
    isHighBid: dto.isHighBid,
    parentBidId: dto.parentBidId,
    createdAt: dto.createdAt,
    domainName: dto.domainName,
    listingStatus: dto.listingStatus,
    highestBidderShopper: dto.highestBidderShopper,
  };
}

export function mapShopperDetail(dto: ShopperDetailDto): ShopperDetail {
  return {
    shopperId: dto.shopperId,
    memberId: dto.memberId,
    customerId: dto.customerId,
    displayName: dto.displayName,
    bidHistory: dto.bidHistory.map(mapShopperBid),
  };
}
