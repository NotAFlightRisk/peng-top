<script lang="ts">
  import Graph from '../components/Graph.svelte'
  import Meter from '../components/Meter.svelte'
  import Panel from '../components/Panel.svelte'
  import Unavailable from '../components/Unavailable.svelte'
  import { bytes, pct, rate } from '../lib/format'
  import { series } from '../lib/series'
  import { settings } from '../lib/settings.svelte'
  import { stream } from '../lib/stream.svelte'

  let { compact = false }: { compact?: boolean } = $props()

  const mounts = $derived(stream.latest?.disks?.mounts ?? [])
  const io = $derived(stream.latest?.disks?.io ?? [])
  const size = (n: number) => bytes(n, settings.binary)
  const speed = (n: number) => rate(n, settings.binary)
  const used = (m: { used: number; free: number }) => (m.used / (m.used + m.free)) * 100
  const ioHistory = (name: string, key: 'read' | 'write') =>
    series(stream.history, (s) => s.disks?.io?.find((d) => d.name === name)?.[key])
</script>

<div class="disks-view" class:compact>
  <Panel title="disks" hotkey={compact ? '4' : undefined} class="disks-main">
    {#if mounts.length}
      {#if compact}
        {#each mounts as m (m.path)}
          <Meter label={m.path} value={used(m)} detail="{size(m.used)} / {size(m.total)}" />
        {/each}
        {#each io as d (d.name)}
          <p class="io-line">
            <span class="muted">{d.name}</span>
            <span>r {speed(d.read)}</span>
            <span>w {speed(d.write)}</span>
          </p>
        {/each}
      {:else}
        <div class="mount-scroll">
          <table class="mount-table">
            <thead>
              <tr>
                <th scope="col">mount</th>
                <th scope="col" class="opt">device</th>
                <th scope="col" class="opt">fs</th>
                <th scope="col" class="num">size</th>
                <th scope="col" class="num">free</th>
                <th scope="col" class="bar">used</th>
                <th scope="col" class="num opt">inodes</th>
              </tr>
            </thead>
            <tbody>
              {#each mounts as m (m.path)}
                <tr>
                  <th scope="row" class="mount-path" title={m.path}>{m.path}</th>
                  <td class="opt muted">{m.device}</td>
                  <td class="opt muted">{m.fs}</td>
                  <td class="num">{size(m.total)}</td>
                  <td class="num">{size(m.free)}</td>
                  <td class="bar"><Meter label={m.path} value={used(m)} /></td>
                  <td class="num opt">{m.inodes ? pct(m.inodes) : '-'}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    {:else}
      <Unavailable area="disk usage" what="filesystems" />
    {/if}
  </Panel>

  {#if !compact}
    {#each io as d (d.name)}
      {#snippet ioNote()}{#if d.busy !== undefined}<span>{pct(d.busy)} busy</span>{/if}{/snippet}
      <Panel title={d.name} aside={ioNote}>
        <Graph
          series={[ioHistory(d.name, 'read'), ioHistory(d.name, 'write')]}
          tone="series"
          mirror
          scale={speed}
          label="{d.name} reads above the line and writes below, now {speed(d.read)} read and {speed(d.write)} written"
        />
        <p class="io-legend">
          <span class="key-a">read {speed(d.read)}</span>
          <span class="key-b">write {speed(d.write)}</span>
          <span class="muted">{Math.round(d.readOps + d.writeOps)} ops/s</span>
        </p>
        {#if d.busy !== undefined}
          <Meter label="busy" value={d.busy} detail={pct(d.busy)} />
        {/if}
        <p class="muted">{size(d.readTotal)} read, {size(d.writeTotal)} written since boot</p>
      </Panel>
    {:else}
      <Panel title="disk activity">
        <Unavailable area="disk io" what="disk activity" />
      </Panel>
    {/each}
  {/if}
</div>

<style>
  .disks-view {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 44ch), 1fr));
    gap: 1.25lh 2ch;

    &.compact {
      display: contents;
    }

    &:not(.compact) :global(.disks-main) {
      grid-column: 1 / -1;
    }

    & :global(.meter) {
      --label-width: 12ch;
      --value-width: 18ch;
    }
  }

  .io-line,
  .io-legend {
    display: flex;
    flex-wrap: wrap;
    gap: 0 2ch;
  }

  .io-line span:first-child {
    min-inline-size: 12ch;
  }


  .mount-scroll {
    overflow-x: auto;
    container-type: inline-size;
  }

  .mount-table {
    inline-size: 100%;
    border-collapse: collapse;

    & th,
    & td {
      padding: 0 1ch;
      text-align: left;
      font-weight: 400;
      white-space: nowrap;
    }

    & thead th {
      background: var(--surface-raised);
      color: var(--text-muted);
    }

    & .num {
      text-align: right;
    }

    & .bar {
      inline-size: 40%;
      min-inline-size: 16ch;

      & :global(.meter) {
        grid-template-columns: 0 1fr auto;
        --value-width: 4ch;
      }

      & :global(.meter-label) {
        visibility: hidden;
      }
    }

    & .mount-path {
      max-inline-size: 30ch;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }

  @container (max-width: 70ch) {
    .mount-table .opt {
      display: none;
    }
  }
</style>
