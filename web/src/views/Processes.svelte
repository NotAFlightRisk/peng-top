<script lang="ts">
  import Panel from '../components/Panel.svelte'
  import { bytes, duration, pct } from '../lib/format'
  import { rows, sortBy, table, type SortKey } from '../lib/procs.svelte'
  import { settings } from '../lib/settings.svelte'
  import { stream } from '../lib/stream.svelte'

  const columns: { label: string; sort?: SortKey; class: string }[] = [
    { label: 'pid', sort: 'pid', class: 'col-pid num' },
    { label: 'user', sort: 'user', class: 'col-user' },
    { label: 'cpu%', sort: 'cpu', class: 'col-cpu num' },
    { label: 'mem%', sort: 'mem', class: 'col-mem num' },
    { label: 'rss', class: 'col-rss num' },
    { label: 'thr', sort: 'threads', class: 'col-threads num' },
    { label: 'state', class: 'col-state' },
    { label: 'name', sort: 'name', class: 'col-name' },
  ]

  const procs = $derived(stream.latest?.procs)
  const all = $derived(procs ? rows(procs.list, settings.showKernel) : [])
  const capped = $derived(table.all || table.filter.trim() ? all : all.slice(0, settings.rows))
  const scale = $derived(settings.cpuPerCore ? 1 : stream.info?.threads || 1)
  const picked = $derived(procs?.list.find((p) => p.pid === table.selected))
  const now = $derived(stream.latest?.t ?? Date.now())

  const toggles = [
    { key: 'T', label: 'tree', get: () => table.tree, flip: () => (table.tree = !table.tree) },
    {
      key: 'K',
      label: 'kernel threads',
      get: () => settings.showKernel,
      flip: () => (settings.showKernel = !settings.showKernel),
    },
    { key: 'a', label: 'show all', get: () => table.all, flip: () => (table.all = !table.all) },
  ]

  const ariaSort = (key?: SortKey) =>
    key && table.sort === key ? (table.desc ? 'descending' : 'ascending') : undefined
</script>

<Panel title="processes" hotkey="7">
  {#snippet aside()}
    {#if procs}{procs.total} procs · {procs.running} running · {procs.threads} threads{/if}
  {/snippet}

  <div class="proc-wrap">
    <div class="proc-tools">
      <label class="proc-filter">
        <span><kbd>/</kbd> filter</span>
        <input
          id="proc-filter"
          type="search"
          bind:value={table.filter}
          placeholder="pid, name, user or command"
          autocomplete="off"
          spellcheck="false"
        />
      </label>
      {#each toggles as t (t.key)}
        <button class="proc-toggle" aria-pressed={t.get()} onclick={t.flip}>
          <span class="proc-mark" aria-hidden="true">{t.get() ? '[x]' : '[ ]'}</span>
          {t.label} <kbd>{t.key}</kbd>
        </button>
      {/each}
    </div>

    {#if !procs}
      <p class="muted">can't read processes here</p>
    {:else}
      <div class="proc-scroll">
        <table class="proc-table">
          <thead>
            <tr>
              {#each columns as col (col.label)}
                <th class={col.class} aria-sort={ariaSort(col.sort)} scope="col">
                  {#if col.sort}
                    <button class="proc-sort" onclick={() => sortBy(col.sort!)}>
                      {col.label}{#if table.sort === col.sort}<span aria-hidden="true"
                          >{table.desc ? '▼' : '▲'}</span
                        >{/if}
                    </button>
                  {:else}
                    {col.label}
                  {/if}
                </th>
              {/each}
            </tr>
          </thead>
          <tbody>
            {#each capped as p (p.pid)}
              <tr class="proc-row" class:selected={p.pid === table.selected} data-pid={p.pid}>
                <td class="col-pid num">{p.pid}</td>
                <td class="col-user">{p.user}</td>
                <td class="col-cpu num">{pct(p.cpu / scale)}</td>
                <td class="col-mem num">{pct(p.mem)}</td>
                <td class="col-rss num">{bytes(p.rss, settings.binary)}</td>
                <td class="col-threads num">{p.threads}</td>
                <td class="col-state">{p.state}</td>
                <td class="col-name" title={p.cmd || p.name}>
                  <button
                    class="proc-pick"
                    style:padding-inline-start="{p.depth * 2}ch"
                    onclick={() => (table.selected = p.pid)}
                    >{#if p.depth}<span class="muted">└─ </span>{/if}{p.name}{#if p.cmd}<span
                        class="proc-cmd muted">{p.cmd}</span
                      >{/if}</button
                  >
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      {#if capped.length < all.length || table.all}
        <p class="proc-count muted">
          {#if table.all}
            showing all {all.length} - press <kbd>a</kbd> to trim
          {:else}
            showing {capped.length} of {all.length} - press <kbd>a</kbd> for all
          {/if}
        </p>
      {:else if !all.length}
        <p class="muted">nothing matches that</p>
      {/if}
    {/if}

    {#if picked}
      <section class="proc-detail" aria-label="selected process">
        <dl>
          <dt>pid</dt>
          <dd>{picked.pid}</dd>
          <dt>parent</dt>
          <dd>{picked.ppid}</dd>
          <dt>user</dt>
          <dd>{picked.user || 'unknown'}</dd>
          <dt>state</dt>
          <dd>{picked.state || 'unknown'}</dd>
          <dt>threads</dt>
          <dd>{picked.threads}</dd>
          <dt>started</dt>
          <dd>
            {duration((now - picked.started) / 1000)} ago
            <span class="muted">({new Date(picked.started).toLocaleString()})</span>
          </dd>
          <dt>memory</dt>
          <dd>{bytes(picked.rss, settings.binary)}</dd>
        </dl>
        <code class="proc-full">{picked.cmd || picked.name}</code>
      </section>
    {/if}
  </div>
</Panel>

<style>
  .proc-wrap {
    container-type: inline-size;
    display: flex;
    flex-direction: column;
    gap: 0.5lh;
    min-width: 0;
  }

  .proc-tools {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.25lh 2ch;
  }

  .proc-filter {
    display: flex;
    align-items: center;
    gap: 1ch;
    flex: 1 1 30ch;

    & input {
      flex: 1;
      min-width: 0;
      padding: 0 1ch;
      border: var(--border-width) solid var(--border);
      border-radius: var(--radius);
      background: var(--surface-sunken);

      &::placeholder {
        color: var(--text-muted);
      }

      &:focus-visible {
        outline: none;
        border-color: var(--accent);
      }
    }
  }

  .proc-toggle,
  .proc-sort,
  .proc-pick {
    padding: 0;
    border: 0;
    background: none;
    cursor: pointer;
    text-align: inherit;
  }

  .proc-toggle {
    color: var(--text-muted);

    &[aria-pressed='true'] {
      color: var(--text);

      & .proc-mark {
        color: var(--accent);
      }
    }
  }

  .proc-scroll {
    max-block-size: max(12lh, calc(100dvh - 16lh));
    overflow: auto;
    overscroll-behavior: contain;
  }

  .proc-table {
    inline-size: 100%;
    table-layout: fixed;
    border-collapse: collapse;

    & th,
    & td {
      padding: 0 1ch 0 0;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      text-align: left;
      font-weight: 400;
    }

    & .num {
      text-align: right;
    }

    & thead th {
      position: sticky;
      top: 0;
      z-index: 1;
      background: var(--surface-raised);
      color: var(--text-muted);

      &[aria-sort] {
        color: var(--text);
      }
    }
  }

  .col-pid {
    inline-size: 9ch;
  }

  .col-user {
    inline-size: 11ch;
  }

  .col-cpu,
  .col-mem {
    inline-size: 7ch;
  }

  .col-rss {
    inline-size: 10ch;
  }

  .col-threads {
    inline-size: 5ch;
  }

  .col-state {
    inline-size: 8ch;
  }

  .proc-row {
    &:hover {
      background: var(--surface-sunken);
    }

    &.selected {
      background: var(--accent);
      color: var(--on-accent);

      & .muted {
        color: inherit;
      }
    }
  }

  .proc-pick {
    display: block;
    inline-size: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .proc-cmd {
    margin-inline-start: 1ch;
  }

  .proc-pick:focus-visible {
    outline-offset: -1px;
  }

  .proc-detail {
    display: grid;
    gap: 0.5lh;
    padding-top: 0.5lh;
    border-top: var(--border-width) solid var(--border);

    & dl {
      display: grid;
      grid-template-columns: max-content 1fr;
      gap: 0 2ch;
    }

    & dt {
      color: var(--text-muted);
    }
  }

  .proc-full {
    font: inherit;
    padding: 0.25lh 1ch;
    background: var(--surface-sunken);
    border-radius: var(--radius);
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  @container (max-width: 50rem) {
    .col-user,
    .col-rss,
    .col-threads,
    .col-state {
      display: none;
    }
  }
</style>
