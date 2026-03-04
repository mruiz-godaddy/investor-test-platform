import type { ListingStatus } from '../../../domain/entities/Listing';

const STATUS_STYLES: Record<string, string> = {
  OPEN: 'bg-green-100 text-green-800',
  SOLD: 'bg-blue-100 text-blue-800',
  CLOSED: 'bg-red-100 text-red-800',
};

interface Props {
  status: ListingStatus;
}

export default function ListingStatusBadge({ status }: Props) {
  return (
    <span className={`inline-flex rounded-full px-2 py-1 text-xs font-semibold ${STATUS_STYLES[status] ?? 'bg-gray-100 text-gray-800'}`}>
      {status}
    </span>
  );
}
