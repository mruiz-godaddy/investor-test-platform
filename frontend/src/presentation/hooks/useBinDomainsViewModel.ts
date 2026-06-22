import { useQuery } from '@tanstack/react-query';
import { useUseCase } from './useDiContainer';
import { GetListingsUseCase } from '../../domain/usecases/GetListingsUseCase';
import { LISTINGS_POLL_INTERVAL_MS, BIN_AUCTION_TYPE_IDS } from '../../lib/constants';

export function useBinDomainsViewModel() {
  const getListings = useUseCase(GetListingsUseCase);

  const { data: allListings = [], isLoading, error } = useQuery({
    queryKey: ['listings'],
    queryFn: () => getListings.execute(),
    refetchInterval: LISTINGS_POLL_INTERVAL_MS,
  });

  const listings = allListings.filter((l) => BIN_AUCTION_TYPE_IDS.includes(l.auctionTypeId));

  const totalCount = listings.length;
  const openCount = listings.filter((l) => l.listingStatus === 'OPEN').length;
  const soldCount = listings.filter((l) => l.listingStatus === 'SOLD').length;
  const closedCount = listings.filter((l) => l.listingStatus === 'CLOSED').length;

  return { listings, isLoading, error, totalCount, openCount, soldCount, closedCount };
}
