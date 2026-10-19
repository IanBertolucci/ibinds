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
        <span class="current">{current(p)}</span>
      {/if}
    </label>
  {/each}
</div>

<style>
  .profile-list {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    max-height: 40vh;
    overflow-y: auto;
  }
  .row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  .select-all {
    font-weight: bold;
  }
  .baseline {
    font-weight: bold;
  }
  .current {
    color: #888;
    margin-left: auto;
  }
</style>
