<script lang="ts">
  import { settings } from '../lib/settings.svelte'
  import { sum } from '../lib/series'

  let {
    series,
    label,
    max,
    tone = 'load',
    mirror = false,
    stack = false,
    scale,
  }: {
    series: number[][]
    label: string
    max?: number
    tone?: 'load' | 'series'
    mirror?: boolean
    stack?: boolean
    scale?: (n: number) => string
  } = $props()

  let canvas = $state<HTMLCanvasElement>()
  let cell = $state<DOMRectReadOnly>()
  let size = $state<DOMRectReadOnly>()

  const addUp = (layers: number[][]) =>
    layers[0].map((_, i) => sum(layers.map((layer) => layer[i])))
  const lanes = $derived(stack ? series.map((_, n) => addUp(series.slice(0, n + 1))) : series)
  const ceiling = $derived(max ?? niceMax(Math.max(1, ...lanes.flat()), settings.binary ? 1024 : 1000))

  // rounds up to 1, 2 or 5 times a power of ten in the peak's display unit
  function niceMax(peak: number, base: number) {
    const unit = base ** Math.floor(Math.log(peak) / Math.log(base))
    const inUnit = peak / unit
    const step = 10 ** Math.floor(Math.log10(inUnit))
    return ([1, 2, 5, 10].map((m) => m * step).find((n) => n >= inUnit) ?? inUnit) * unit
  }

  function colours(el: HTMLElement) {
    const css = getComputedStyle(el)
    const read = (name: string) => css.getPropertyValue(name).trim()
    return {
      ok: read('--ok'),
      warn: read('--warn'),
      crit: read('--crit'),
      a: read('--series-a'),
      b: read('--series-b'),
      info: read('--info'),
      faint: read('--border'),
    }
  }

  $effect(() => {
    if (!canvas || !cell?.width || !size?.width) return
    const box = size
    void settings.theme
    void settings.graph
    const dpr = devicePixelRatio || 1
    const width = Math.round(box.width * dpr)
    const height = Math.round(box.height * dpr)
    if (canvas.width !== width || canvas.height !== height) Object.assign(canvas, { width, height })
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    ctx.resetTransform()
    ctx.clearRect(0, 0, width, height)
    ctx.scale(dpr, dpr)

    const ink = colours(canvas)
    const step = { x: cell.width / 2, y: cell.height / 4 }
    const columns = Math.floor(box.width / step.x)
    const laneHeight = mirror ? box.height / 2 : box.height
    const rows = Math.floor(laneHeight / step.y)
    const toRows = (v: number) => Math.round((Math.min(v, ceiling) / ceiling) * rows)
    const drawn = mirror ? lanes.slice(0, 2) : lanes
    // a stack gets topped up to the ceiling in faint ink, so you can see what's left
    const layers = stack ? [...drawn, drawn[0].map(() => ceiling)] : drawn
    const shade = (lane: number, frac: number) => {
      if (lane === drawn.length) return ink.faint
      if (tone === 'series') return lane === 0 ? ink.a : ink.b
      if (lane > 0) return ink.info
      if (frac < 0.5) return ink.ok
      return frac < 0.8 ? ink.warn : ink.crit
    }

    layers.forEach((values, lane) => {
      const down = mirror && lane === 1
      const base = mirror ? box.height / 2 : box.height
      const shown = values.slice(-columns)
      const floor = stack && lane ? layers[lane - 1].slice(-columns) : shown.map(() => 0)
      if (settings.graph === 'line') {
        drawLine(ctx, shown, floor, columns, step.x, base, laneHeight, down, shade(lane, 0.3))
        return
      }
      // stacked layers get smaller dots, so they're told apart by more than colour
      const pad = stack && lane ? 0.3 : 0.2
      const size = 1 - 2 * pad
      shown.forEach((v, i) => {
        const x = box.width - (shown.length - i) * step.x
        for (let r = toRows(floor[i]); r < toRows(v); r++) {
          ctx.fillStyle = shade(lane, r / rows)
          const y = down ? base + r * step.y : base - (r + 1) * step.y
          ctx.fillRect(x + step.x * pad, y + step.y * pad, step.x * size, step.y * size)
        }
      })
    })
    if (mirror) {
      ctx.fillStyle = ink.faint
      ctx.fillRect(0, box.height / 2, box.width, 1)
    }
  })

  function drawLine(
    ctx: CanvasRenderingContext2D,
    values: number[],
    floor: number[],
    columns: number,
    stepX: number,
    base: number,
    height: number,
    down: boolean,
    colour: string,
  ) {
    const width = columns * stepX
    const offset = ctx.canvas.width / (devicePixelRatio || 1) - width
    const y = (v: number) => base + (down ? 1 : -1) * (Math.min(v, ceiling) / ceiling) * height
    const x = (i: number) => offset + width - (values.length - 1 - i) * stepX
    ctx.beginPath()
    values.forEach((v, i) => (i ? ctx.lineTo(x(i), y(v)) : ctx.moveTo(x(i), y(v))))
    ctx.strokeStyle = colour
    ctx.lineWidth = 1.5
    ctx.stroke()
    for (let i = values.length - 1; i >= 0; i--) ctx.lineTo(x(i), y(floor[i]))
    ctx.globalAlpha = 0.18
    ctx.fillStyle = colour
    ctx.fill()
    ctx.globalAlpha = 1
  }
</script>

<div class="graph" bind:contentRect={size} role="img" aria-label={label}>
  <canvas bind:this={canvas} aria-hidden="true"></canvas>
  {#if scale}<span class="graph-scale" aria-hidden="true">{mirror ? '±' : ''}{scale(ceiling)}</span>{/if}
  <span class="graph-probe" bind:contentRect={cell} aria-hidden="true"></span>
</div>

<style>
  .graph {
    position: relative;
    flex: 1 0 auto;
    block-size: var(--graph-height, 5lh);
    min-width: 0;
    background: var(--surface-sunken);
    border-radius: var(--radius);
    overflow: hidden;
  }

  canvas {
    position: absolute;
    inset: 0;
    inline-size: 100%;
    block-size: 100%;
  }

  .graph-scale {
    position: absolute;
    inset: 0 auto auto 0;
    padding-inline: 0.5ch;
    background: var(--surface-sunken);
    color: var(--text-muted);
    pointer-events: none;
  }

  .graph-probe {
    position: absolute;
    inline-size: 1ch;
    block-size: 1lh;
    visibility: hidden;
  }
</style>
