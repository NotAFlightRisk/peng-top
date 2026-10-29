<script lang="ts">
  import type { Component } from 'svelte'
  import { onKey } from './lib/keys'
  import { nav, views, type View } from './lib/nav.svelte'
  import { settings } from './lib/settings.svelte'
  import { stream } from './lib/stream.svelte'
  import { duration, pct } from './lib/format'
  import Overview from './views/Overview.svelte'
  import Cpu from './views/Cpu.svelte'
  import Memory from './views/Memory.svelte'
  import Disks from './views/Disks.svelte'
  import Network from './views/Network.svelte'
  import Sensors from './views/Sensors.svelte'
  import Processes from './views/Processes.svelte'
  import System from './views/System.svelte'
  import Settings from './views/Settings.svelte'

  const screens: Record<View, Component<any>> = {
    overview: Overview,
    cpu: Cpu,
    memory: Memory,
    disks: Disks,
    network: Network,
    sensors: Sensors,
    processes: Processes,
    system: System,
    settings: Settings,
  }
  const Screen = $derived(screens[nav.view])

  const modeLabel = { host: 'host', container: 'container', 'host-via-container': 'host via container' }
  const statusLabel = { connecting: 'connecting', live: 'live', paused: 'paused', lost: 'reconnecting' }

  let prefersDark = $state(matchMedia('(prefers-color-scheme: dark)').matches)
  let now = $state(Date.now())
  const info = $derived(stream.info)
  const issues = $derived(Object.keys(info?.issues ?? {}).length)

  stream.start()
  matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => (prefersDark = e.matches))
  setInterval(() => (now = Date.now()), 1000)

  $effect(() => {
    const root = document.documentElement
    root.dataset.theme =
      settings.theme === 'auto' ? (prefersDark ? 'peng' : 'catppuccin-latte') : settings.theme
    root.dataset.size = settings.size
    root.dataset.corners = settings.corners
  })

  $effect(() => {
    if (nav.view === 'system') nav.seenIssues = issues
  })

  $effect(() => {
    const cpu = stream.latest?.cpu
    const host = info?.hostname
    document.title = [host, cpu && `cpu ${pct(cpu.total)}`, 'peng-top'].filter(Boolean).join(' · ')
  })
</script>

<svelte:window onkeydown={onKey} />

<div class="shell">
  <header class="topbar">
    <h1 class="brand">peng-top</h1>
    {#if info}
      <span class="host">{info.hostname}</span>
      <span class="muted host-detail">{info.platform} {info.release}</span>
      <span class="muted">up {duration((now - info.bootTime) / 1000)}</span>
      <span class="badge mode-{info.mode}" title="whose numbers these are">{modeLabel[info.mode]}</span>
    {/if}
    <span class="spacer"></span>
    {#if issues}
      <a class="issues" class:unseen={issues > nav.seenIssues} href="#system">
        {issues} unavailable
      </a>
    {/if}
    <button class="status status-{stream.status}" onclick={() => stream.toggle()} title="pause or resume (p)">
      <span class="status-dot" aria-hidden="true"></span>{statusLabel[stream.status]}
    </button>
    <time class="clock muted">{new Date(now).toLocaleTimeString()}</time>
    <a class="settings-link" href="#settings" aria-current={nav.view === 'settings' ? 'page' : undefined}>
      settings
    </a>
  </header>

  <nav class="tabs" aria-label="views">
    {#each views as view, i (view)}
      <a href="#{view}" class="tab" aria-current={nav.view === view ? 'page' : undefined}>
        <kbd>{i + 1}</kbd>{view}
      </a>
    {/each}
  </nav>

  <main class="screen">
    {#if stream.latest || nav.view === 'settings' || nav.view === 'system'}
      <Screen />
    {:else}
      <p class="waiting muted">{stream.status === 'lost' ? "can't reach peng-top, retrying" : 'waiting for the first reading'}</p>
    {/if}
  </main>

  <footer class="keybar" aria-label="keys">
    {#each [['1-8', 'tabs'], ['/', 'filter'], ['p', 'pause'], ['t', 'theme'], ['s', 'settings'], ['?', 'help']] as [key, what] (key)}
      <span class="keybar-item"><kbd>{key}</kbd>{what}</span>
    {/each}
  </footer>
</div>

<style>
  .shell {
    display: grid;
    grid-template-rows: auto auto 1fr auto;
    min-height: 100dvh;

    @media (min-width: 40.01rem) {
      height: 100dvh;
    }
  }

  .topbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0 2ch;
    padding: 0.25lh 2ch;
    background: var(--surface-raised);
  }

  .brand {
    color: var(--accent);
  }

  .host {
    font-weight: 700;
  }

  .spacer {
    flex: 1;
  }

  .badge,
  .issues {
    padding-inline: 1ch;
    background: var(--surface-sunken);
    color: var(--text);
    text-decoration: none;
  }

  /* a warn block in front rather than warn text, so it reads in every theme */
  .mode-container,
  .issues.unseen {
    display: inline-flex;
    align-items: center;
    gap: 1ch;

    &::before {
      content: '';
      inline-size: 1ch;
      block-size: 1ch;
      background: var(--warn);
    }
  }

  .issues:not(.unseen) {
    color: var(--text-muted);
    background: none;
  }

  .status {
    display: inline-flex;
    align-items: center;
    gap: 1ch;
    padding: 0;
    border: 0;
    background: none;
    cursor: pointer;
  }

  .status-dot {
    inline-size: 1ch;
    block-size: 1ch;
    border-radius: 50%;
    background: var(--text-muted);

    .status-live & {
      background: var(--ok);
    }

    .status-lost & {
      background: var(--crit);
    }

    .status-paused & {
      background: var(--warn);
    }
  }

  .settings-link,
  .tab {
    color: var(--text-muted);
    text-decoration: none;

    &:hover {
      color: var(--text);
    }

    &[aria-current='page'] {
      background: var(--accent);
      color: var(--on-accent);

      & kbd {
        color: inherit;
      }
    }
  }

  .settings-link {
    padding-inline: 1ch;
    margin-inline: -1ch;
  }

  .tabs {
    display: flex;
    gap: 1ch;
    padding: 0.25lh 1ch;
    overflow-x: auto;
    scrollbar-width: none;
    border-bottom: var(--border-width) solid var(--border);
  }

  .tab {
    flex: none;
    padding-inline: 1ch;
    border-radius: var(--radius);

    & kbd {
      margin-right: 0.5ch;
    }
  }

  .screen {
    display: grid;
    align-content: start;
    padding: 1lh 2ch;
    min-width: 0;
    overflow: auto;

    &:has(> :global(.overview)) {
      align-content: stretch;
    }
  }

  .waiting {
    padding: 2lh 0;
    text-align: center;
  }

  .keybar {
    position: sticky;
    bottom: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 0 2ch;
    padding: 0.25lh 2ch;
    background: var(--surface-raised);
    color: var(--text-muted);
  }

  .keybar-item kbd {
    margin-right: 0.5ch;
  }

  @media (max-width: 40rem) {
    .host-detail,
    .clock {
      display: none;
    }

    .screen {
      padding-inline: 1ch;
    }

    .keybar {
      display: none;
    }

    .tabs {
      flex-wrap: wrap;
    }
  }
</style>
