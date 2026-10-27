<script lang="ts">
  import type { ProfileDTO } from '../lib/api';

  let {
    profiles,
    activeId,
    onSelect,
  }: {
    profiles: ProfileDTO[];
    activeId: string;
    onSelect: (id: string) => void;
  } = $props();
</script>

<nav class="sidebar">
  <h2>Profiles</h2>
  {#each profiles as p (p.id)}
    <button
      class:active={p.id === activeId}
      onclick={() => onSelect(p.id)}
    >
      {p.displayName}{p.isBaseline ? ' (baseline)' : ''}
    </button>
  {/each}
</nav>

<style>
  .sidebar {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  button {
    text-align: left;
  }
  button.active {
    font-weight: bold;
  }
</style>
