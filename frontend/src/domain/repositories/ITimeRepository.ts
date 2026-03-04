import type { TimeResponse, TimeUpdate } from '../entities/ServerTime';

export interface ITimeRepository {
  get(): Promise<TimeResponse>;
  update(update: TimeUpdate): Promise<TimeResponse>;
}
