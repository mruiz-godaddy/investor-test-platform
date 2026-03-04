import { container } from 'tsyringe';
import { TOKENS } from './tokens';

import { ListingRepositoryImpl } from '../data/repositories/ListingRepositoryImpl';
import { ShopperRepositoryImpl } from '../data/repositories/ShopperRepositoryImpl';
import { ConfigRepositoryImpl } from '../data/repositories/ConfigRepositoryImpl';
import { TimeRepositoryImpl } from '../data/repositories/TimeRepositoryImpl';
import { ScenarioRepositoryImpl } from '../data/repositories/ScenarioRepositoryImpl';
import { AdminRepositoryImpl } from '../data/repositories/AdminRepositoryImpl';
import { AppRepositoryImpl } from '../data/repositories/AppRepositoryImpl';

container.register(TOKENS.IListingRepository, { useClass: ListingRepositoryImpl });
container.register(TOKENS.IShopperRepository, { useClass: ShopperRepositoryImpl });
container.register(TOKENS.IConfigRepository, { useClass: ConfigRepositoryImpl });
container.register(TOKENS.ITimeRepository, { useClass: TimeRepositoryImpl });
container.register(TOKENS.IScenarioRepository, { useClass: ScenarioRepositoryImpl });
container.register(TOKENS.IAdminRepository, { useClass: AdminRepositoryImpl });
container.register(TOKENS.IAppRepository, { useClass: AppRepositoryImpl });

export { container };
