/**
 * Fixed-point decimal arithmetic — USED ONLY BY THE MOCK BACKEND.
 *
 * This module exists because the mock adapter stands in for the Go domain
 * services, which own every financial calculation. It is deliberately confined
 * to `src/api/mock`: nothing in `features/`, `components/` or `hooks/` may
 * import it. Application code formats backend decimals, it never computes them.
 *
 * Values are handled as scaled BigInt to avoid floating-point drift, mirroring
 * PostgreSQL `NUMERIC` / Go `decimal.Decimal`.
 */

const TEN = 10n

function pow10(exponent: number): bigint {
  return TEN ** BigInt(exponent)
}

/** Parses a decimal string into an integer scaled by `scale`, half-up. */
export function toScaled(value: string, scale: number): bigint {
  const trimmed = value.trim()
  const negative = trimmed.startsWith('-')
  const unsigned = negative ? trimmed.slice(1) : trimmed
  const [intPart = '0', fracPart = ''] = unsigned.split('.')

  const padded = `${fracPart}${'0'.repeat(scale + 1)}`
  const kept = padded.slice(0, scale)
  const nextDigit = Number(padded[scale] ?? '0')

  let scaledValue = BigInt(`${intPart || '0'}${kept}`)
  if (nextDigit >= 5) scaledValue += 1n

  return negative ? -scaledValue : scaledValue
}

/** Renders a scaled integer back into a decimal string. */
export function fromScaled(value: bigint, scale: number): string {
  const negative = value < 0n
  const abs = negative ? -value : value
  const divisor = pow10(scale)
  const intPart = abs / divisor
  const fracPart = abs % divisor
  const sign = negative ? '-' : ''

  if (scale === 0) return `${sign}${intPart}`
  return `${sign}${intPart}.${fracPart.toString().padStart(scale, '0')}`
}

/** Integer division rounding half away from zero. */
function divideRound(numerator: bigint, denominator: bigint): bigint {
  const quotient = numerator / denominator
  const remainder = numerator % denominator
  if (remainder === 0n) return quotient

  const absRemainder = remainder < 0n ? -remainder : remainder
  const absDenominator = denominator < 0n ? -denominator : denominator
  if (absRemainder * 2n < absDenominator) return quotient

  const negative = numerator < 0n !== denominator < 0n
  return quotient + (negative ? -1n : 1n)
}

function rescale(value: bigint, from: number, to: number): bigint {
  if (to === from) return value
  if (to > from) return value * pow10(to - from)
  return divideRound(value, pow10(from - to))
}

const WORK_SCALE = 18

/** `a * b`, rounded to `scale` decimal places. */
export function multiply(a: string, b: string, scale = 2): string {
  const product = toScaled(a, WORK_SCALE) * toScaled(b, WORK_SCALE)
  return fromScaled(rescale(product, WORK_SCALE * 2, scale), scale)
}

/** `a / b`, rounded to `scale` decimal places. Returns `null` when `b` is 0. */
export function divide(a: string, b: string, scale = 4): string | null {
  const divisor = toScaled(b, WORK_SCALE)
  if (divisor === 0n) return null
  const numerator = toScaled(a, WORK_SCALE) * pow10(scale)
  return fromScaled(divideRound(numerator, divisor), scale)
}

export function add(values: string[], scale = 2): string {
  const total = values.reduce(
    (sum, value) => sum + toScaled(value, WORK_SCALE),
    0n,
  )
  return fromScaled(rescale(total, WORK_SCALE, scale), scale)
}

export function subtract(a: string, b: string, scale = 2): string {
  const result = toScaled(a, WORK_SCALE) - toScaled(b, WORK_SCALE)
  return fromScaled(rescale(result, WORK_SCALE, scale), scale)
}

/** `part / total * 100`. */
export function percentOf(
  part: string,
  total: string,
  scale = 2,
): string | null {
  const ratio = divide(part, total, WORK_SCALE)
  if (ratio === null) return null
  return multiply(ratio, '100', scale)
}

/** `value * (1 + pct/100)`. */
export function applyPercent(value: string, pct: string, scale = 2): string {
  const factor = add(['1', multiply(pct, '0.01', WORK_SCALE)], WORK_SCALE)
  return multiply(value, factor, scale)
}

export function compare(a: string, b: string): number {
  const left = toScaled(a, WORK_SCALE)
  const right = toScaled(b, WORK_SCALE)
  if (left === right) return 0
  return left > right ? 1 : -1
}

export function isZero(value: string): boolean {
  return toScaled(value, WORK_SCALE) === 0n
}

export function negate(value: string, scale = 2): string {
  return fromScaled(rescale(-toScaled(value, WORK_SCALE), WORK_SCALE, scale), scale)
}
