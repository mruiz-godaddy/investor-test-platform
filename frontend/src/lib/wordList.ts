const TLD_POOL = ['.com', '.net', '.org', '.co', '.info', '.tv', '.us', '.cc', '.ws', '.biz', '.io'];

/** Parse a JSONL file where each line is a JSON-encoded string. */
export function parseWordListJsonl(text: string): string[] {
  return text
    .trim()
    .split('\n')
    .filter(Boolean)
    .map((line) => JSON.parse(line) as string);
}

/** Return all words starting with the given letter (case-insensitive). */
export function getWordsForLetter(words: string[], letter: string): string[] {
  const lower = letter.toLowerCase();
  return words.filter((w) => w[0] === lower);
}

export function pickRandom<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)];
}

export function randomInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

export function randomTld(): string {
  return pickRandom(TLD_POOL);
}

/**
 * Build a domain name from the word list.
 * If `letter` is provided, prefer words starting with that letter.
 * Falls back to the full list when no words match the letter (e.g. 'x').
 */
export function buildDomain(words: string[], letter?: string): string {
  const pool = letter ? getWordsForLetter(words, letter) : words;
  const candidates = pool.length > 0 ? pool : words;
  const word = pickRandom(candidates);
  const suffix = randomInt(10, 99);
  return `${word}${suffix}${randomTld()}`;
}

let cachedWords: string[] | null = null;

/** Fetch and cache the word list from the public JSONL file. */
export async function fetchWordList(): Promise<string[]> {
  if (cachedWords) return cachedWords;
  const res = await fetch('/data/english-words.jsonl');
  const text = await res.text();
  cachedWords = parseWordListJsonl(text);
  return cachedWords;
}

/** Clear the in-memory cache (useful for tests). */
export function clearWordListCache(): void {
  cachedWords = null;
}
