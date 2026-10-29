import { tick } from 'svelte'
import { go, nav, step, views } from './nav.svelte'
import { sortBy, table } from './procs.svelte'
import { settings, themes } from './settings.svelte'
import { stream } from './stream.svelte'

interface Binding {
  keys: string[]
  label: string
  run?: () => void
  only?: 'processes'
}

function cycleTheme() {
  const names = themes.map(([name]) => name)
  settings.theme = names[(names.indexOf(settings.theme) + 1) % names.length]
}

function move(by: number) {
  const rows = [...document.querySelectorAll<HTMLElement>('[data-pid]')]
  const at = rows.findIndex((r) => Number(r.dataset.pid) === table.selected)
  const next = rows[Math.min(Math.max(at + by, 0), rows.length - 1)]
  if (!next) return
  table.selected = Number(next.dataset.pid)
  next.scrollIntoView({ block: 'nearest' })
}

export const bindings: Binding[] = [
  { keys: ['1-8'], label: 'jump to a tab' },
  { keys: ['←', '→', 'h', 'l'], label: 'previous / next tab' },
  { keys: ['/'], label: 'filter processes' },
  { keys: ['p', 'Space'], label: 'pause / resume', run: () => stream.toggle() },
  { keys: ['t'], label: 'next theme', run: cycleTheme },
  { keys: ['s', ','], label: 'settings', run: () => go(nav.view === 'settings' ? 'overview' : 'settings') },
  { keys: ['?'], label: 'keys and help', run: () => go('settings') },
  { keys: ['Esc'], label: 'clear the filter / close settings' },
  { keys: ['j', 'k', '↓', '↑'], label: 'move through processes', only: 'processes' },
  { keys: ['c', 'm', 'i', 'n', 'u'], label: 'sort by cpu, mem, pid, name, user', only: 'processes' },
  { keys: ['r'], label: 'reverse the sort', run: () => (table.desc = !table.desc), only: 'processes' },
  { keys: ['T'], label: 'tree view', run: () => (table.tree = !table.tree), only: 'processes' },
  { keys: ['K'], label: 'show kernel threads', run: () => (settings.showKernel = !settings.showKernel), only: 'processes' },
  { keys: ['a'], label: 'show every process', run: () => (table.all = !table.all), only: 'processes' },
]

const sortKeys: Record<string, Parameters<typeof sortBy>[0]> = {
  c: 'cpu', m: 'mem', i: 'pid', n: 'name', u: 'user',
}

const typing = (el: EventTarget | null) =>
  el instanceof HTMLElement && (el.isContentEditable || /^(INPUT|SELECT|TEXTAREA)$/.test(el.tagName))

export function onKey(e: KeyboardEvent) {
  if (e.ctrlKey || e.metaKey || e.altKey) return
  if (e.key === 'Escape') {
    if (table.filter) table.filter = ''
    else if (nav.view === 'settings') go('overview')
    ;(document.activeElement as HTMLElement | null)?.blur()
    return
  }
  if (typing(e.target)) return
  const pressing = e.target instanceof HTMLElement && e.target.closest('button, a, summary')
  if (pressing && (e.key === ' ' || e.key === 'Enter')) return

  const key = e.key === ' ' ? 'Space' : e.key
  const procs = nav.view === 'processes'
  const digit = Number(key)
  if (digit >= 1 && digit <= views.length) go(views[digit - 1])
  else if (key === 'ArrowRight' || key === 'l') step(1)
  else if (key === 'ArrowLeft' || key === 'h') step(-1)
  else if (key === '/') {
    go('processes')
    tick().then(() => document.getElementById('proc-filter')?.focus())
  } else if (procs && (key === 'j' || key === 'ArrowDown')) move(1)
  else if (procs && (key === 'k' || key === 'ArrowUp')) move(-1)
  else if (nav.view === 'processes' && sortKeys[key]) sortBy(sortKeys[key])
  else {
    const hit = bindings.find(
      (b) => b.keys.includes(key) && (!b.only || b.only === nav.view),
    )
    if (!hit?.run) return
    hit.run()
  }
  e.preventDefault()
}
