import { log } from './log'

export const themes = [
  ['auto', 'auto (peng / latte)'],
  ['peng', 'peng'],
  ['catppuccin-mocha', 'catppuccin mocha'],
  ['catppuccin-latte', 'catppuccin latte'],
  ['dracula', 'dracula'],
  ['gruvbox', 'gruvbox'],
  ['nord', 'nord'],
  ['tokyo-night', 'tokyo night'],
  ['one-dark', 'one dark'],
  ['rose-pine', 'rosé pine'],
  ['solarized-light', 'solarized light'],
] as const

export type Theme = (typeof themes)[number][0]
export const overviewPanels = ['cpu', 'memory', 'network', 'disks', 'sensors', 'processes'] as const
export type OverviewPanel = (typeof overviewPanels)[number]

const defaults = {
  theme: 'auto' as Theme,
  size: 'm' as 's' | 'm' | 'l',
  corners: 'square' as 'square' | 'round',
  graph: 'dots' as 'dots' | 'line',
  binary: true,
  bits: false,
  tempUnit: 'c' as 'c' | 'f',
  cpuPerCore: true,
  showKernel: false,
  showVirtual: false,
  rows: 100,
  overview: Object.fromEntries(overviewPanels.map((p) => [p, true])) as Record<OverviewPanel, boolean>,
}

export type Settings = typeof defaults
const key = 'peng-top:settings'

function load(): Settings {
  try {
    const saved = JSON.parse(localStorage.getItem(key) ?? '{}')
    return { ...defaults, ...saved, overview: { ...defaults.overview, ...saved.overview } }
  } catch (err) {
    log('warn', "couldn't read saved settings, using the defaults", err)
    return structuredClone(defaults)
  }
}

export const settings = $state(load())

export function reset() {
  Object.assign(settings, structuredClone(defaults))
}

$effect.root(() => {
  $effect(() => {
    const json = JSON.stringify(settings)
    try {
      localStorage.setItem(key, json)
    } catch (err) {
      log('debug', "settings won't stick in this browser", err)
    }
  })
})
