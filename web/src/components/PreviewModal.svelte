<script lang="ts">
  import type { PlanResp } from '../lib/api';
  import { highlight } from '../lib/conflicts';

  let {
    plan,
    force,
    forceOverrides,
    onForce,
    onForceOverride,
    onApply,
    onCancel,
  }: {
    plan: PlanResp;
    force: boolean;
    forceOverrides: Record<string, boolean>;
    onForce: (v: boolean) => void;
    onForceOverride: (profileId: string, v: boolean) => void;
    onApply: () => void;
    onCancel: () => void;
  } = $props();

  const conflicts = $derived(highlight(plan));
</script>

<div class="preview-panel">
  <div class="preview-head">
    <h3>Preview</h3>
    <span class="badge">{plan.plans.length} profiles</span>
  </div>

  <div class="plan-list">
    {#each plan.plans as p (p.profileId)}
      <div class="plan">
        <div class="plan-head">
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
              onchange={(e) => onForceOverride(p.profileId, e.currentTarget.checked)}
            />
            overwrite conflicts in this profile
          </label>
        {:else if p.warning}
          <span class="warning">{p.warning}</span>
        {/if}
      </div>
    {/each}
  </div>

  <label class="force-all">
    <input type="checkbox" checked={force} onchange={(e) => onForce(e.currentTarget.checked)} />
    overwrite all conflicts
  </label>

  <div class="preview-actions">
    <button class="btn" onclick={onCancel}>Cancel</button>
    <button class="btn btn-primary" onclick={onApply}>Apply</button>
  </div>
</div>

<style>
  .preview-panel {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    padding: 0.85rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface-2);
  }
  .preview-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }
  .preview-head h3 {
    font-size: 0.92rem;
    font-weight: 700;
  }
  .plan-list {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
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
  .plan-head {
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
  .force-row,
  .force-all {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    color: var(--muted);
    font-size: 0.78rem;
  }
  .force-all {
    font-weight: 700;
    color: var(--text);
  }
  .preview-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }
  input[type="checkbox"] {
    accent-color: var(--accent);
  }
</style>
