export type TimeMode = 'realtime' | 'offset' | 'frozen';

export interface TimeResponse {
  serverTime: string;
  mode: TimeMode;
}

export type TimeUpdate =
  | { offsetSeconds: number }
  | { freezeAt: string }
  | { reset: true };
