import { injectable } from 'tsyringe';
import type { ITimeRepository } from '../../domain/repositories/ITimeRepository';
import type { TimeResponse, TimeUpdate } from '../../domain/entities/ServerTime';
import { AdminApiDataSource } from '../datasources/AdminApiDataSource';
import { mapTime } from '../mappers/TimeMapper';

@injectable()
export class TimeRepositoryImpl implements ITimeRepository {
  constructor(private ds: AdminApiDataSource) {}

  async get(): Promise<TimeResponse> {
    const dto = await this.ds.getTime();
    return mapTime(dto);
  }

  async update(update: TimeUpdate): Promise<TimeResponse> {
    const dto = await this.ds.updateTime(update as Record<string, unknown>);
    return mapTime(dto);
  }
}
