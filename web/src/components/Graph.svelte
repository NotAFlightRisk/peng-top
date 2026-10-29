<script lang="ts">
  import { settings } from '../lib/settings.svelte'

  let {
    series,
    label,
    max,
    tone = 'load',
    mirror = false,
  }: {
    series: number[][]
    label: string
    max?: number
    tone?: 'load' | 'series'
    mirror?: boolean
  } = $props()

  let canvas = $state<HTMLCanvasElement>()
  let probe = $state<HTMLSpanElement>()
  let size = $state<DOMRectReadOnly>()

  const ceiling = $derived(max ?? niceMax(Math.max(1, ...series.flat())))

  // rounds up to 1, 2 or 5 times a power of ten, so the scale doesn't wobble every tick
  function niceMax(peak: number) {
    const step = 10 ** Math.floor(Math.log10(peak))
    return [1, 2, 5, 10].map((m) => m * step).find((n) => n >= peak) ?? peak
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
      faint: read('--border'),
    }
  }

  $effect(() => {
    if (!canvas || !probe || !size?.width) return
    const box = size
    void settings.theme
    void settings.graph
    const dpr = devicePixelRatio || 1
    canvas.width = Math.round(box.width * dpr)
    canvas.height = Math.round(box.height * dpr)
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    ctx.scale(dpr, dpr)
    ctx.clearRect(0, 0, box.width, box.height)

    const ink = colours(canvas)
    const cell = probe.getBoundingClientRect()
    const step = { x: cell.width / 2, y: cell.height / 4 }
    const lanes = mirror ? series.slice(0, 2) : series
    const laneHeight = mirror ? box.height / 2 : box.height
    const shade = (lane: number, frac: number) => {
      if (tone === 'series') return lane === 0 ? ink.a : ink.b
      if (frac < 0.5) return ink.ok
      return frac < 0.8 ? ink.warn : ink.crit
    }

    lanes.forEach((values, lane) => {
      const down = mirror && lane === 1
      const base = mirror ? box.height / 2 : box.height
      const columns = Math.floor(box.width / step.x)
      const shown = values.slice(-columns)
      if (settings.graph === 'line') {
        drawLine(ctx, shown, columns, step.x, base, laneHeight, down, shade(lane, 0.3))
        return
      }
      const rows = Math.floor(laneHeight / step.y)
      shown.forEach((v, i) => {
        const x = box.width - (shown.length - i) * step.x
        const filled = Math.round((Math.min(v, ceiling) / ceiling) * rows)
        for (let r = 0; r < filled; r++) {
          ctx.fillStyle = shade(lane, r / rows)
          const y = down ? base + r * step.y : base - (r + 1) * step.y
          ctx.fillRect(x + step.x * 0.2, y + step.y * 0.2, step.x * 0.6, step.y * 0.6)
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
    ctx.lineTo(x(values.length - 1), base)
    ctx.lineTo(x(0), base)
    ctx.globalAlpha = 0.18
    ctx.fillStyle = colour
    ctx.fill()
    ctx.globalAlpha = 1
  }
</script>

<div class="graph" bind:contentRect={size} role="img" aria-label={label}>
  <canvas bind:this={canvas} aria-hidden="true"></canvas>
  <span class="graph-probe" bind:this={probe} aria-hidden="true"></span>
</div>

<style>
  .graph {
    position: relative;
    flex: 1 0 auto;
    block-size: var(--graph-height, 5lh);
    max-block-size: 22lh;
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

  .graph-probe {
    position: absolute;
    inline-size: 1ch;
    block-size: 1lh;
    visibility: hidden;
  }
</style>
