/** Shortens an address or hash for display: `0x71C7…9F2b`. */
export function truncateMiddle(
  value: string,
  options?: { lead?: number; tail?: number },
): string {
  const lead = options?.lead ?? 6
  const tail = options?.tail ?? 4
  if (value.length <= lead + tail + 1) return value
  return `${value.slice(0, lead)}…${value.slice(-tail)}`
}
