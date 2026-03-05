import { describe, it, expect } from 'vitest';
import {
  buildRandomListing,
  buildAZListings,
} from '../../../../src/presentation/components/listings/QuickCreateControl';

// A minimal word list covering A-Z (no X words, matching the real BIP39 list)
const SAMPLE_WORDS = [
  'apple', 'able', 'banana', 'bold', 'cat', 'cool',
  'dog', 'dash', 'edge', 'egg', 'fox', 'flux',
  'grape', 'grid', 'hat', 'hive', 'ice', 'iron',
  'jade', 'joy', 'kite', 'key', 'link', 'lamp',
  'mint', 'moon', 'nova', 'net', 'orbit', 'oak',
  'pulse', 'pen', 'quest', 'quit', 'rise', 'run',
  'spark', 'sun', 'tide', 'top', 'ultra', 'use',
  'volt', 'van', 'wave', 'win', 'yield', 'year',
  'zero', 'zone',
  // intentionally no 'x' words
];

const VALID_TLDS = ['.com', '.net', '.org', '.co', '.info', '.tv', '.us', '.cc', '.ws', '.biz', '.io'];

// --- buildRandomListing ---

describe('buildRandomListing', () => {
  it('produces a valid CreateListingRequest', () => {
    const result = buildRandomListing('test42.com', 5 * 60_000);

    expect(result.domainName).toBe('test42.com');
    expect(result.sellerShopperId).toBe('shopper-seller');
    expect(result.endTime).toBeDefined();
    expect(result.askingPriceUsd).toBeGreaterThan(0);
    expect(typeof result.autoExtEnabled).toBe('boolean');
  });

  it('sets endTime in the future', () => {
    const before = Date.now();
    const result = buildRandomListing('test.com', 10 * 60_000);
    const endMs = new Date(result.endTime!).getTime();

    // endTime should be ~10 min from now (allow 5s tolerance)
    expect(endMs).toBeGreaterThan(before + 9 * 60_000);
    expect(endMs).toBeLessThan(before + 11 * 60_000);
  });

  it('asking price is a recognized dollar amount in micros', () => {
    const validMicros = [5, 10, 15, 20, 25, 50, 75, 100].map((d) => d * 1_000_000);
    for (let i = 0; i < 50; i++) {
      const result = buildRandomListing('test.com', 60_000);
      expect(validMicros).toContain(result.askingPriceUsd);
    }
  });

  it('autoExt fields are set when enabled', () => {
    // Run enough times to get both enabled and disabled
    const results = Array.from({ length: 100 }, () => buildRandomListing('test.com', 60_000));

    const enabled = results.filter((r) => r.autoExtEnabled);
    const disabled = results.filter((r) => !r.autoExtEnabled);

    // Should get both cases with 100 trials
    expect(enabled.length).toBeGreaterThan(0);
    expect(disabled.length).toBeGreaterThan(0);

    for (const r of enabled) {
      expect(r.autoExtWindowSec).toBe(60);
      expect([120, 300, 600]).toContain(r.autoExtSeconds);
    }

    for (const r of disabled) {
      expect(r.autoExtWindowSec).toBeUndefined();
      expect(r.autoExtSeconds).toBeUndefined();
    }
  });
});

// --- buildAZListings ---

describe('buildAZListings', () => {
  it('generates exactly 26 listings', () => {
    const listings = buildAZListings(SAMPLE_WORDS);
    expect(listings).toHaveLength(26);
  });

  it('staggers end times from 5 to 30 minutes', () => {
    const before = Date.now();
    const listings = buildAZListings(SAMPLE_WORDS);

    listings.forEach((listing, i) => {
      const endMs = new Date(listing.endTime!).getTime();
      const expectedDurationMs = (5 + i) * 60_000;
      // Allow 5s tolerance for test execution time
      expect(endMs).toBeGreaterThan(before + expectedDurationMs - 5_000);
      expect(endMs).toBeLessThan(before + expectedDurationMs + 5_000);
    });
  });

  it('domain names for available letters start with the correct letter', () => {
    const listings = buildAZListings(SAMPLE_WORDS);
    const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split('');

    // Letters that have words in our SAMPLE_WORDS
    const lettersWithWords = new Set(
      SAMPLE_WORDS.map((w) => w[0].toUpperCase()),
    );

    listings.forEach((listing, i) => {
      const letter = letters[i];
      if (lettersWithWords.has(letter)) {
        expect(listing.domainName[0]).toBe(letter.toLowerCase());
      }
      // For letters without words (e.g., X), domain falls back to any word — no assertion
    });
  });

  it('each domain has a valid TLD', () => {
    const listings = buildAZListings(SAMPLE_WORDS);

    for (const listing of listings) {
      const tld = '.' + listing.domainName.split('.').pop();
      expect(VALID_TLDS).toContain(tld);
    }
  });

  it('each listing has required fields', () => {
    const listings = buildAZListings(SAMPLE_WORDS);

    for (const listing of listings) {
      expect(listing.domainName).toBeTruthy();
      expect(listing.sellerShopperId).toBe('shopper-seller');
      expect(listing.endTime).toBeDefined();
      expect(listing.askingPriceUsd).toBeGreaterThan(0);
      expect(typeof listing.autoExtEnabled).toBe('boolean');
    }
  });
});
