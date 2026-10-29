import type { NIC, Snapshot } from './types'

export const series = (history: Snapshot[], pick: (s: Snapshot) => number | undefined) =>
  history.map((s) => pick(s) ?? 0)

export const visibleNics = (nics: NIC[] | undefined, showVirtual: boolean) =>
  (nics ?? []).filter((n) => showVirtual || !n.virtual)

export const sum = (values: number[]) => values.reduce((a, b) => a + b, 0)
