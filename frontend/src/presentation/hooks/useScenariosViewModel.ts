import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useUseCase } from './useDiContainer';
import { LoadScenarioUseCase } from '../../domain/usecases/LoadScenarioUseCase';
import type { ScenarioName, ScenarioResult } from '../../domain/entities/Scenario';

export function useScenariosViewModel() {
  const queryClient = useQueryClient();
  const loadScenario = useUseCase(LoadScenarioUseCase);
  const [result, setResult] = useState<ScenarioResult | null>(null);
  const [loadingName, setLoadingName] = useState<ScenarioName | null>(null);

  const mutation = useMutation({
    mutationFn: (name: ScenarioName) => {
      setLoadingName(name);
      return loadScenario.execute(name);
    },
    onSuccess: (data) => {
      setResult(data);
      setLoadingName(null);
      queryClient.invalidateQueries({ queryKey: ['listings'] });
      queryClient.invalidateQueries({ queryKey: ['shoppers'] });
    },
    onError: () => setLoadingName(null),
  });

  return { result, loadingName, loadScenario: mutation.mutate };
}
