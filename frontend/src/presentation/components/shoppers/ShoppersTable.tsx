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
    <div className="overflow-x-auto rounded-lg border border-gray-200">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            {['Shopper ID', 'Member ID', 'Customer ID', 'Display Name'].map((h) => (
              <th key={h} className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">{h}</th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200 bg-white">
          {shoppers.map((s) => (
            <tr
              key={s.shopperId}
              onClick={() => onRowClick?.(s)}
              className={onRowClick ? 'cursor-pointer hover:bg-gray-50' : ''}
            >
              <td className="px-4 py-3 text-sm font-medium text-gray-900">{s.shopperId}</td>
              <td className="px-4 py-3 text-sm text-gray-500">{s.memberId}</td>
              <td className="px-4 py-3 text-sm text-gray-500">{s.customerId || '-'}</td>
              <td className="px-4 py-3 text-sm text-gray-500">{s.displayName || '-'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
