import type { AdminBid, BidType, BidStatus } from '../../domain/entities/Bid';
import type { z } from 'zod';
import type { adminBidSchema } from '../schemas/bidSchema';

type AdminBidDto = z.infer<typeof adminBidSchema>;

export function mapAdminBid(dto: AdminBidDto): AdminBid {
  return {
    bidId: dto.bidId,
    shopperId: dto.shopperId,
    bidAmountUsd: dto.bidAmountUsd,
    bidType: dto.bidType as BidType,
    bidStatus: dto.bidStatus as BidStatus,
    isHighBid: dto.isHighBid,
    parentBidId: dto.parentBidId,
    createdAt: dto.createdAt,
  };
}
