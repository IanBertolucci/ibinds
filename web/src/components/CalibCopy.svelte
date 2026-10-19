<script lang="ts">
  import { api, type ProfileDTO } from '../lib/api';

  let {
    profiles,
    selected,
    onToggle,
  }: {
    profiles: ProfileDTO[];
    selected: Record<string, boolean>;
    onToggle: (id: string) => void;
  } = $props();

  let source = $state('');
  let message = $state('');

  const targets = $derived(Object.entries(selected).filter(([, v]) => v).map(([id]) => id));

  async function apply() {
    if (!source) {
      message = 'Pick a source profile.';
      return;
    }
    try {
      const res = await api.applyCalib({ sourceProfileId: source, profileIds: targets });
      message = `Calibration copied to ${res.applied.length} profile(s).`;
    } catch (e) {
      message = String(e);
    }
  }
</script>

<div class="calib">
  <label>
    Source profile
    <select bind:value={source}>
      <option value="">Select…</option>
      {#each profiles as p (p.id)}
        <option value={p.id}>{p.displayName}</option>
      {/each}
    </select>
  </label>

  <p>Targets ({targets.length}):</p>
  {#each profiles as p (p.id)}
    {#if p.id !== source}
      <label>
        <input type="checkbox" checked={selected[p.id]} onchange={() => onToggle(p.id)} />
        {p.displayName}
      </label>
    {/if}
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
</style>
