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

<div class="modal">
  <h2>Preview</h2>
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
            onchange={(e) => onForceOverride(p.profileId, e.currentTarget.checked)}
          />
          overwrite conflicts
        </label>
      {:else if p.warning}
        <span class="warning">warning: {p.warning}</span>
      {/if}
    </div>
  {/each}

  <label>
    <input type="checkbox" checked={force} onchange={(e) => onForce(e.currentTarget.checked)} />
    overwrite all conflicts
  </label>

  <div class="actions">
    <button onclick={onCancel}>Cancel</button>
    <button onclick={onApply}>Apply</button>
  </div>
</div>

<style>
  .modal {
    border: 1px solid #ccc;
    padding: 1rem;
  }
  .plan {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    margin-bottom: 0.5rem;
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
  .actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 1rem;
  }
</style>
