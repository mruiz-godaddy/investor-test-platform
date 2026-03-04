import { format, formatDistanceToNow } from 'date-fns';

const MU = 1_000_000;

export function formatMicrosUsd(micros: number): string {
  const dollars = micros / MU;
  return `$${dollars.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

export function formatDateTime(isoString: string): string {
  return format(new Date(isoString), 'MMM d, yyyy HH:mm:ss');
}

export function formatRelativeTime(isoString: string): string {
  return formatDistanceToNow(new Date(isoString), { addSuffix: true });
}

export function formatCountdown(diffMs: number): string {
  if (diffMs <= 0) return '00:00:00';
  const totalSeconds = Math.floor(diffMs / 1000);
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  return [hours, minutes, seconds].map((v) => String(v).padStart(2, '0')).join(':');
}
