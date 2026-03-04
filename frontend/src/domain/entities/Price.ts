const MU_MULTIPLIER = 1_000_000;

export interface PriceArray {
  cost: number;
  currency: string;
}

export function microsToUsd(micros: number): number {
  return micros / MU_MULTIPLIER;
}

export function usdToMicros(usd: number): number {
  return Math.round(usd * MU_MULTIPLIER);
}

export function microsToPriceArray(micros: number): PriceArray[] {
  return [{ cost: micros, currency: 'USD' }];
}

export function priceArrayToMicros(priceArray: PriceArray[]): number {
  if (priceArray.length === 0) return 0;
  return priceArray[0].cost;
}
