import type { AdminListing, ListingStatus } from '../../domain/entities/Listing';
import type { z } from 'zod';
import type { adminListingSchema } from '../schemas/listingSchema';
import { mapAdminBid } from './BidMapper';

type AdminListingDto = z.infer<typeof adminListingSchema>;

export function mapAdminListing(dto: AdminListingDto): AdminListing {
  return {
    listingId: dto.listingId,
    domainName: dto.domainName,
    listingStatus: dto.listingStatus as ListingStatus,
    listingType: dto.listingType,
    auctionTypeId: dto.auctionTypeId,
    startTime: dto.startTime,
    endTime: dto.endTime,
    askingPriceUsd: dto.askingPriceUsd,
    currentPriceUsd: dto.currentPriceUsd,
    salePriceUsd: dto.salePriceUsd,
    nextBidPriceUsd: dto.nextBidPriceUsd,
    biddersCount: dto.biddersCount,
    bidsCount: dto.bidsCount,
    isAutoExtended: dto.isAutoExtended,
    sellerShopperId: dto.sellerShopperId,
    highestBidderShopper: dto.highestBidderShopper,
    autoExtEnabled: dto.autoExtEnabled,
    autoExtWindowSec: dto.autoExtWindowSec,
    autoExtSeconds: dto.autoExtSeconds,
    createdAt: dto.createdAt,
    bidHistory: dto.bidHistory.map(mapAdminBid),
  };
}
