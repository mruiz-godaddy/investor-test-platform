import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useUseCase } from './useDiContainer';
import { GetConfigUseCase } from '../../domain/usecases/GetConfigUseCase';
import { UpdateConfigUseCase } from '../../domain/usecases/UpdateConfigUseCase';
import { GetTimeUseCase } from '../../domain/usecases/GetTimeUseCase';
import { UpdateTimeUseCase } from '../../domain/usecases/UpdateTimeUseCase';
import { WipeDatabaseUseCase } from '../../domain/usecases/WipeDatabaseUseCase';
import { ExportDatabaseUseCase } from '../../domain/usecases/ExportDatabaseUseCase';
import { ImportDatabaseUseCase } from '../../domain/usecases/ImportDatabaseUseCase';
import type { ConfigUpdate } from '../../domain/entities/ServerConfig';
import type { TimeUpdate } from '../../domain/entities/ServerTime';
import { useToastStore } from '../stores/toastStore';

export function useSettingsViewModel() {
  const queryClient = useQueryClient();
  const addToast = useToastStore((s) => s.addToast);
  const getConfig = useUseCase(GetConfigUseCase);
  const updateConfig = useUseCase(UpdateConfigUseCase);
  const getTime = useUseCase(GetTimeUseCase);
  const updateTime = useUseCase(UpdateTimeUseCase);
  const wipeDatabase = useUseCase(WipeDatabaseUseCase);
  const exportDatabase = useUseCase(ExportDatabaseUseCase);
  const importDatabase = useUseCase(ImportDatabaseUseCase);

  const { data: config } = useQuery({
    queryKey: ['config'],
    queryFn: () => getConfig.execute(),
  });

  const { data: time } = useQuery({
    queryKey: ['time'],
    queryFn: () => getTime.execute(),
    refetchInterval: 1000,
  });

  const configMutation = useMutation({
    mutationFn: (update: ConfigUpdate) => updateConfig.execute(update),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config'] });
      addToast({ type: 'success', message: 'Configuration saved' });
    },
  });

  const timeMutation = useMutation({
    mutationFn: (update: TimeUpdate) => updateTime.execute(update),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['time'] }),
  });

  const wipeMutation = useMutation({
    mutationFn: () => wipeDatabase.execute(),
    onSuccess: () => {
      queryClient.invalidateQueries();
      addToast({ type: 'success', message: 'Database wiped' });
    },
  });

  const exportMutation = useMutation({
    mutationFn: () => exportDatabase.execute(),
    onSuccess: (data) => {
      const json = JSON.stringify(data, null, 2);
      const blob = new Blob([json], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `biddings-db-${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.json`;
      a.click();
      URL.revokeObjectURL(url);
      addToast({ type: 'success', message: `Exported ${data.shoppers.length} shoppers, ${data.listings.length} listings, ${data.bids.length} bids` });
    },
  });

  const importMutation = useMutation({
    mutationFn: async (file: File) => {
      const text = await file.text();
      const parsed = JSON.parse(text);
      return importDatabase.execute({
        shoppers: parsed.shoppers ?? [],
        listings: parsed.listings ?? [],
        bids: parsed.bids ?? [],
      });
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries();
      addToast({ type: 'success', message: `Imported ${data.shoppers} shoppers, ${data.listings} listings, ${data.bids} bids` });
    },
  });

  return {
    config,
    time,
    updateConfig: configMutation.mutate,
    isUpdatingConfig: configMutation.isPending,
    updateTime: timeMutation.mutate,
    wipeDatabase: wipeMutation.mutate,
    isWiping: wipeMutation.isPending,
    exportDatabase: () => exportMutation.mutate(),
    isExporting: exportMutation.isPending,
    importDatabase: importMutation.mutate,
    isImporting: importMutation.isPending,
  };
}
