<script lang="ts">
  import { api, type ProfileDTO } from '../lib/api';
  import { buildCopyReq } from '../lib/store';
  import { highlight } from '../lib/conflicts';
  import type { PlanResp } from '../lib/api';

  let {
    profiles,
    sourceId,
    controls,
    onClose,
  }: {
    profiles: ProfileDTO[];
    sourceId: string;
    controls: string[];
    onClose: () => void;
  } = $props();

  const targets = $derived(profiles.filter((p) => p.id !== sourceId));

  let selected = $state<Record<string, boolean>>(
    Object.fromEntries(targets.map((p) => [p.id, true])),
  );
  let plan = $state<PlanResp | null>(null);
  let force = $state(false);
  let forceOverrides = $state<Record<string, boolean>>({});
  let message = $state('');
  let error = $state('');

  const allSelected = $derived(targets.length > 0 && targets.every((p) => selected[p.id]));
  const selectedTargets = $derived(targets.filter((p) => selected[p.id]).map((p) => p.id));
  const conflicts = $derived(plan ? highlight(plan) : {});

  function toggle(id: string) {
    selected = { ...selected, [id]: !selected[id] };
  }

  function toggleAll() {
    const next = !allSelected;
    const sel: Record<string, boolean> = {};
    for (const p of targets) sel[p.id] = next;
    selected = sel;
  }

  async function preview() {
    error = '';
    try {
      plan = await api.planCopy(buildCopyReq(sourceId, controls, selectedTargets, force, forceOverrides));
    } catch (e) {
      error = String(e);
    }
  }

  async function apply() {
    error = '';
    try {
      const res = await api.applyCopy(buildCopyReq(sourceId, controls, selectedTargets, force, forceOverrides));
      message = `Copied to ${res.applied.length} profile(s), skipped ${res.skipped.length}.`;
      plan = null;
    } catch (e) {
      error = String(e);
    }
  }
</script>

<div class="overlay">
  <div class="modal">
  <h2>Copy bindings</h2>
  <p class="subtitle">Copying {controls.length} control(s) from {sourceId}.</p>

  <label class="select-all">
    <input type="checkbox" checked={allSelected} onchange={toggleAll} />
    All targets
  </label>
  {#each targets as p (p.id)}
    <label>
      <input type="checkbox" checked={selected[p.id] ?? false} onchange={() => toggle(p.id)} />
      {p.displayName}
    </label>
  {/each}

  {#if plan}
    {#each plan.plans as p (p.profileId)}
      <div class="plan">
        <strong>{p.profileId}</strong>
        <span class="diff">{p.diffSummary}</span>
        {#if conflicts[p.profileId]}
          <span class="conflict">conflict: {conflicts[p.profileId].map((c) => c.control).join(', ')}</span>
          <label>
            <input
              type="checkbox"
              checked={forceOverrides[p.profileId] ?? false}
              onchange={(e) => (forceOverrides = { ...forceOverrides, [p.profileId]: e.currentTarget.checked })}
            />
            overwrite conflicts
          </label>
        {:else if p.warning}
          <span class="warning">warning: {p.warning}</span>
        {/if}
      </div>
    {/each}

    <label>
      <input type="checkbox" checked={force} onchange={(e) => (force = e.currentTarget.checked)} />
      overwrite all conflicts
    </label>
  {/if}

  {#if error}<p class="error">{error}</p>{/if}
  {#if message}<p class="ok">{message}</p>{/if}

  <div class="actions">
    <button onclick={onClose}>Close</button>
    {#if !plan}
      <button onclick={preview} disabled={selectedTargets.length === 0}>Preview</button>
    {:else}
      <button onclick={apply}>Apply</button>
    {/if}
  </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.4);
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding: 2rem 1rem;
    overflow-y: auto;
    z-index: 100;
  }
  .modal {
    background: #fff;
    border: 1px solid #ccc;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-width: 40rem;
    width: 100%;
  }
  .select-all {
    font-weight: bold;
  }
  .plan {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  .conflict {
    color: #b00;
  }
  .warning {
    color: #a70;
  }
  .diff {
    color: #666;
  }
  .error {
    color: #b00;
  }
  .ok {
    color: #080;
  }
  .actions {
    display: flex;
    gap: 0.5rem;
  }
</style>
