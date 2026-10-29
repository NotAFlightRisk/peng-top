<script lang="ts">
  import Panel from '../components/Panel.svelte'
  import { bindings } from '../lib/keys'
  import { overviewPanels, reset, settings, themes } from '../lib/settings.svelte'

  type Choice = 'size' | 'corners' | 'graph' | 'binary' | 'bits' | 'tempUnit' | 'cpuPerCore' | 'rows'
  type Flag = 'showKernel' | 'showVirtual'

  const look: [Choice, string, [unknown, string][]][] = [
    ['size', 'text size', [['s', 'small'], ['m', 'medium'], ['l', 'large']]],
    ['corners', 'corners', [['square', 'square'], ['round', 'round']]],
    ['graph', 'graphs', [['dots', 'dots'], ['line', 'line']]],
  ]
  const units: [Choice, string, [unknown, string][]][] = [
    ['binary', 'sizes', [[true, 'KiB (1024)'], [false, 'kB (1000)']]],
    ['bits', 'network', [[false, 'bytes/s'], [true, 'bits/s']]],
    ['tempUnit', 'temperature', [['c', '°C'], ['f', '°F']]],
  ]
  const procs: [Choice, string, [unknown, string][]][] = [
    ['cpuPerCore', 'cpu %', [[true, 'per core'], [false, 'whole machine']]],
    ['rows', 'rows to show', [[50, '50'], [100, '100'], [250, '250'], [500, '500']]],
  ]

  const pick = (key: Choice, value: unknown) => ((settings as Record<Choice, unknown>)[key] = value)
  const flip = (key: Flag) => (settings[key] = !settings[key])
</script>

<div class="settings">
  {#snippet choices(rows: [Choice, string, [unknown, string][]][])}
    {#each rows as [key, legend, options] (key)}
      <fieldset class="setting-row">
        <legend>{legend}</legend>
        <div class="setting-options">
          {#each options as [value, label] (label)}
            <label class="setting-option">
              <input
                class="visually-hidden"
                type="radio"
                name={key}
                checked={settings[key] === value}
                onchange={() => pick(key, value)}
              />
              <span class="setting-mark" aria-hidden="true"></span>{label}
            </label>
          {/each}
        </div>
      </fieldset>
    {/each}
  {/snippet}

  {#snippet check(key: Flag, label: string)}
    <label class="setting-option setting-check">
      <input class="visually-hidden" type="checkbox" checked={settings[key]} onchange={() => flip(key)} />
      <span class="setting-mark setting-box" aria-hidden="true"></span><span>{label}</span>
    </label>
  {/snippet}

  <Panel title="look">
    <fieldset class="setting-row">
      <legend>theme</legend>
      <div class="themes">
        {#each themes as [name, label] (name)}
          <label class="setting-option theme-option">
            <input
              class="visually-hidden"
              type="radio"
              name="theme"
              checked={settings.theme === name}
              onchange={() => (settings.theme = name)}
            />
            <span class="setting-mark" aria-hidden="true"></span>
            <span class="swatch" data-theme={name === 'auto' ? 'peng' : name} aria-hidden="true">
              {#each ['accent', 'ok', 'warn', 'crit', 'info'] as role (role)}<i style:background="var(--{role})"></i>{/each}
            </span>
            {label}
          </label>
        {/each}
      </div>
    </fieldset>
    {@render choices(look)}
  </Panel>

  <Panel title="units">
    {@render choices(units)}
    {@render check('showVirtual', 'show virtual network interfaces (lo, docker, veth)')}
  </Panel>

  <Panel title="processes">
    {@render choices(procs)}
    {@render check('showKernel', 'show kernel threads')}
  </Panel>

  <Panel title="overview">
    {#each overviewPanels as panel (panel)}
      <label class="setting-option setting-check">
        <input class="visually-hidden" type="checkbox" bind:checked={settings.overview[panel]} />
        <span class="setting-mark setting-box" aria-hidden="true"></span><span>{panel}</span>
      </label>
    {/each}
  </Panel>

  <Panel title="keys">
    <dl class="keys">
      {#each bindings as b (b.label)}
        <dt>
          {#each b.keys as k (k)}<kbd>{k}</kbd>{/each}
        </dt>
        <dd>{b.label}{#if b.only}<span class="key-scope muted">({b.only})</span>{/if}</dd>
      {/each}
    </dl>
  </Panel>

  <div class="settings-foot">
    <button class="setting-reset" onclick={reset}>reset to defaults</button>
    <span class="muted">settings live in this browser only</span>
  </div>
</div>

<style>
  .settings {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 44ch), 1fr));
    align-items: start;
    gap: 1lh 2ch;
  }

  .setting-row {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0 2ch;
    margin: 0;
    padding: 0;
    border: 0;
    min-width: 0;

    & legend {
      float: left;
      inline-size: 13ch;
      padding: 0;
      color: var(--text-muted);
    }
  }

  .setting-options {
    display: flex;
    flex-wrap: wrap;
    gap: 0 2ch;
  }

  .setting-option {
    cursor: pointer;

    & input:focus-visible + .setting-mark {
      outline: var(--border-width) solid var(--accent);
    }
  }

  .setting-check {
    display: flex;
    align-items: baseline;
  }

  .key-scope {
    margin-inline-start: 1ch;
  }

  .setting-mark {
    flex: none;
    margin-right: 1ch;
    color: var(--text-muted);

    &::before {
      content: '( )';
    }

    :checked + &::before {
      content: '(•)';
      color: var(--accent);
    }

    &.setting-box::before {
      content: '[ ]';
    }

    :checked + &.setting-box::before {
      content: '[x]';
    }
  }

  .themes {
    display: grid;
  }

  .theme-option {
    display: flex;
    align-items: center;
  }

  /* each swatch wears its own theme, so these are that theme's colours */
  .swatch {
    display: inline-flex;
    gap: 0.25ch;
    margin-right: 1ch;
    padding: 0.2lh 0.5ch;
    background: var(--surface);
    border: var(--border-width) solid var(--border);

    & i {
      inline-size: 1ch;
      block-size: 1ch;
    }
  }

  .keys {
    display: grid;
    grid-template-columns: max-content 1fr;
    gap: 0 2ch;

    & dt {
      display: flex;
      gap: 1ch;
    }
  }

  .settings-foot {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0 2ch;
    grid-column: 1 / -1;
  }

  .setting-reset {
    padding: 0 1ch;
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius);
    background: var(--surface-sunken);
    cursor: pointer;

    &:hover {
      border-color: var(--crit);
    }
  }
</style>
