const iec = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB']
const si = ['B', 'kB', 'MB', 'GB', 'TB', 'PB']

export function bytes(n: number, binary = true): string {
  const base = binary ? 1024 : 1000
  const units = binary ? iec : si
  let i = 0
  while (Math.abs(n) >= base && i < units.length - 1) {
    n /= base
    i++
  }
  const digits = i === 0 || n >= 100 ? 0 : 1
  return `${n.toFixed(digits)} ${units[i]}`
}

export function rate(n: number, binary = true, bits = false): string {
  if (!bits) return `${bytes(n, binary)}/s`
  return `${bytes(n * 8, false).replace('B', 'b')}ps`
}

export const pct = (n: number) => `${n < 10 && n > 0 ? n.toFixed(1) : Math.round(n)}%`

export function duration(seconds: number): string {
  const s = Math.max(0, Math.floor(seconds))
  const parts: [number, string][] = [
    [Math.floor(s / 86400), 'd'],
    [Math.floor((s % 86400) / 3600), 'h'],
    [Math.floor((s % 3600) / 60), 'm'],
    [s % 60, 's'],
  ]
  const shown = parts.filter(([n]) => n > 0).slice(0, 2)
  return shown.length ? shown.map(([n, u]) => `${n}${u}`).join(' ') : '0s'
}

export const temp = (c: number, unit: 'c' | 'f') =>
  unit === 'f' ? `${Math.round((c * 9) / 5 + 32)}°F` : `${Math.round(c)}°C`

export const ghz = (mhz: number) => `${(mhz / 1000).toFixed(2)} GHz`
