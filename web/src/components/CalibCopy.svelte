<script lang="ts">
  import { api, type ProfileDTO } from '../lib/api';

  let {
    profiles,
    selected,
    onToggle,
    source,
  }: {
    profiles: ProfileDTO[];
    selected: Record<string, boolean>;
    onToggle: (id: string) => void;
    source?: string;
  } = $props();

  let pickedSource = $state(source ?? '');
  let message = $state('');

  const fixed = $derived(source !== undefined);
  const activeSource = $derived(fixed ? source : pickedSource);
  const visible = $derived(profiles.filter((p) => p.id !== activeSource));
  const targets = $derived(visible.filter((p) => selected[p.id]).map((p) => p.id));
  const allSelected = $derived(visible.length > 0 && visible.every((p) => selected[p.id]));

  function toggleAll() {
    const next = !allSelected;
    for (const p of visible) {
      if (Boolean(selected[p.id]) !== next) onToggle(p.id);
    }
  }

  async function apply() {
    if (!activeSource) {
      message = 'Pick a source profile.';
      return;
    }
    try {
      const res = await api.applyCalib({ sourceProfileId: activeSource, profileIds: targets });
      message = `Calibration copied to ${res.applied.length} profile(s).`;
    } catch (e) {
      message = String(e);
    }
  }
</script>

<div class="calib">
  {#if !fixed}
    <label>
      Source profile
      <select bind:value={pickedSource}>
        <option value="">Select…</option>
        {#each profiles as p (p.id)}
          <option value={p.id}>{p.displayName}</option>
        {/each}
      </select>
    </label>
  {/if}

  <p>Targets ({targets.length}):</p>
  <label class="select-all">
    <input type="checkbox" checked={allSelected} onchange={toggleAll} />
    All targets
  </label>
  {#each visible as p (p.id)}
    <label>
      <input type="checkbox" checked={selected[p.id]} onchange={() => onToggle(p.id)} />
      {p.displayName}
    </label>
  {/each}

  <button onclick={apply}>Copy calibration</button>
  {#if message}<p>{message}</p>{/if}
</div>

<style>
  .calib {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .select-all {
    font-weight: bold;
  }
</style>
