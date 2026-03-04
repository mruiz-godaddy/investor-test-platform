import axios from 'axios';
import { useToastStore } from '../presentation/stores/toastStore';

const api = axios.create({
  baseURL: '',
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (axios.isAxiosError(error) && error.response) {
      const data = error.response.data as Record<string, unknown> | undefined;
      const message =
        (data?.error as string) ??
        (data?.message as string) ??
        error.response.statusText;
      const code = data?.code as string | undefined;

      useToastStore.getState().addToast({ type: 'error', message });

      const enriched = new Error(message) as Error & { code?: string };
      if (code) enriched.code = code;
      return Promise.reject(enriched);
    }
    return Promise.reject(error);
  },
);

export default api;
