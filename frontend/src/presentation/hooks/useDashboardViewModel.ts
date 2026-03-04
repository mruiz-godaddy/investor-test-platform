import { useQuery } from '@tanstack/react-query';
import { useUseCase } from './useDiContainer';
import { GetListingsUseCase } from '../../domain/usecases/GetListingsUseCase';
import { LISTINGS_POLL_INTERVAL_MS } from '../../lib/constants';

export function useDashboardViewModel() {
  const getListings = useUseCase(GetListingsUseCase);

  const { data: listings = [], isLoading, error } = useQuery({
    queryKey: ['listings'],
    queryFn: () => getListings.execute(),
    refetchInterval: LISTINGS_POLL_INTERVAL_MS,
  });

  const totalCount = listings.length;
  const openCount = listings.filter((l) => l.listingStatus === 'OPEN').length;
  const soldCount = listings.filter((l) => l.listingStatus === 'SOLD').length;
  const closedCount = listings.filter((l) => l.listingStatus === 'CLOSED').length;

  return { listings, isLoading, error, totalCount, openCount, soldCount, closedCount };
}
