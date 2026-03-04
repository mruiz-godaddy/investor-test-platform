import { useMemo } from 'react';
import { container } from '../../di/container';
import type { InjectionToken } from 'tsyringe';

export function useResolve<T>(token: InjectionToken<T>): T {
  return useMemo(() => container.resolve(token), [token]);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function useUseCase<T>(UseCaseClass: new (...args: any[]) => T): T {
  return useMemo(() => container.resolve(UseCaseClass), [UseCaseClass]);
}
