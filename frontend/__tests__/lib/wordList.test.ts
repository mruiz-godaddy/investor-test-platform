import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  parseWordListJsonl,
  getWordsForLetter,
  pickRandom,
  randomInt,
  randomTld,
  buildDomain,
  fetchWordList,
  clearWordListCache,
} from '../../src/lib/wordList';

const VALID_TLDS = ['.com', '.net', '.org', '.co', '.info', '.tv', '.us', '.cc', '.ws', '.biz', '.io'];

// --- parseWordListJsonl ---

describe('parseWordListJsonl', () => {
  it('parses JSONL text into string array', () => {
    const text = '"apple"\n"banana"\n"cherry"';
    expect(parseWordListJsonl(text)).toEqual(['apple', 'banana', 'cherry']);
  });

  it('handles trailing newline', () => {
    const text = '"apple"\n"banana"\n';
    expect(parseWordListJsonl(text)).toEqual(['apple', 'banana']);
  });

  it('handles empty string', () => {
    expect(parseWordListJsonl('')).toEqual([]);
  });

  it('skips blank lines', () => {
    const text = '"apple"\n\n"banana"\n\n';
    expect(parseWordListJsonl(text)).toEqual(['apple', 'banana']);
  });

  it('handles single word', () => {
    expect(parseWordListJsonl('"hello"')).toEqual(['hello']);
  });

  it('preserves word casing', () => {
    expect(parseWordListJsonl('"Hello"\n"WORLD"')).toEqual(['Hello', 'WORLD']);
  });
});

// --- getWordsForLetter ---

describe('getWordsForLetter', () => {
  const words = ['apple', 'able', 'banana', 'bold', 'cat', 'dog'];

  it('filters words by letter (lowercase input)', () => {
    expect(getWordsForLetter(words, 'a')).toEqual(['apple', 'able']);
  });

  it('filters words by letter (uppercase input)', () => {
    expect(getWordsForLetter(words, 'B')).toEqual(['banana', 'bold']);
  });

  it('returns empty array for letter with no matches', () => {
    expect(getWordsForLetter(words, 'z')).toEqual([]);
  });

  it('returns empty array for empty word list', () => {
    expect(getWordsForLetter([], 'a')).toEqual([]);
  });
});

// --- pickRandom ---

describe('pickRandom', () => {
  it('returns an element from the array', () => {
    const arr = [1, 2, 3, 4, 5];
    for (let i = 0; i < 50; i++) {
      expect(arr).toContain(pickRandom(arr));
    }
  });

  it('returns the only element for single-item array', () => {
    expect(pickRandom([42])).toBe(42);
  });
});

// --- randomInt ---

describe('randomInt', () => {
  it('returns values within range (inclusive)', () => {
    for (let i = 0; i < 100; i++) {
      const val = randomInt(5, 10);
      expect(val).toBeGreaterThanOrEqual(5);
      expect(val).toBeLessThanOrEqual(10);
    }
  });

  it('returns exact value when min equals max', () => {
    expect(randomInt(7, 7)).toBe(7);
  });

  it('returns integers only', () => {
    for (let i = 0; i < 50; i++) {
      const val = randomInt(1, 100);
      expect(Number.isInteger(val)).toBe(true);
    }
  });
});

// --- randomTld ---

describe('randomTld', () => {
  it('returns a TLD supported by the app', () => {
    for (let i = 0; i < 50; i++) {
      expect(VALID_TLDS).toContain(randomTld());
    }
  });
});

// --- buildDomain ---

describe('buildDomain', () => {
  const words = ['apple', 'able', 'banana', 'bold', 'cat'];

  it('returns domain matching word + 2-digit suffix + TLD', () => {
    const domain = buildDomain(words);
    expect(domain).toMatch(/^[a-z]+\d{2}\.[a-z]+$/);
  });

  it('uses a word starting with the specified letter', () => {
    for (let i = 0; i < 30; i++) {
      const domain = buildDomain(words, 'B');
      expect(domain).toMatch(/^b/);
    }
  });

  it('falls back to full list when letter has no matches', () => {
    const domain = buildDomain(words, 'Z');
    // Should still produce a valid domain
    expect(domain).toMatch(/^[a-z]+\d{2}\.[a-z]+$/);
  });

  it('suffix is between 10 and 99', () => {
    for (let i = 0; i < 50; i++) {
      const domain = buildDomain(words);
      const match = domain.match(/(\d+)/);
      expect(match).not.toBeNull();
      const num = Number(match![1]);
      expect(num).toBeGreaterThanOrEqual(10);
      expect(num).toBeLessThanOrEqual(99);
    }
  });

  it('TLD is from the supported list', () => {
    for (let i = 0; i < 50; i++) {
      const domain = buildDomain(words);
      const tld = '.' + domain.split('.').pop();
      expect(VALID_TLDS).toContain(tld);
    }
  });
});

// --- fetchWordList ---

describe('fetchWordList', () => {
  beforeEach(() => {
    clearWordListCache();
    vi.restoreAllMocks();
  });

  it('fetches and parses the JSONL file', async () => {
    const mockResponse = '"alpha"\n"beta"\n"gamma"';
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({ text: () => Promise.resolve(mockResponse) }),
    );

    const words = await fetchWordList();
    expect(words).toEqual(['alpha', 'beta', 'gamma']);
    expect(fetch).toHaveBeenCalledWith('/data/english-words.jsonl');
  });

  it('caches the result on subsequent calls', async () => {
    const mockResponse = '"alpha"\n"beta"';
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({ text: () => Promise.resolve(mockResponse) }),
    );

    await fetchWordList();
    await fetchWordList();
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  it('re-fetches after clearing cache', async () => {
    const mockResponse = '"alpha"';
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({ text: () => Promise.resolve(mockResponse) }),
    );

    await fetchWordList();
    clearWordListCache();
    await fetchWordList();
    expect(fetch).toHaveBeenCalledTimes(2);
  });
});
