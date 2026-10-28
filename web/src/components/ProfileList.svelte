<script lang="ts">
  import type { ControlDTO, ProfileDTO } from '../lib/api';
  import { currentBinding } from '../lib/store';

  let {
    profiles,
    selected,
    controlName,
    controlsByProfile,
    onToggle,
    onToggleAll,
  }: {
    profiles: ProfileDTO[];
    selected: Record<string, boolean>;
    controlName: string;
    controlsByProfile: Record<string, ControlDTO[]>;
    onToggle: (id: string) => void;
    onToggleAll: () => void;
  } = $props();

  const allSelected = $derived(profiles.every((p) => selected[p.id]));

  function current(p: ProfileDTO): string {
    if (!controlName) return '';
    const c = controlsByProfile[p.id]?.find((x) => x.name === controlName);
    return c ? currentBinding(c) : 'none';
  }
</script>

<div class="profile-list">
  <label class="row select-all">
    <input type="checkbox" checked={allSelected} onchange={() => onToggleAll()} />
    <span>All profiles</span>
  </label>

  {#each profiles as p (p.id)}
    <label class="row">
      <input type="checkbox" checked={selected[p.id]} onchange={() => onToggle(p.id)} />
      <span class:baseline={p.isBaseline}>
        {p.displayName}{p.isBaseline ? ' (baseline)' : ''}
      </span>
      {#if controlName}
        <span class="current mono">{current(p)}</span>
      {/if}
    </label>
  {/each}
</div>

<style>
  .profile-list {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    max-height: 40vh;
    overflow-y: auto;
  }
  .row {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    padding: 0.35rem 0.4rem;
    border-radius: 6px;
    font-size: 0.82rem;
  }
  .row:hover {
    background: var(--surface-2);
  }
  .select-all {
    margin-bottom: 0.15rem;
    border-bottom: 1px solid var(--border);
    border-radius: 0;
    color: var(--muted);
    font-weight: 700;
  }
  .baseline {
    color: var(--accent-ink);
    font-weight: 700;
  }
  .current {
    margin-left: auto;
    color: var(--faint);
    font-size: 0.72rem;
  }
  input[type="checkbox"] {
    accent-color: var(--accent);
  }
</style>
