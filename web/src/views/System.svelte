<script lang="ts">
  import Panel from '../components/Panel.svelte'
  import { duration, ghz } from '../lib/format'
  import { stream } from '../lib/stream.svelte'

  const info = $derived(stream.info)
  const now = $derived(stream.latest?.t ?? Date.now())
  const issues = $derived(Object.entries(info?.issues ?? {}))

  const facts = $derived(
    info
      ? [
          ['hostname', info.hostname],
          ['os', info.os],
          ['platform', [info.platform, info.release].filter(Boolean).join(' ')],
          ['kernel', info.kernel],
          ['arch', info.arch],
          ['virtualisation', info.virt],
          ['cpu', info.cpuModel],
          ['cores', info.cores && `${info.cores} cores, ${info.threads} threads`],
          ['max speed', info.maxMhz && ghz(info.maxMhz)],
          [
            'booted',
            info.bootTime &&
              `${new Date(info.bootTime).toLocaleString()} (up ${duration((now - info.bootTime) / 1000)})`,
          ],
          ['peng-top', info.version],
          ['readings', `every ${info.interval / 1000}s`],
        ].filter((fact): fact is [string, string] => Boolean(fact[1]))
      : [],
  )

  const modes = {
    host: "Running directly on this machine, so these are its own numbers.",
    'host-via-container': "In a container, reading the host's /proc and /sys.",
    container: "In a container without the host mounted, so you're seeing the container's own numbers.",
  }
</script>

<div class="system">
  <Panel title="system" hotkey="8">
    {#if info}
      <dl class="facts">
        {#each facts as [label, value] (label)}
          <dt>{label}</dt>
          <dd>{value}</dd>
        {/each}
      </dl>
    {:else}
      <p class="muted">fetching the details</p>
    {/if}
  </Panel>

  {#if info}
    <Panel title="what's watched">
      <p>{modes[info.mode]}</p>
      {#if info.mode === 'container'}
        <p class="muted">
          To watch the host instead, run it with <code>--pid host --network host -v /:/host:ro</code>
        </p>
      {/if}
    </Panel>

    <Panel title="unavailable here">
      {#if issues.length}
        <dl class="facts">
          {#each issues as [area, why] (area)}
            <dt>{area}</dt>
            <dd>{why}</dd>
          {/each}
        </dl>
      {:else}
        <p>Everything's readable.</p>
      {/if}
    </Panel>
  {/if}
</div>

<style>
  .system {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 48ch), 1fr));
    align-items: start;
    gap: 1lh 2ch;
  }

  .facts {
    display: grid;
    grid-template-columns: max-content minmax(0, 1fr);
    gap: 0 2ch;

    & dt {
      color: var(--text-muted);
    }

    & dd {
      overflow-wrap: anywhere;
    }
  }

  code {
    font: inherit;
    color: var(--text);
    overflow-wrap: anywhere;
  }
</style>
