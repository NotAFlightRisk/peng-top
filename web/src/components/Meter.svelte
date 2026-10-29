<script lang="ts">
  import type { Snippet } from 'svelte'

  let {
    label,
    value,
    detail,
    tone = 'load',
    extra,
  }: {
    label: string
    value: number
    detail?: string
    tone?: 'load' | 'info' | 'series-a' | 'series-b'
    extra?: Snippet
  } = $props()

  const fill = $derived(Math.min(Math.max(value, 0), 100))
</script>

<div class="meter tone-{tone}">
  <span class="meter-label" title={label}>{label}</span>
  <span
    class="meter-track"
    role="meter"
    aria-label={label}
    aria-valuemin={0}
    aria-valuemax={100}
    aria-valuenow={Math.round(fill)}
  >
    <span class="meter-fill" style:clip-path="inset(0 {100 - fill}% 0 0)"></span>
  </span>
  <span class="meter-value">{detail ?? `${Math.round(value)}%`}</span>
  {@render extra?.()}
</div>

<style>
  .meter {
    --label-width: 8ch;
    --value-width: 6ch;
    display: grid;
    grid-template-columns: minmax(0, var(--label-width)) minmax(4ch, 1fr) minmax(var(--value-width), auto);
    align-items: center;
    gap: 1ch;
    min-width: 0;
  }

  .meter-label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--text-muted);
  }

  .meter-value {
    text-align: right;
    white-space: nowrap;
  }

  /* chopped into character cells, so it reads like a row of ▮ blocks */
  .meter-track {
    --cells: repeating-linear-gradient(90deg, #000 0 calc(1ch - 2px), transparent 0 1ch);
    position: relative;
    block-size: 0.9lh;
    container-type: inline-size;
    background: var(--surface-sunken);
    mask: var(--cells);
  }

  .meter-fill {
    position: absolute;
    inset: 0;
    transition: clip-path 0.6s var(--ease);
  }

  .tone-load .meter-fill {
    background: linear-gradient(90deg, var(--ok) 0 50cqi, var(--warn) 75cqi, var(--crit) 95cqi)
      0 0 / 100cqi 100%;
  }

  .tone-info .meter-fill {
    background: var(--info);
  }

  .tone-series-a .meter-fill {
    background: var(--series-a);
  }

  .tone-series-b .meter-fill {
    background: var(--series-b);
  }
</style>
