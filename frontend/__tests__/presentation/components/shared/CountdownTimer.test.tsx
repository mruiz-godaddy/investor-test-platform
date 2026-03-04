import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import CountdownTimer from '../../../../src/presentation/components/listings/CountdownTimer';

describe('CountdownTimer', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2024-01-01T00:00:00Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows formatted countdown', () => {
    // 1 hour and 30 minutes from now
    const endTime = new Date('2024-01-01T01:30:00Z').toISOString();
    render(<CountdownTimer endTime={endTime} />);
    expect(screen.getByText('01:30:00')).toBeDefined();
  });

  it('shows Expired for past times', () => {
    const endTime = new Date('2023-12-31T23:00:00Z').toISOString();
    render(<CountdownTimer endTime={endTime} />);
    expect(screen.getByText('Expired')).toBeDefined();
  });
});
