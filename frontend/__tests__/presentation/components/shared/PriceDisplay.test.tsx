import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import PriceDisplay from '../../../../src/presentation/components/shared/PriceDisplay';

describe('PriceDisplay', () => {
  it('formats micros to dollar amount', () => {
    render(<PriceDisplay micros={5_000_000} />);
    expect(screen.getByText('$5.00')).toBeDefined();
  });

  it('formats zero', () => {
    render(<PriceDisplay micros={0} />);
    expect(screen.getByText('$0.00')).toBeDefined();
  });

  it('formats large amounts with commas', () => {
    render(<PriceDisplay micros={1_500_000_000} />);
    expect(screen.getByText('$1,500.00')).toBeDefined();
  });
});
