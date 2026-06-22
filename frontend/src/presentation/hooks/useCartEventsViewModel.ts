import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useUseCase } from './useDiContainer';
import { GetCartEventsUseCase } from '../../domain/usecases/GetCartEventsUseCase';
import { ClearCartEventsUseCase } from '../../domain/usecases/ClearCartEventsUseCase';
import { LISTINGS_POLL_INTERVAL_MS } from '../../lib/constants';
import { useToastStore } from '../stores/toastStore';

export function useCartEventsViewModel() {
  const queryClient = useQueryClient();
  const addToast = useToastStore((s) => s.addToast);
  const getCartEvents = useUseCase(GetCartEventsUseCase);
  const clearCartEvents = useUseCase(ClearCartEventsUseCase);

  const { data: events = [], isLoading } = useQuery({
    queryKey: ['cart-events'],
    queryFn: () => getCartEvents.execute(),
    refetchInterval: LISTINGS_POLL_INTERVAL_MS,
  });

  const clearMutation = useMutation({
    mutationFn: () => clearCartEvents.execute(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cart-events'] });
      addToast({ type: 'success', message: 'Captured ITC codes cleared' });
    },
  });

  return {
    events,
    isLoading,
    clear: clearMutation.mutate,
    isClearing: clearMutation.isPending,
  };
}
