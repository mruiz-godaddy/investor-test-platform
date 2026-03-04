import { formatMicrosUsd } from '../../../lib/formatters';

interface Props {
  micros: number;
  className?: string;
}

export default function PriceDisplay({ micros, className }: Props) {
  return <span className={className}>{formatMicrosUsd(micros)}</span>;
}
