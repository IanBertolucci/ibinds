<script lang="ts">
  import ProfileList from './ProfileList.svelte';
  import BindingWizard from './BindingWizard.svelte';
  import PreviewModal from './PreviewModal.svelte';
  import {
    api,
    type BindingInput,
    type ControlDTO,
    type PlanResp,
    type ProfileDTO,
  } from '../lib/api';
  import { loadControlsByProfile } from '../lib/store';

  let {
    profiles,
    controls,
    presetControl,
    onClose,
  }: {
    profiles: ProfileDTO[];
    controls: ControlDTO[];
    presetControl: string;
    onClose: () => void;
  } = $props();

  let selected = $state<Record<string, boolean>>(
    Object.fromEntries(profiles.map((p) => [p.id, true])),
  );
  let controlName = $state(presetControl);
  let controlsByProfile = $state<Record<string, ControlDTO[]>>({});
  let binding = $state<BindingInput>({});
  let plan = $state<PlanResp | null>(null);
  let force = $state(false);
  let forceOverrides = $state<Record<string, boolean>>({});
  let message = $state('');
  let error = $state('');

  async function onControl(name: string) {
    controlName = name;
    controlsByProfile = await loadControlsByProfile(profiles, api.fetchControls);
  }

  function toggle(id: string) {
    selected = { ...selected, [id]: !selected[id] };
  }

  function toggleAll() {
    const all = profiles.every((p) => selected[p.id]);
    const sel: Record<string, boolean> = {};
    for (const p of profiles) sel[p.id] = !all;
    selected = sel;
  }

  function onBinding(b: BindingInput) {
    binding = b;
  }

  async function onPlan() {
    const ids = Object.entries(selected)
      .filter(([, v]) => v)
      .map(([id]) => id);
    try {
      plan = await api.planBinding({
        control: controlName,
        binding,
        profileIds: ids,
        force,
        profileOverrides: Object.entries(forceOverrides)
          .filter(([, v]) => v)
          .map(([profileId]) => ({ profileId, force: true })),
      });
    } catch (e) {
      error = String(e);
    }
  }

  async function onApply() {
    const ids = Object.entries(selected)
      .filter(([, v]) => v)
      .map(([id]) => id);
    try {
      const res = await api.applyBinding({
        control: controlName,
        binding,
        profileIds: ids,
        force,
        profileOverrides: Object.entries(forceOverrides)
          .filter(([, v]) => v)
          .map(([profileId]) => ({ profileId, force: true })),
      });
      message = `Applied to ${res.applied.length} profile(s), skipped ${res.skipped.length}.`;
      plan = null;
    } catch (e) {
      error = String(e);
    }
  }

  $effect(() => {
    if (presetControl) {
      onControl(presetControl);
    }
  });
</script>

<div class="modal">
  <div class="header">
    <h2>New binding</h2>
    <button onclick={onClose}>Close</button>
  </div>

  {#if error}<p class="error">{error}</p>{/if}
  {#if message}<p class="ok">{message}</p>{/if}

  <div class="layout">
    <section>
      <h3>Profiles</h3>
      <ProfileList
        {profiles}
        {selected}
        {controlName}
        {controlsByProfile}
        onToggle={toggle}
        onToggleAll={toggleAll}
      />
    </section>
    <section>
      <BindingWizard
        {controls}
        {controlName}
        {binding}
        onControl={onControl}
        onBinding={onBinding}
        onPlan={onPlan}
      />
    </section>
  </div>

  {#if plan}
    <PreviewModal
      {plan}
      {force}
      {forceOverrides}
      onForce={(v) => (force = v)}
      onForceOverride={(id, v) => (forceOverrides = { ...forceOverrides, [id]: v })}
      onApply={onApply}
      onCancel={() => (plan = null)}
    />
  {/if}
</div>

<style>
  .modal {
    border: 1px solid #ccc;
    padding: 1rem;
  }
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .layout {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
  }
  .error {
    color: #b00;
  }
  .ok {
    color: #080;
  }
</style>
