<script lang="ts">
  import Graph from '../components/Graph.svelte'
  import Meter from '../components/Meter.svelte'
  import Panel from '../components/Panel.svelte'
  import Unavailable from '../components/Unavailable.svelte'
  import { bytes, pct } from '../lib/format'
  import { series } from '../lib/series'
  import { settings } from '../lib/settings.svelte'
  import { stream } from '../lib/stream.svelte'
  import type { Mem } from '../lib/types'

  let { compact = false }: { compact?: boolean } = $props()

  const mem = $derived(stream.latest?.mem)
  const size = (n: number) => bytes(n, settings.binary)
  const total = $derived(mem ? mem.limit || mem.total : 1)
  const share = (n: number) => (n / total) * 100
  const cache = (m: Mem) => m.cached + m.buffers
  const swapShare = $derived(mem?.swapTotal ? (mem.swapUsed / mem.swapTotal) * 100 : 0)
  const overTime = (pick: (m: Mem) => number) =>
    series(stream.history, (s) => s.mem && (pick(s.mem) / (s.mem.limit || s.mem.total)) * 100)
  const usedHistory = $derived(overTime((m) => m.used))
  const cacheHistory = $derived(overTime(cache))
  const swapHistory = $derived(
    series(stream.history, (s) => s.mem?.swapTotal && (s.mem.swapUsed / s.mem.swapTotal) * 100),
  )
  const hungriest = $derived(
    [...(stream.latest?.procs?.list ?? [])].sort((a, b) => b.rss - a.rss).slice(0, 10),
  )
</script>

{#snippet summary()}
  {#if mem}<span>{size(mem.used)} of {size(total)}</span>{/if}
{/snippet}

<div class="memory-view" class:compact>
  <Panel title="memory" hotkey={compact ? '3' : undefined} aside={summary} class="memory-main">
    {#if mem}
      <Graph
        series={[usedHistory, cacheHistory]}
        max={100}
        stack
        label="memory over time, used with cache stacked on top, now {pct(share(mem.used))} used and {pct(share(cache(mem)))} cache"
      />
      <Meter label="used" value={share(mem.used)} detail={size(mem.used)} />
      <Meter label="cache" value={share(cache(mem))} detail={size(cache(mem))} tone="info" />
      <Meter
        label="available"
        value={share(mem.available)}
        detail={size(mem.available)}
        tone="series-a"
      />
      {#if mem.swapTotal}
        <Meter label="swap" value={swapShare} detail={size(mem.swapUsed)} />
      {/if}
      {#if mem.limit}
        <p class="muted">container limit {size(mem.limit)}, the host has {size(mem.total)}</p>
      {/if}
    {:else}
      <Unavailable area="memory" what="memory readings" />
    {/if}
  </Panel>

  {#if !compact && mem}
    <Panel title="swap" aside={swapNote}>
      {#if mem.swapTotal}
        <Graph series={[swapHistory]} max={100} label="swap use over time, now {pct(swapShare)}" />
      {:else}
        <Unavailable area="swap" what="swap" />
      {/if}
    </Panel>

    <Panel title="the numbers">
      <dl class="figures">
        {#each [['total', mem.total], ['used', mem.used], ['available', mem.available], ['cached', mem.cached], ['buffers', mem.buffers], ['free', mem.free], ['swap total', mem.swapTotal], ['swap used', mem.swapUsed]] as const as [label, value] (label)}
          <dt class="muted">{label}</dt>
          <dd>{size(value)}</dd>
        {/each}
      </dl>
    </Panel>

    <Panel title="hungriest processes" class="memory-procs">
      <ol class="hungry">
        {#each hungriest as proc (proc.pid)}
          <li>
            <Meter label={proc.name} value={share(proc.rss)} detail={size(proc.rss)} tone="info" />
          </li>
        {/each}
      </ol>
    </Panel>
  {/if}
</div>

{#snippet swapNote()}
  {#if mem?.swapTotal}<span>{size(mem.swapUsed)} of {size(mem.swapTotal)}</span>{/if}
{/snippet}

<style>
  .memory-view {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 44ch), 1fr));
    gap: 1.25lh 2ch;

    &.compact {
      display: contents;
    }

    & :global(.memory-main .meter) {
      --label-width: 9ch;
    }

    &:not(.compact) :global(.memory-main) {
      grid-column: 1 / -1;

      & :global(.graph) {
        --graph-height: 8lh;
      }
    }
  }

  .figures {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0 2ch;

    & dd {
      text-align: right;
    }
  }

  .hungry :global(.meter) {
    --label-width: 18ch;
    --value-width: 9ch;
  }
</style>
