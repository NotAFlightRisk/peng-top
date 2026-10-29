<script lang="ts">
  import type { Snippet } from 'svelte'

  let {
    title,
    hotkey,
    aside,
    children,
    class: extra = '',
  }: { title: string; hotkey?: string; aside?: Snippet; children: Snippet; class?: string } =
    $props()

  const id = $props.id()
</script>

<section class="panel {extra}" aria-labelledby={id}>
  <header class="panel-head">
    <h2 class="panel-title" {id}>
      {#if hotkey}<kbd class="panel-key">{hotkey}</kbd>{/if}{title}
    </h2>
    {#if aside}<div class="panel-aside">{@render aside()}</div>{/if}
  </header>
  <div class="panel-body">{@render children()}</div>
</section>

<style>
  .panel {
    position: relative;
    display: flex;
    flex-direction: column;
    min-width: 0;
    border: var(--border-width) solid var(--border);
    border-radius: var(--radius);
    padding: 0 var(--gap) var(--gap-y);

    &:focus-within {
      border-color: var(--border-strong);
    }
  }

  .panel-head {
    display: flex;
    justify-content: space-between;
    gap: var(--gap);
    margin-top: -0.7lh;
    margin-bottom: 0.25lh;
    white-space: nowrap;
    min-width: 0;
  }

  .panel-title,
  .panel-aside {
    background: var(--surface);
    padding-inline: 1ch;
    margin-inline: -1ch;
  }

  .panel-title {
    color: var(--text);
  }

  .panel-key {
    margin-right: 1ch;
    color: var(--accent);

    &::before {
      content: '[';
      color: var(--text-muted);
    }

    &::after {
      content: ']';
      color: var(--text-muted);
    }
  }

  .panel-aside {
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--text-muted);
    margin-inline: 0 -1ch;
  }

  .panel-body {
    display: flex;
    flex-direction: column;
    gap: 0.25lh;
    min-width: 0;
    flex: 1;
  }
</style>
