/**
 * Deterministic pseudo-randomness for the mock backend.
 *
 * The same wallet always produces the same history and the same transactions,
 * so screenshots, tests and manual QA stay reproducible.
 */

export function hashString(input: string): number {
  let hash = 2166136261
  for (let index = 0; index < input.length; index += 1) {
    hash ^= input.charCodeAt(index)
    hash = Math.imul(hash, 16777619)
  }
  return hash >>> 0
}

/** Mulberry32 — small, fast, stable across runs. */
export function createRandom(seed: number): () => number {
  let state = seed >>> 0
  return () => {
    state = (state + 0x6d2b79f5) >>> 0
    let t = state
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

export function pick<T>(random: () => number, items: readonly T[]): T {
  const index = Math.floor(random() * items.length)
  return items[Math.min(index, items.length - 1)] as T
}

/** Uniform float in `[min, max)`, rendered with `decimals` places. */
export function randomDecimal(
  random: () => number,
  min: number,
  max: number,
  decimals: number,
): string {
  return (min + random() * (max - min)).toFixed(decimals)
}
