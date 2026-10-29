import type { Proc } from './types'

export const sortKeys = ['cpu', 'mem', 'pid', 'name', 'user', 'threads'] as const
export type SortKey = (typeof sortKeys)[number]

export const table = $state({
  sort: 'cpu' as SortKey,
  desc: true,
  filter: '',
  selected: null as number | null,
  tree: false,
  all: false,
})

export function sortBy(key: SortKey) {
  if (table.sort === key) table.desc = !table.desc
  else Object.assign(table, { sort: key, desc: key !== 'name' && key !== 'user' })
}

const compare = (a: Proc, b: Proc) => {
  const x = a[table.sort]
  const y = b[table.sort]
  const order = typeof x === 'string' ? x.localeCompare(y as string) : (x as number) - (y as number)
  return (table.desc ? -order : order) || a.pid - b.pid
}

export type Row = Proc & { depth: number }

// builds the visible rows, as a flat sorted list or as a tree with children under parents
export function rows(list: Proc[], showKernel: boolean): Row[] {
  const needle = table.filter.trim().toLowerCase()
  const visible = list.filter(
    (p) =>
      (showKernel || !p.kernel) &&
      (!needle || `${p.pid} ${p.name} ${p.user} ${p.cmd ?? ''}`.toLowerCase().includes(needle)),
  )
  if (!table.tree || needle) return visible.sort(compare).map((p) => ({ ...p, depth: 0 }))

  const pids = new Set(visible.map((p) => p.pid))
  const children = new Map<number, Proc[]>()
  for (const p of visible) {
    const parent = pids.has(p.ppid) && p.ppid !== p.pid ? p.ppid : 0
    children.set(parent, [...(children.get(parent) ?? []), p])
  }
  const out: Row[] = []
  const walk = (parent: number, depth: number) => {
    for (const p of (children.get(parent) ?? []).sort(compare)) {
      out.push({ ...p, depth })
      walk(p.pid, depth + 1)
    }
  }
  walk(0, 0)
  return out
}
