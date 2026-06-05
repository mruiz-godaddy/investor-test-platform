import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useUseCase } from './useDiContainer';
import { GetListingsUseCase } from '../../domain/usecases/GetListingsUseCase';
import { CreateListingUseCase } from '../../domain/usecases/CreateListingUseCase';
import { UpdateListingStatusUseCase } from '../../domain/usecases/UpdateListingStatusUseCase';
import { UpdateEndTimeUseCase } from '../../domain/usecases/UpdateEndTimeUseCase';
import { PlaceSniperBidUseCase } from '../../domain/usecases/PlaceSniperBidUseCase';
import { UpdateRadarVisibleUseCase } from '../../domain/usecases/UpdateRadarVisibleUseCase';
import { GetShoppersUseCase } from '../../domain/usecases/GetShoppersUseCase';
import { SetupSystemUseCase } from '../../domain/usecases/SetupSystemUseCase';
import type { CreateListingRequest, ListingStatus } from '../../domain/entities/Listing';
import { LISTINGS_POLL_INTERVAL_MS } from '../../lib/constants';
import { useToastStore } from '../stores/toastStore';

export function useListingsViewModel() {
  const queryClient = useQueryClient();
  const addToast = useToastStore((s) => s.addToast);
  const getListings = useUseCase(GetListingsUseCase);
  const createListing = useUseCase(CreateListingUseCase);
  const updateStatus = useUseCase(UpdateListingStatusUseCase);
  const updateEndTime = useUseCase(UpdateEndTimeUseCase);
  const placeSniperBid = useUseCase(PlaceSniperBidUseCase);
  const updateRadarVisible = useUseCase(UpdateRadarVisibleUseCase);
  const getShoppers = useUseCase(GetShoppersUseCase);
  const setupSystem = useUseCase(SetupSystemUseCase);

  const { data: listings = [], isLoading } = useQuery({
    queryKey: ['listings'],
    queryFn: () => getListings.execute(),
    refetchInterval: LISTINGS_POLL_INTERVAL_MS,
  });

  const { data: shoppers = [] } = useQuery({
    queryKey: ['shoppers'],
    queryFn: () => getShoppers.execute(),
  });

  const invalidateAll = () => queryClient.invalidateQueries();

  const createMutation = useMutation({
    mutationFn: (req: CreateListingRequest) => createListing.execute(req),
    onSuccess: invalidateAll,
  });

  const setupMutation = useMutation({
    mutationFn: (opts: { durationMinutes?: number; appShopperId?: string } = {}) =>
      setupSystem.execute(opts.durationMinutes, opts.appShopperId),
    onSuccess: (data) => {
      invalidateAll();
      addToast({ type: 'success', message: `System ready: ${data.shoppers} shoppers + ${data.listings} listings` });
    },
  });

  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: ListingStatus }) => updateStatus.execute(id, status),
    onSuccess: invalidateAll,
  });

  const endTimeMutation = useMutation({
    mutationFn: ({ id, update }: { id: number; update: { endTime?: string; addSeconds?: number } }) =>
      updateEndTime.execute(id, update),
    onSuccess: invalidateAll,
  });

  const sniperBidMutation = useMutation({
    mutationFn: ({ id, shopperId, bidAmountUsd }: { id: number; shopperId: string; bidAmountUsd: number }) =>
      placeSniperBid.execute(id, shopperId, bidAmountUsd),
    onSuccess: invalidateAll,
  });

  const radarMutation = useMutation({
    mutationFn: ({ id, radarVisible }: { id: number; radarVisible: boolean }) =>
      updateRadarVisible.execute(id, radarVisible),
    onSuccess: invalidateAll,
  });

  return {
    listings,
    shoppers,
    isLoading,
    createListing: createMutation.mutate,
    setupSystem: setupMutation.mutateAsync,
    isSettingUp: setupMutation.isPending,
    isCreating: createMutation.isPending,
    updateStatus: statusMutation.mutate,
    updateEndTime: endTimeMutation.mutate,
    placeSniperBid: sniperBidMutation.mutate,
    toggleRadar: radarMutation.mutate,
  };
}
