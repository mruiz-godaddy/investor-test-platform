import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useUseCase } from './useDiContainer';
import { GetShoppersUseCase } from '../../domain/usecases/GetShoppersUseCase';
import { CreateShopperUseCase } from '../../domain/usecases/CreateShopperUseCase';

export function useShoppersViewModel() {
  const queryClient = useQueryClient();
  const getShoppers = useUseCase(GetShoppersUseCase);
  const createShopper = useUseCase(CreateShopperUseCase);

  const { data: shoppers = [], isLoading } = useQuery({
    queryKey: ['shoppers'],
    queryFn: () => getShoppers.execute(),
  });

  const mutation = useMutation({
    mutationFn: (req: { shopperId: string; memberId: number; customerId?: string; displayName?: string }) =>
      createShopper.execute(req),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['shoppers'] }),
  });

  return { shoppers, isLoading, createShopper: mutation.mutate };
}
