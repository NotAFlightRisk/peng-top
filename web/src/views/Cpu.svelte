<script lang="ts">
  import Graph from '../components/Graph.svelte'
  import Meter from '../components/Meter.svelte'
  import Panel from '../components/Panel.svelte'
  import Unavailable from '../components/Unavailable.svelte'
  import { ghz, pct } from '../lib/format'
  import { series } from '../lib/series'
  import { stream } from '../lib/stream.svelte'

  let { compact = false }: { compact?: boolean } = $props()

  const cpu = $derived(stream.latest?.cpu)
  const history = $derived(stream.history)
  const total = $derived(series(history, (s) => s.cpu?.total))
  const threads = $derived(stream.info?.threads || cpu?.cores.length || 1)
  const breakdown = $derived(
    cpu
      ? ([
          ['user', cpu.user],
          ['system', cpu.system],
          ['iowait', cpu.iowait],
          ['steal', cpu.steal],
        ] as const)
      : [],
  )
</script>

{#snippet summary()}
  {#if cpu}
    <span>{pct(cpu.total)}</span>
    {#if cpu.mhz}<span> · {ghz(cpu.mhz)}</span>{/if}
    {#if cpu.load}<span> · load {cpu.load.join(' ')}</span>{/if}
  {/if}
{/snippet}

<div class="cpu-view" class:compact>
  <Panel title="cpu" hotkey={compact ? '2' : undefined} aside={summary} class="cpu-main">
    {#if cpu}
      <Graph series={[total]} max={100} label="total cpu use over time, now {pct(cpu.total)}" />
      <div class="cores">
        {#each cpu.cores as value, i (i)}
          <Meter label="c{i}" {value} />
        {/each}
      </div>
    {:else}
      <Unavailable area="cpu" what="cpu readings" />
    {/if}
  </Panel>

  {#if !compact && cpu}
    <Panel title="where it goes">
      {#each breakdown as [label, value] (label)}
        <Meter {label} {value} detail={pct(value)} />
      {/each}
      {#if cpu.quota}
        <p class="muted">container's allowed {cpu.quota.toFixed(1)} of {threads} cores</p>
      {/if}
    </Panel>

    <Panel title="load average" aside={loadNote}>
      {#if cpu.load}
        {#each ['1 min', '5 min', '15 min'] as label, i (label)}
          <Meter {label} value={(cpu.load[i] / threads) * 100} detail={cpu.load[i].toFixed(2)} />
        {/each}
      {:else}
        <Unavailable area="load average" what="load average" />
      {/if}
    </Panel>

    <Panel title="per core" class="cpu-cores">
      <div class="core-graphs">
        {#each cpu.cores as value, i (i)}
          <figure class="core">
            <figcaption><span class="muted">c{i}</span> {pct(value)}</figcaption>
            <Graph
              series={[series(history, (s) => s.cpu?.cores[i])]}
              max={100}
              label="core {i} over time, now {pct(value)}"
            />
          </figure>
        {/each}
      </div>
    </Panel>
  {/if}
</div>

{#snippet loadNote()}
  <span>{threads} threads</span>
{/snippet}

<style>
  .cpu-view {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 44ch), 1fr));
    gap: 1.25lh 2ch;

    &.compact {
      display: contents;
    }

    &:not(.compact) {
      & :global(.cpu-main),
      & :global(.cpu-cores) {
        grid-column: 1 / -1;
      }

      & :global(.cpu-main .graph) {
        --graph-height: 8lh;
      }
    }
  }

  .cores {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(18ch, 1fr));
    gap: 0 2ch;

    & :global(.meter) {
      --label-width: 4ch;
      --value-width: 4ch;
    }
  }

  .core-graphs {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(24ch, 1fr));
    gap: 0.5lh 2ch;
  }

  .core {
    margin: 0;

    & :global(.graph) {
      --graph-height: 3lh;
    }
  }
</style>
