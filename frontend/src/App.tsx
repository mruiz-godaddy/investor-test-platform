import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import AppLayout from './presentation/layouts/AppLayout';
import DashboardPage from './presentation/pages/DashboardPage';
import ListingsPage from './presentation/pages/ListingsPage';
import ListingDetailPage from './presentation/pages/ListingDetailPage';
import ShoppersPage from './presentation/pages/ShoppersPage';
import ShopperDetailPage from './presentation/pages/ShopperDetailPage';
import ScenariosPage from './presentation/pages/ScenariosPage';
import SettingsPage from './presentation/pages/SettingsPage';
import AutoGeneratorPage from './presentation/pages/AutoGeneratorPage';
import { ROUTES } from './presentation/router/routes';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 1000,
    },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route element={<AppLayout />}>
            <Route path={ROUTES.DASHBOARD} element={<DashboardPage />} />
            <Route path={ROUTES.LISTINGS} element={<ListingsPage />} />
            <Route path={ROUTES.LISTING_DETAIL} element={<ListingDetailPage />} />
            <Route path={ROUTES.SHOPPERS} element={<ShoppersPage />} />
            <Route path={ROUTES.SHOPPER_DETAIL} element={<ShopperDetailPage />} />
            <Route path={ROUTES.SCENARIOS} element={<ScenariosPage />} />
            <Route path={ROUTES.SETTINGS} element={<SettingsPage />} />
            <Route path={ROUTES.AUTO_GENERATOR} element={<AutoGeneratorPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
