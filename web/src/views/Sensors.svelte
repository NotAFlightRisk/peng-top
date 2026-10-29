<script lang="ts">
  import Graph from '../components/Graph.svelte'
  import Meter from '../components/Meter.svelte'
  import Panel from '../components/Panel.svelte'
  import Unavailable from '../components/Unavailable.svelte'
  import { temp } from '../lib/format'
  import { series } from '../lib/series'
  import { settings } from '../lib/settings.svelte'
  import { stream } from '../lib/stream.svelte'
  import type { Sensor } from '../lib/types'

  let { compact = false }: { compact?: boolean } = $props()

  const sensors = $derived(stream.latest?.sensors ?? [])
  const shown = $derived(compact ? [...sensors].sort((a, b) => b.temp - a.temp).slice(0, 6) : sensors)
  const ceiling = (s: Sensor) => s.crit || s.high || 100
  const hottest = $derived(series(stream.history, (s) => Math.max(0, ...(s.sensors ?? []).map((x) => x.temp))))
  const peak = $derived(Math.max(0, ...sensors.map((s) => s.temp)))
</script>

{#snippet summary()}
  {#if sensors.length}<span>hottest {temp(peak, settings.tempUnit)}</span>{/if}
{/snippet}

<div class="sensors-view" class:compact>
  <Panel title="sensors" hotkey={compact ? '6' : undefined} aside={summary} class="sensors-main">
    {#if sensors.length}
      {#if !compact}
        <Graph series={[hottest]} max={110} label="hottest sensor over time, now {temp(peak, settings.tempUnit)}" />
      {/if}
      {#each shown as s (s.name)}
        <Meter
          label={s.name}
          value={(s.temp / ceiling(s)) * 100}
          detail={temp(s.temp, settings.tempUnit)}
        />
      {/each}
      {#if !compact}
        <p class="muted">bars run up to each sensor's critical point, or 100°C when it doesn't say</p>
      {/if}
    {:else}
      <Unavailable area="sensors" what="temperature sensors" />
    {/if}
  </Panel>
</div>

<style>
  .sensors-view {
    display: grid;
    gap: 1.25lh 2ch;

    &.compact {
      display: contents;
    }

    & :global(.meter) {
      --label-width: 24ch;
      --value-width: 6ch;
    }
  }
</style>
