import type { Shopper } from '../../../domain/entities/Shopper';
import EmptyState from '../shared/EmptyState';

interface Props {
  shoppers: Shopper[];
  onRowClick?: (shopper: Shopper) => void;
}

export default function ShoppersTable({ shoppers, onRowClick }: Props) {
  if (shoppers.length === 0) {
    return <EmptyState title="No shoppers" description="Create a shopper to get started." />;
  }

  return (
    <div className="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
      <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
        <thead className="bg-gray-50 dark:bg-gray-950">
          <tr>
            {['Shopper ID', 'Member ID', 'Customer ID', 'Display Name'].map((h) => (
              <th key={h} className="px-4 py-3 text-left text-base font-medium uppercase text-gray-500 dark:text-gray-400">{h}</th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200 dark:divide-gray-700 bg-white dark:bg-gray-900">
          {shoppers.map((s) => (
            <tr
              key={s.shopperId}
              onClick={() => onRowClick?.(s)}
              className={onRowClick ? 'cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800' : ''}
            >
              <td className="px-4 py-3 text-lg font-medium text-gray-900 dark:text-white">{s.shopperId}</td>
              <td className="px-4 py-3 text-lg text-gray-500 dark:text-gray-400">{s.memberId}</td>
              <td className="px-4 py-3 text-lg text-gray-500 dark:text-gray-400">{s.customerId || '-'}</td>
              <td className="px-4 py-3 text-lg text-gray-500 dark:text-gray-400">{s.displayName || '-'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
