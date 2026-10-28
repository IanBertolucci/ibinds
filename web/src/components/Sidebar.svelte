<script lang="ts">
  import type { ProfileDTO } from '../lib/api';
  import Icon from './Icon.svelte';

  let {
    profiles,
    activeId,
    onSelect,
  }: {
    profiles: ProfileDTO[];
    activeId: string;
    onSelect: (id: string) => void;
  } = $props();

  let search = $state('');

  const filtered = $derived(
    profiles.filter((p) => {
      const q = search.trim().toLowerCase();
      if (!q) return true;
      return p.displayName.toLowerCase().includes(q) || p.id.toLowerCase().includes(q);
    }),
  );
</script>

<nav class="sidebar" aria-label="Profiles">
  <div class="sidebar-head">
    <h2>Profiles</h2>
    <span class="count">{profiles.length}</span>
  </div>

  <label class="search">
    <Icon name="search" size={15} />
    <input type="search" placeholder="Search profiles…" aria-label="Search profiles" bind:value={search} />
  </label>

  <div class="profile-list">
    {#each filtered as p (p.id)}
      <button
        class="profile-button"
        class:active={p.id === activeId}
        class:baseline={p.isBaseline}
        onclick={() => onSelect(p.id)}
      >
        {p.displayName}{p.isBaseline ? ' (baseline)' : ''}
      </button>
    {:else}
      <p class="empty">No profiles match “{search}”.</p>
    {/each}
  </div>
</nav>

<style>
  .sidebar {
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
    padding: 0.85rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--surface);
    box-shadow: var(--shadow-sm);
  }
  .sidebar-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }
  .sidebar-head h2 {
    color: var(--muted);
    font-size: 0.72rem;
    font-weight: 800;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }
  .count {
    min-width: 1.5rem;
    padding: 0.05rem 0.4rem;
    border-radius: 999px;
    background: var(--surface-2);
    color: var(--muted);
    font-size: 0.7rem;
    font-weight: 700;
    text-align: center;
  }
  .search {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0 0.5rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface-2);
    color: var(--faint);
  }
  .search:focus-within {
    border-color: var(--accent);
    background: var(--surface);
  }
  .search input {
    width: 100%;
    min-width: 0;
    border: 0;
    padding: 0.45rem 0;
    background: transparent;
    outline: none;
  }
  .profile-list {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    max-height: calc(100vh - 230px);
    overflow-y: auto;
  }
  .profile-button {
    position: relative;
    width: 100%;
    padding: 0.5rem 0.55rem;
    border: 1px solid transparent;
    border-radius: 7px;
    background: transparent;
    color: var(--text);
    font-size: 0.84rem;
    text-align: left;
  }
  .profile-button:hover {
    background: var(--surface-2);
  }
  .profile-button.active {
    border-color: var(--accent-soft);
    background: var(--accent-soft);
    color: var(--accent-ink);
    font-weight: 700;
  }
  .profile-button.baseline::after {
    content: "";
    position: absolute;
    top: 50%;
    right: 0.55rem;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--accent);
    transform: translateY(-50%);
  }
  .empty {
    padding: 0.6rem 0.4rem;
    color: var(--muted);
    font-size: 0.8rem;
  }
  @media (min-width: 1101px) {
    .sidebar {
      height: 100%;
      min-height: 0;
    }
    .profile-list {
      flex: 1;
      min-height: 0;
      max-height: none;
    }
  }
  @media (max-width: 1100px) {
    .profile-list {
      max-height: 15rem;
    }
  }
</style>
