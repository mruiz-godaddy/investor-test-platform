import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IListingRepository } from '../repositories/IListingRepository';
import type { CreateListingRequest, CreateListingResponse } from '../entities/Listing';

export interface AutoGenConfig {
  sellerShopperId: string;
  domainPattern: string;
  minPriceMicros: number;
  maxPriceMicros: number;
  autoExtEnabled: boolean;
  autoExtWindowSec: number;
  autoExtSeconds: number;
  reservePriceMicros: number;
  endTimeOffsetMinutes: number;
}

@injectable()
export class AutoGenerateListingsUseCase {
  constructor(@inject(TOKENS.IListingRepository) private repo: IListingRepository) {}

  execute(config: AutoGenConfig, seq: number): Promise<CreateListingResponse> {
    const domainName = this.generateDomainName(config.domainPattern, seq);
    const askingPriceUsd = this.randomPrice(config.minPriceMicros, config.maxPriceMicros);

    const req: CreateListingRequest = {
      domainName,
      sellerShopperId: config.sellerShopperId,
      askingPriceUsd,
      reservePriceUsd: config.reservePriceMicros,
      autoExtEnabled: config.autoExtEnabled,
      autoExtWindowSec: config.autoExtWindowSec,
      autoExtSeconds: config.autoExtSeconds,
    };

    if (config.endTimeOffsetMinutes > 0) {
      const endTime = new Date(Date.now() + config.endTimeOffsetMinutes * 60_000);
      req.endTime = endTime.toISOString();
    }

    return this.repo.create(req);
  }

  private generateDomainName(pattern: string, seq: number): string {
    const suffix = Math.random().toString(36).substring(2, 6);
    return pattern.replace('{n}', String(seq)).replace('{r}', suffix) + '.com';
  }

  private randomPrice(min: number, max: number): number {
    return Math.floor(Math.random() * (max - min + 1)) + min;
  }
}
