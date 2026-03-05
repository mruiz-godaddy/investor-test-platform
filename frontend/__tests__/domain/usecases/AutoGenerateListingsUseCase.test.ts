import 'reflect-metadata';
import { describe, it, expect, vi } from 'vitest';
import { AutoGenerateListingsUseCase } from '../../../src/domain/usecases/AutoGenerateListingsUseCase';
import type { IListingRepository } from '../../../src/domain/repositories/IListingRepository';

describe('AutoGenerateListingsUseCase', () => {
  const mockRepo: IListingRepository = {
    getAll: vi.fn(),
    getById: vi.fn(),
    create: vi.fn().mockResolvedValue({
      listingId: 1,
      domainName: 'generated.com',
      endTime: '2024-01-01T00:05:00Z',
      listingStatus: 'OPEN',
    }),
    updateStatus: vi.fn(),
    updateEndTime: vi.fn(),
    placeSniperBid: vi.fn(),
  };

  it('generates a domain name with sequence number', async () => {
    const uc = new AutoGenerateListingsUseCase(mockRepo);
    await uc.execute(
      {
        sellerShopperId: 'shopper-seller',
        domainPattern: 'test-{n}-{r}',
        minPriceMicros: 5_000_000,
        maxPriceMicros: 10_000_000,
        autoExtEnabled: true,
        autoExtWindowSec: 60,
        autoExtSeconds: 300,
        endTimeOffsetMinutes: 5,
      },
      42,
    );

    expect(mockRepo.create).toHaveBeenCalledTimes(1);
    const call = (mockRepo.create as ReturnType<typeof vi.fn>).mock.calls[0][0];
    expect(call.domainName).toContain('42');
    expect(call.domainName).toMatch(/\.com$/);
    expect(call.sellerShopperId).toBe('shopper-seller');
    expect(call.askingPriceUsd).toBeGreaterThanOrEqual(5_000_000);
    expect(call.askingPriceUsd).toBeLessThanOrEqual(10_000_000);
  });

  it('sets endTime when endTimeOffsetMinutes > 0', async () => {
    const uc = new AutoGenerateListingsUseCase(mockRepo);
    await uc.execute(
      {
        sellerShopperId: 'shopper-seller',
        domainPattern: 'auto-{n}',
        minPriceMicros: 5_000_000,
        maxPriceMicros: 5_000_000,
        autoExtEnabled: false,
        autoExtWindowSec: 60,
        autoExtSeconds: 300,
        endTimeOffsetMinutes: 10,
      },
      1,
    );

    const call = (mockRepo.create as ReturnType<typeof vi.fn>).mock.calls[1][0];
    expect(call.endTime).toBeDefined();
  });
});
