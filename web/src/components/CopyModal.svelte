<script lang="ts">
  import { untrack } from 'svelte';
  import { api, type PlanResp, type ProfileDTO } from '../lib/api';
  import { buildCopyReq } from '../lib/store';
  import { highlight } from '../lib/conflicts';
  import Icon from './Icon.svelte';
  import { focusTrap } from '../lib/focusTrap';

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
  const sourceName = $derived(
    profiles.find((p) => p.id === sourceId)?.displayName ?? sourceId,
  );

  let selected = $state<Record<string, boolean>>(
    untrack(() => Object.fromEntries(targets.map((p) => [p.id, true]))),
  );
  let plan = $state<PlanResp | null>(null);
  let force = $state(false);
  let forceOverrides = $state<Record<string, boolean>>({});
  let message = $state('');
  let error = $state('');

  const allSelected = $derived(targets.length > 0 && targets.every((p) => selected[p.id]));
  const selectedTargets = $derived(targets.filter((p) => selected[p.id]).map((p) => p.id));
  const conflicts = $derived(plan ? highlight(plan) : {});

  let lastTargetKey = untrack(() => selectedTargets.join(','));
  $effect(() => {
    const key = selectedTargets.join(',');
    if (key !== lastTargetKey) {
      lastTargetKey = key;
      plan = null;
    }
  });

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
    message = '';
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

<div
  class="overlay"
  role="presentation"
  onclick={(e) => e.target === e.currentTarget && onClose()}
>
  <div
    class="modal copy-modal"
    use:focusTrap={onClose}
    tabindex="-1"
    role="dialog"
    aria-modal="true"
    aria-labelledby="copy-title"
  >
    <div class="modal-header">
      <div>
        <h2 id="copy-title">Copy bindings</h2>
        <p class="modal-sub">
          Copying {controls.length} control(s) from <strong>{sourceName}</strong>.
        </p>
      </div>
      <button class="icon-btn" aria-label="Close copy dialog" onclick={onClose}>
        <Icon name="close" />
      </button>
    </div>

    <div class="modal-body">
      <section class="targets">
        <div class="targets-head">
          <span class="field-label">Target profiles</span>
          <label class="select-all">
            <input type="checkbox" checked={allSelected} onchange={toggleAll} />
            All targets
          </label>
        </div>

        {#each targets as p (p.id)}
          <label class="target-row">
            <input type="checkbox" checked={selected[p.id] ?? false} onchange={() => toggle(p.id)} />
            <span>{p.displayName}</span>
            {#if p.isBaseline}<span class="badge">Baseline</span>{/if}
          </label>
        {:else}
          <p class="empty">No other profiles to copy to.</p>
        {/each}
      </section>

      {#if plan}
        <section class="plan-list">
          <div class="plan-head">
            <span class="field-label">Preview</span>
            <label class="force-all">
              <input type="checkbox" checked={force} onchange={(e) => (force = e.currentTarget.checked)} />
              overwrite all conflicts
            </label>
          </div>

          {#each plan.plans as p (p.profileId)}
            <div class="plan">
              <div class="plan-title">
                <strong>{p.profileId}</strong>
                <span class="diff mono">{p.diffSummary}</span>
              </div>
              {#if conflicts[p.profileId]}
                <span class="conflict">
                  Conflict: {conflicts[p.profileId].map((c) => c.control).join(', ')}
                </span>
                <label class="force-row">
                  <input
                    type="checkbox"
                    checked={forceOverrides[p.profileId] ?? false}
                    onchange={(e) =>
                      (forceOverrides = { ...forceOverrides, [p.profileId]: e.currentTarget.checked })}
                  />
                  overwrite conflicts in this profile
                </label>
              {:else if p.warning}
                <span class="warning">{p.warning}</span>
              {/if}
            </div>
          {/each}
        </section>
      {/if}

      {#if error}<div class="banner banner-error" role="alert">{error}</div>{/if}
      {#if message}<div class="banner banner-ok">{message}</div>{/if}
    </div>

    <div class="modal-actions">
      <button class="btn" onclick={onClose}>Close</button>
      {#if !plan}
        <button class="btn btn-primary" onclick={preview} disabled={selectedTargets.length === 0}>
          Preview
        </button>
      {:else}
        <button class="btn btn-primary" onclick={apply}>Apply</button>
      {/if}
    </div>
  </div>
</div>

<style>
  .copy-modal {
    max-width: 52rem;
  }
  .modal-sub {
    margin-top: 0.1rem;
    color: var(--muted);
    font-size: 0.78rem;
  }
  .targets {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    padding: 0.7rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface-2);
  }
  .targets-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    margin-bottom: 0.2rem;
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
  .plan-list {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    padding: 0.7rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface-2);
  }
  .plan-head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }
  .force-all {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    color: var(--text);
    font-size: 0.78rem;
    font-weight: 700;
  }
  .plan {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    padding: 0.55rem 0.65rem;
    border: 1px solid var(--border);
    border-radius: 7px;
    background: var(--surface);
  }
  .plan-title {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    justify-content: space-between;
    gap: 0.4rem;
    font-size: 0.82rem;
  }
  .diff {
    color: var(--muted);
    font-size: 0.75rem;
  }
  .conflict {
    color: var(--danger);
    font-size: 0.78rem;
  }
  .warning {
    color: var(--warning);
    font-size: 0.78rem;
  }
  .force-row {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    color: var(--muted);
    font-size: 0.78rem;
  }
  .empty {
    color: var(--muted);
    font-size: 0.8rem;
  }
  input[type="checkbox"] {
    accent-color: var(--accent);
  }
</style>
