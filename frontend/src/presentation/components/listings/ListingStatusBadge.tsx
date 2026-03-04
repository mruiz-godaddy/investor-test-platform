import type { ListingStatus } from '../../../domain/entities/Listing';

const STATUS_STYLES: Record<string, string> = {
  OPEN: 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200',
  SOLD: 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200',
  CLOSED: 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200',
};

interface Props {
  status: ListingStatus;
}

export default function ListingStatusBadge({ status }: Props) {
  return (
    <span className={`inline-flex rounded-full px-2 py-1 text-xs font-semibold ${STATUS_STYLES[status] ?? 'bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-100'}`}>
      {status}
    </span>
  );
}
