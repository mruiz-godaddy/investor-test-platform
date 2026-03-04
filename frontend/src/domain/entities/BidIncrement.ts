const MU = 1_000_000;

export const BID_INCREMENT_TABLE: { maxDollars: number; incrementMicros: number }[] = [
  { maxDollars: 499, incrementMicros: 5 * MU },
  { maxDollars: 999, incrementMicros: 10 * MU },
  { maxDollars: 2499, incrementMicros: 25 * MU },
  { maxDollars: 4999, incrementMicros: 50 * MU },
  { maxDollars: 9999, incrementMicros: 100 * MU },
  { maxDollars: 24999, incrementMicros: 250 * MU },
  { maxDollars: 49999, incrementMicros: 500 * MU },
];

const DEFAULT_INCREMENT = 1000 * MU;

export function getBidIncrement(currentPriceMicros: number): number {
  const priceDollars = Math.floor(currentPriceMicros / MU);
  for (const entry of BID_INCREMENT_TABLE) {
    if (priceDollars <= entry.maxDollars) {
      return entry.incrementMicros;
    }
  }
  return DEFAULT_INCREMENT;
}
