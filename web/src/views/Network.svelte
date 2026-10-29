<script lang="ts">
  import Graph from '../components/Graph.svelte'
  import Panel from '../components/Panel.svelte'
  import Unavailable from '../components/Unavailable.svelte'
  import { bytes, rate } from '../lib/format'
  import { series, sum, visibleNics } from '../lib/series'
  import { settings } from '../lib/settings.svelte'
  import { stream } from '../lib/stream.svelte'
  import type { NIC, Snapshot } from '../lib/types'

  let { compact = false }: { compact?: boolean } = $props()

  const nics = $derived(visibleNics(stream.latest?.net, settings.showVirtual))
  const speed = (n: number) => rate(n, settings.binary, settings.bits)
  const size = (n: number) => bytes(n, settings.binary)
  const pick = (s: Snapshot, key: 'rx' | 'tx', name?: string) =>
    sum(visibleNics(s.net, settings.showVirtual).filter((n) => !name || n.name === name).map((n) => n[key]))
  const traffic = (name?: string) => [
    series(stream.history, (s) => pick(s, 'rx', name)),
    series(stream.history, (s) => pick(s, 'tx', name)),
  ]
  const totals = $derived({ rx: sum(nics.map((n) => n.rx)), tx: sum(nics.map((n) => n.tx)) })
  const describe = (what: string, rx: number, tx: number) =>
    `${what}, downloads above the line and uploads below, now ${speed(rx)} down and ${speed(tx)} up`
</script>

{#snippet legend(rx: number, tx: number)}
  <p class="legend">
    <span class="key-a">down {speed(rx)}</span>
    <span class="key-b">up {speed(tx)}</span>
  </p>
{/snippet}


<div class="network-view" class:compact>
  {#if compact}
    <Panel title="network" hotkey="5">
      {#if stream.latest?.net}
        <Graph series={traffic()} tone="series" mirror scale={speed} label={describe('all traffic', totals.rx, totals.tx)} />
        {@render legend(totals.rx, totals.tx)}
        {#each nics as n (n.name)}
          <p class="nic-line">
            <span class="nic-name">{n.name}</span>
            <span>↓ {speed(n.rx)}</span>
            <span>↑ {speed(n.tx)}</span>
          </p>
        {/each}
      {:else}
        <Unavailable area="network" what="network readings" />
      {/if}
    </Panel>
  {:else if !stream.latest?.net}
    <Panel title="network"><Unavailable area="network" what="network readings" /></Panel>
  {:else}
    <div class="network-tools">
      <button class="toggle" aria-pressed={settings.showVirtual} onclick={() => (settings.showVirtual = !settings.showVirtual)}>
        <span class="toggle-mark" aria-hidden="true">{settings.showVirtual ? '[x]' : '[ ]'}</span>
        show virtual interfaces <span class="muted">(lo, docker, veth)</span>
      </button>
    </div>
    {#each nics as n (n.name)}
      {@render nicPanel(n)}
    {:else}
      <Panel title="network">
        <p class="muted">only virtual interfaces here, tick the box above to see them</p>
      </Panel>
    {/each}
  {/if}
</div>

{#snippet nicPanel(n: NIC)}
  <Panel title={n.name} class={n.up ? 'nic' : 'nic nic-down'}>
    <Graph series={traffic(n.name)} tone="series" mirror scale={speed} label={describe(n.name, n.rx, n.tx)} />
    {@render legend(n.rx, n.tx)}
    <dl class="nic-facts">
      <dt class="muted">state</dt>
      <dd>{n.up ? 'up' : 'down'}{n.virtual ? ', virtual' : ''}</dd>
      <dt class="muted">since boot</dt>
      <dd>{size(n.rxTotal)} down, {size(n.txTotal)} up</dd>
      {#if n.addrs?.length}
        <dt class="muted">addresses</dt>
        <dd class="addrs">{#each n.addrs as a (a)}<span>{a}</span>{/each}</dd>
      {/if}
      {#if n.mac}
        <dt class="muted">mac</dt>
        <dd>{n.mac}</dd>
      {/if}
      {#if n.mtu}
        <dt class="muted">mtu</dt>
        <dd>{n.mtu}</dd>
      {/if}
      <dt class="muted">errors / drops</dt>
      <dd class:bad={n.errors + n.drops > 0}>{n.errors} / {n.drops}</dd>
    </dl>
  </Panel>
{/snippet}

<style>
  .network-view {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 44ch), 1fr));
    gap: 1.25lh 2ch;

    &.compact {
      display: contents;
    }
  }

  .legend,
  .nic-line {
    display: flex;
    flex-wrap: wrap;
    gap: 0 2ch;
  }

  .nic-name {
    min-inline-size: 12ch;
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .nic-facts {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0 2ch;

    & dd {
      min-width: 0;
      overflow-wrap: anywhere;
    }
  }

  .addrs {
    display: flex;
    flex-direction: column;
  }

  .bad {
    color: var(--crit);
    font-weight: 700;
  }

  .network-tools {
    grid-column: 1 / -1;
  }

  .toggle {
    padding: 0;
    border: 0;
    background: none;
    color: var(--text-muted);
    cursor: pointer;

    &[aria-pressed='true'] {
      color: var(--text);

      & .toggle-mark {
        color: var(--accent);
      }
    }
  }

  .network-view :global(.nic .graph) {
    flex-grow: 0;
  }

  .network-view :global(.nic-down) {
    opacity: 0.7;
  }
</style>
