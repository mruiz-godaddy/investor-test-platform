import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useUseCase } from './useDiContainer';
import { GetListingUseCase } from '../../domain/usecases/GetListingUseCase';
import { UpdateListingStatusUseCase } from '../../domain/usecases/UpdateListingStatusUseCase';
import { UpdateEndTimeUseCase } from '../../domain/usecases/UpdateEndTimeUseCase';
import { PlaceSniperBidUseCase } from '../../domain/usecases/PlaceSniperBidUseCase';
import { GetShoppersUseCase } from '../../domain/usecases/GetShoppersUseCase';
import type { ListingStatus } from '../../domain/entities/Listing';
import { LISTINGS_POLL_INTERVAL_MS } from '../../lib/constants';

export function useListingDetailViewModel(id: number) {
  const queryClient = useQueryClient();
  const getListing = useUseCase(GetListingUseCase);
  const updateStatus = useUseCase(UpdateListingStatusUseCase);
  const updateEndTime = useUseCase(UpdateEndTimeUseCase);
  const placeSniperBid = useUseCase(PlaceSniperBidUseCase);
  const getShoppers = useUseCase(GetShoppersUseCase);

  const { data: listing, isLoading } = useQuery({
    queryKey: ['listing', id],
    queryFn: () => getListing.execute(id),
    refetchInterval: LISTINGS_POLL_INTERVAL_MS,
  });

  const { data: shoppers = [] } = useQuery({
    queryKey: ['shoppers'],
    queryFn: () => getShoppers.execute(),
  });

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['listing', id] });
    queryClient.invalidateQueries({ queryKey: ['listings'] });
  };

  const statusMutation = useMutation({
    mutationFn: (status: ListingStatus) => updateStatus.execute(id, status),
    onSuccess: invalidate,
  });

  const endTimeMutation = useMutation({
    mutationFn: (update: { endTime?: string; addSeconds?: number }) => updateEndTime.execute(id, update),
    onSuccess: invalidate,
  });

  const sniperBidMutation = useMutation({
    mutationFn: ({ shopperId, bidAmountUsd }: { shopperId: string; bidAmountUsd: number }) =>
      placeSniperBid.execute(id, shopperId, bidAmountUsd),
    onSuccess: invalidate,
  });

  return {
    listing,
    shoppers,
    isLoading,
    updateStatus: statusMutation.mutate,
    updateEndTime: endTimeMutation.mutate,
    placeSniperBid: sniperBidMutation.mutate,
  };
}
