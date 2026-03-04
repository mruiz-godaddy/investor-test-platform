import type { ConfigSnapshot, ConfigUpdate } from '../entities/ServerConfig';

export interface IConfigRepository {
  get(): Promise<ConfigSnapshot>;
  update(config: ConfigUpdate): Promise<ConfigSnapshot>;
}
