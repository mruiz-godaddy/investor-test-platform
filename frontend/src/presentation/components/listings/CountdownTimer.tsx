import { useState, useEffect } from 'react';
import { formatCountdown } from '../../../lib/formatters';

interface Props {
  endTime: string;
  className?: string;
}

export default function CountdownTimer({ endTime, className }: Props) {
  const [diffMs, setDiffMs] = useState(() => new Date(endTime).getTime() - Date.now());

  useEffect(() => {
    const timer = setInterval(() => {
      setDiffMs(new Date(endTime).getTime() - Date.now());
    }, 1000);
    return () => clearInterval(timer);
  }, [endTime]);

  const isExpired = diffMs <= 0;
  const isUrgent = diffMs > 0 && diffMs < 60_000;

  return (
    <span
      className={`font-mono text-sm ${
        isExpired ? 'text-gray-400' : isUrgent ? 'text-red-600 font-bold' : 'text-gray-900'
      } ${className ?? ''}`}
    >
      {isExpired ? 'Expired' : formatCountdown(diffMs)}
    </span>
  );
}
