import { useEffect, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useUseCase } from './useDiContainer';
import { AutoGenerateListingsUseCase } from '../../domain/usecases/AutoGenerateListingsUseCase';
import { useAutoGeneratorStore } from '../stores/autoGeneratorStore';
import { usdToMicros } from '../../domain/entities/Price';

export function useAutoGeneratorViewModel() {
  const queryClient = useQueryClient();
  const autoGenerate = useUseCase(AutoGenerateListingsUseCase);
  const { isRunning, config, generatedCount, setRunning, incrementCount, reset } = useAutoGeneratorStore();
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const startTimeRef = useRef<number>(0);
  const seqRef = useRef(0);

  const start = () => {
    seqRef.current = generatedCount;
    startTimeRef.current = Date.now();
    setRunning(true);
  };

  const stop = () => {
    setRunning(false);
  };

  useEffect(() => {
    if (!isRunning) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      return;
    }

    const tick = async () => {
      if (config.durationMs > 0 && Date.now() - startTimeRef.current >= config.durationMs) {
        stop();
        return;
      }
      seqRef.current++;
      try {
        const result = await autoGenerate.execute(
          {
            sellerShopperId: config.sellerShopperId,
            domainPattern: config.domainPattern,
            minPriceMicros: usdToMicros(config.minPriceUsd),
            maxPriceMicros: usdToMicros(config.maxPriceUsd),
            autoExtEnabled: config.autoExtEnabled,
            autoExtWindowSec: config.autoExtWindowSec,
            autoExtSeconds: config.autoExtSeconds,
            endTimeOffsetMinutes: config.endTimeOffsetMinutes,
          },
          seqRef.current,
        );
        incrementCount(result.domainName);
        queryClient.invalidateQueries({ queryKey: ['listings'] });
      } catch {
        // Error already handled by axios interceptor toast
      }
    };

    tick();
    intervalRef.current = setInterval(tick, config.intervalMs);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [isRunning]);

  return { isRunning, start, stop, reset };
}
