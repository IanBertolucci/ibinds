<script lang="ts">
  import { untrack } from 'svelte';
  import { api, type PlanResp, type ProfileDTO } from '../lib/api';
  import Icon from './Icon.svelte';

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

  let pickedSource = $state(untrack(() => source ?? ''));
  let plan = $state<PlanResp | null>(null);
  let message = $state('');
  let error = $state('');

  const fixed = $derived(source !== undefined);
  const activeSource = $derived(fixed ? source : pickedSource);
  const sourceName = $derived(
    profiles.find((p) => p.id === activeSource)?.displayName ?? activeSource ?? '',
  );
  const visible = $derived(profiles.filter((p) => p.id !== activeSource));
  const targets = $derived(visible.filter((p) => selected[p.id]).map((p) => p.id));
  const allSelected = $derived(visible.length > 0 && visible.every((p) => selected[p.id]));

  let lastTargetKey = untrack(() => targets.join(','));
  $effect(() => {
    const key = targets.join(',');
    if (key !== lastTargetKey) {
      lastTargetKey = key;
      plan = null;
      message = '';
    }
  });

  function toggleAll() {
    const next = !allSelected;
    for (const p of visible) {
      if (Boolean(selected[p.id]) !== next) onToggle(p.id);
    }
  }

  async function preview() {
    error = '';
    message = '';
    if (!activeSource || targets.length === 0) {
      error = 'Pick a source and at least one target profile.';
      return;
    }
    try {
      plan = await api.planCalib({ sourceProfileId: activeSource, profileIds: targets });
    } catch (e) {
      error = String(e);
    }
  }

  async function apply() {
    error = '';
    if (!activeSource || targets.length === 0) {
      error = 'Pick a source and at least one target profile.';
      return;
    }
    try {
      const res = await api.applyCalib({ sourceProfileId: activeSource, profileIds: targets });
      message = `Calibration copied to ${res.applied.length} profile(s).`;
      plan = null;
    } catch (e) {
      error = String(e);
    }
  }
</script>

<div class="calib">
  <div class="calib-head">
    <div>
      <h3>Copy calibration</h3>
      <p class="sub">
        Copies <span class="mono">joyCalib.yaml</span> unchanged from
        <strong>{sourceName}</strong> to the selected profiles.
      </p>
    </div>
    <span class="badge">{targets.length} targets</span>
  </div>

  {#if !fixed}
    <label class="field">
      <span class="field-label">Source profile</span>
      <select bind:value={pickedSource}>
        <option value="">Select…</option>
        {#each profiles as p (p.id)}
          <option value={p.id}>{p.displayName}</option>
        {/each}
      </select>
    </label>
  {/if}

  <div class="targets">
    <div class="targets-head">
      <span class="field-label">Targets</span>
      <label class="select-all">
        <input type="checkbox" checked={allSelected} onchange={toggleAll} />
        All targets
      </label>
    </div>

    {#each visible as p (p.id)}
      <label class="target-row">
        <input type="checkbox" checked={selected[p.id]} onchange={() => onToggle(p.id)} />
        <span>{p.displayName}</span>
        {#if p.isBaseline}<span class="badge">Baseline</span>{/if}
      </label>
    {:else}
      <p class="empty">No target profiles available.</p>
    {/each}
  </div>

  {#if plan}
    <div class="plan-list">
      <span class="field-label">Preview</span>
      {#each plan.plans as p (p.profileId)}
        <div class="plan-row">
          <strong>{p.profileId}</strong>
          <span class="diff mono">{p.diffSummary}</span>
          {#if p.warning}<span class="warning">{p.warning}</span>{/if}
        </div>
      {/each}
    </div>
  {/if}

  {#if error}<p class="error">{error}</p>{/if}
  {#if message}<p class="ok">{message}</p>{/if}

  <div class="actions">
    {#if plan}
      <button class="btn" onclick={() => (plan = null)}>Back</button>
      <button class="btn btn-primary" onclick={apply}>
        <Icon name="copy" size={15} />
        Apply
      </button>
    {:else}
      <button class="btn btn-primary" onclick={preview} disabled={targets.length === 0}>
        <Icon name="refresh" size={15} />
        Preview
      </button>
    {/if}
  </div>
</div>

<style>
  .calib {
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
    padding: 1rem;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--surface);
    box-shadow: var(--shadow-sm);
  }
  .calib-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
  }
  .calib-head h3 {
    font-size: 1rem;
    font-weight: 700;
  }
  .sub {
    margin-top: 0.15rem;
    color: var(--muted);
    font-size: 0.8rem;
  }
  .targets {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    padding: 0.75rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface-2);
  }
  .targets-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    margin-bottom: 0.25rem;
  }
  .select-all {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    color: var(--text);
    font-size: 0.8rem;
    font-weight: 700;
  }
  .target-row {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    padding: 0.35rem 0.4rem;
    border-radius: 6px;
    background: var(--surface);
    font-size: 0.84rem;
  }
  .target-row:hover {
    background: #fafcfd;
  }
  input[type="checkbox"] {
    accent-color: var(--accent);
  }
  .plan-list {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    padding: 0.75rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface-2);
  }
  .plan-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.45rem;
    font-size: 0.8rem;
  }
  .diff {
    color: var(--muted);
  }
  .warning {
    color: var(--warning);
  }
  .error {
    color: var(--danger);
    font-size: 0.82rem;
  }
  .ok {
    color: var(--success);
    font-size: 0.82rem;
    font-weight: 600;
  }
  .empty {
    color: var(--muted);
    font-size: 0.8rem;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    justify-content: flex-end;
  }
</style>
