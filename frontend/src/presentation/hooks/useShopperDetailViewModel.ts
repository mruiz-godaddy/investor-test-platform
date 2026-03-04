import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useUseCase } from './useDiContainer';
import { GetShopperUseCase } from '../../domain/usecases/GetShopperUseCase';
import type { ShopperBid } from '../../domain/entities/ShopperBid';
import { LISTINGS_POLL_INTERVAL_MS } from '../../lib/constants';

export function useShopperDetailViewModel(id: string) {
  const getShopper = useUseCase(GetShopperUseCase);

  const { data: shopper, isLoading } = useQuery({
    queryKey: ['shopper', id],
    queryFn: () => getShopper.execute(id),
    refetchInterval: LISTINGS_POLL_INTERVAL_MS,
  });

  const { activeBids, wonBids, lostBids } = useMemo(() => {
    const active: ShopperBid[] = [];
    const won: ShopperBid[] = [];
    const lost: ShopperBid[] = [];

    if (!shopper) return { activeBids: active, wonBids: won, lostBids: lost };

    for (const bid of shopper.bidHistory) {
      if (bid.bidStatus === 'CANCELLED') {
        lost.push(bid);
      } else if (bid.listingStatus === 'OPEN') {
        active.push(bid);
      } else if (bid.listingStatus === 'SOLD' && bid.highestBidderShopper === shopper.shopperId) {
        won.push(bid);
      } else {
        lost.push(bid);
      }
    }

    return { activeBids: active, wonBids: won, lostBids: lost };
  }, [shopper]);

  return {
    shopper,
    isLoading,
    activeBids,
    wonBids,
    lostBids,
  };
}
