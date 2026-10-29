export const views = [
  'overview',
  'cpu',
  'memory',
  'disks',
  'network',
  'sensors',
  'processes',
  'system',
] as const

export type View = (typeof views)[number] | 'settings'

const fromHash = (): View => {
  const name = location.hash.slice(1)
  return name === 'settings' || views.includes(name as never) ? (name as View) : 'overview'
}

export const nav = $state({ view: fromHash(), seenIssues: 0 })

addEventListener('hashchange', () => (nav.view = fromHash()))

export function go(view: View) {
  nav.view = view
  location.hash = view
}

export function step(by: number) {
  const at = views.indexOf(nav.view as never)
  go(views[(Math.max(at, 0) + by + views.length) % views.length])
}
