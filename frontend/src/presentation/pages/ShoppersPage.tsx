import { useShoppersViewModel } from '../hooks/useShoppersViewModel';
import ShoppersTable from '../components/shoppers/ShoppersTable';
import CreateShopperForm from '../components/shoppers/CreateShopperForm';
import LoadingSpinner from '../components/shared/LoadingSpinner';

export default function ShoppersPage() {
  const { shoppers, isLoading, createShopper } = useShoppersViewModel();

  if (isLoading) return <LoadingSpinner />;

  return (
    <div>
      <h2 className="text-xl font-bold text-gray-900">Shoppers</h2>
      <div className="mt-4 space-y-6">
        <CreateShopperForm onSubmit={createShopper} />
        <ShoppersTable shoppers={shoppers} />
      </div>
    </div>
  );
}
