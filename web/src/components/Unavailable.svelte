<script lang="ts">
  import { stream } from '../lib/stream.svelte'

  let { area, what }: { area: string; what: string } = $props()

  const why = $derived(stream.info?.issues[area])
</script>

<p class="unavailable">
  <span class="unavailable-what">no {what} here</span>
  {#if why}<span class="muted">{why}</span>{/if}
</p>

<style>
  .unavailable {
    display: flex;
    flex-wrap: wrap;
    gap: 0 1ch;
    padding-block: 0.5lh;

    &::before {
      content: '--';
      color: var(--text-muted);
    }
  }

  .unavailable-what {
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
</style>
