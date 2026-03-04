import type { ConfigSnapshot } from './ServerConfig';

export const ScenarioName = {
  NORMAL_AUCTION: 'normal_auction',
  SNIPER_BID: 'sniper_bid',
  RACE_CONDITION: 'race_condition',
  AUTO_EXTEND_CHAIN: 'auto_extend_chain',
  RESERVE_NOT_MET: 'reserve_not_met',
  DELAYED_TRANSITION: 'delayed_transition',
  PROXY_OUTBID: 'proxy_outbid',
  PROXY_STACK: 'proxy_stack',
  PROXY_BURN: 'proxy_burn',
} as const;
export type ScenarioName = (typeof ScenarioName)[keyof typeof ScenarioName];

export interface ScenarioResult {
  scenario: string;
  description: string;
  config: ConfigSnapshot;
  shoppers: { shopperId: string; memberId: number }[];
  listings: { listingId: number; domainName: string; endTime: string; [key: string]: unknown }[];
  bidsPlaced: number;
}

export interface ScenarioMetadata {
  name: ScenarioName;
  description: string;
  tags: string[];
}

export const SCENARIOS: ScenarioMetadata[] = [
  {
    name: ScenarioName.NORMAL_AUCTION,
    description: 'Standard auction ending on time, transitions to SOLD',
    tags: ['basic', 'finalization'],
  },
  {
    name: ScenarioName.SNIPER_BID,
    description: 'Sniper bid triggers auto-extension within last 60s',
    tags: ['auto-extend', 'sniper'],
  },
  {
    name: ScenarioName.RACE_CONDITION,
    description: 'Reproduces CreditTip.com race condition — autoFinalize disabled, sniper can bid after endTime',
    tags: ['race', 'critical'],
  },
  {
    name: ScenarioName.AUTO_EXTEND_CHAIN,
    description: 'Multiple snipers trigger chained auto-extensions (120s each)',
    tags: ['auto-extend', 'chain'],
  },
  {
    name: ScenarioName.RESERVE_NOT_MET,
    description: 'Auction ends with bids below reserve price — transitions to CLOSED',
    tags: ['reserve', 'finalization'],
  },
  {
    name: ScenarioName.DELAYED_TRANSITION,
    description: 'Listing stays OPEN for 10s after endTime, then transitions to SOLD',
    tags: ['delay', 'finalization'],
  },
  {
    name: ScenarioName.PROXY_OUTBID,
    description: "Proxy holder auto-outbids a competitor — buyer-A's $20 proxy auto-outbids buyer-B's $10 bid",
    tags: ['proxy', 'outbid'],
  },
  {
    name: ScenarioName.PROXY_STACK,
    description: 'Same customer raises their proxy — old proxy cancelled, new PROXY($50), bidsCount stays 1',
    tags: ['proxy', 'stack'],
  },
  {
    name: ScenarioName.PROXY_BURN,
    description: "Incoming bid exceeds and burns proxy — buyer-B's $15 beats buyer-A's $12 proxy",
    tags: ['proxy', 'burn'],
  },
];
