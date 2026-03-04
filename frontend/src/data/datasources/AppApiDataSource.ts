import { injectable } from 'tsyringe';
import api from '../../lib/axios';
import { appListingSchema } from '../schemas/listingSchema';
import { bidResultSchema } from '../schemas/bidResultSchema';

@injectable()
export class AppApiDataSource {
  async getListing(listingId: number, shopperId: string) {
    const { data } = await api.get(
      `/v1/aftermarket/domains/listings/${listingId}`,
      { params: { shopperId } },
    );
    return appListingSchema.parse(data);
  }

  async placeBid(listingId: number, shopperId: string, body: { usdBidAmount: number; isTosAccepted: boolean }) {
    const { data } = await api.post(
      `/v1/aftermarket/domains/listings/${listingId}/bids`,
      body,
      { params: { shopperId } },
    );
    return bidResultSchema.parse(data);
  }

  async getBiddingListings(shopperId: string) {
    const { data } = await api.get('/v1/aftermarket/domains/bidding', {
      params: { shopperId },
    });
    return data;
  }
}
