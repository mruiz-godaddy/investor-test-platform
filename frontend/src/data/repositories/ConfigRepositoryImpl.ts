import { injectable } from 'tsyringe';
import type { IConfigRepository } from '../../domain/repositories/IConfigRepository';
import type { ConfigSnapshot, ConfigUpdate } from '../../domain/entities/ServerConfig';
import { AdminApiDataSource } from '../datasources/AdminApiDataSource';
import { mapConfig } from '../mappers/ConfigMapper';

@injectable()
export class ConfigRepositoryImpl implements IConfigRepository {
  constructor(private ds: AdminApiDataSource) {}

  async get(): Promise<ConfigSnapshot> {
    const dto = await this.ds.getConfig();
    return mapConfig(dto);
  }

  async update(config: ConfigUpdate): Promise<ConfigSnapshot> {
    const dto = await this.ds.updateConfig(config as Record<string, unknown>);
    return mapConfig(dto);
  }
}
