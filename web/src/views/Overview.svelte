<script lang="ts">
  import Panel from '../components/Panel.svelte'
  import { bytes, pct } from '../lib/format'
  import { settings } from '../lib/settings.svelte'
  import { stream } from '../lib/stream.svelte'
  import Cpu from './Cpu.svelte'
  import Disks from './Disks.svelte'
  import Memory from './Memory.svelte'
  import Network from './Network.svelte'
  import Sensors from './Sensors.svelte'

  const show = $derived(settings.overview)
  const procs = $derived(stream.latest?.procs)
  const busiest = $derived(
    [...(procs?.list ?? [])]
      .filter((p) => settings.showKernel || !p.kernel)
      .sort((a, b) => b.cpu - a.cpu)
      .slice(0, 15),
  )
  const cpuShare = (n: number) => (settings.cpuPerCore ? n : n / (stream.info?.threads || 1))
</script>

{#snippet procCount()}
  {#if procs}<span>{procs.total} · {procs.running} running</span>{/if}
{/snippet}

<div class="overview">
  {#if show.cpu}<Cpu compact />{/if}
  {#if show.memory}<Memory compact />{/if}
  {#if show.disks}<Disks compact />{/if}
  {#if show.network}<Network compact />{/if}
  {#if show.sensors && stream.latest?.sensors}<Sensors compact />{/if}
  {#if show.processes && procs}
    <Panel title="busiest" hotkey="7" aside={procCount} class="overview-procs">
      <table class="top-procs">
        <thead>
          <tr>
            <th scope="col" class="num">pid</th>
            <th scope="col">name</th>
            <th scope="col" class="num">cpu</th>
            <th scope="col" class="num">mem</th>
          </tr>
        </thead>
        <tbody>
          {#each busiest as p (p.pid)}
            <tr>
              <td class="num muted">{p.pid}</td>
              <td class="name" title={p.cmd || p.name}>{p.name}</td>
              <td class="num">{pct(cpuShare(p.cpu))}</td>
              <td class="num">{bytes(p.rss, settings.binary)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </Panel>
  {/if}
</div>

<style>
  .overview {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(min(100%, 46ch), 1fr));
    grid-auto-flow: row dense;
    gap: 1.25lh 2ch;
    container-type: inline-size;

    @media (min-width: 40.01rem) {
      grid-auto-rows: minmax(min-content, 1fr);
      min-block-size: 100%;
    }
  }

  @container (max-width: 60ch) {
    .overview :global(.graph) {
      --graph-height: 3lh;
      flex-grow: 0;
    }

    .top-procs tr:nth-child(n + 9) {
      display: none;
    }
  }

  @media (min-width: 112rem) {
    .overview {
      grid-template-columns: repeat(3, 1fr);
    }
  }

  @container (min-width: 96ch) {
    .overview :global(.cpu-main) {
      grid-column: span 2;
    }
  }

  .top-procs {
    inline-size: 100%;
    border-collapse: collapse;
    table-layout: fixed;

    & th,
    & td {
      padding: 0 1ch 0 0;
      text-align: left;
      font-weight: 400;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    & thead th {
      color: var(--text-muted);
    }

    & .num {
      inline-size: 9ch;
      text-align: right;
    }
  }
</style>
