<script lang="ts">
  import { untrack } from 'svelte';
  import ProfileList from './ProfileList.svelte';
  import BindingWizard from './BindingWizard.svelte';
  import PreviewModal from './PreviewModal.svelte';
  import Icon from './Icon.svelte';
  import {
    api,
    type BindingInput,
    type ControlDTO,
    type PlanResp,
    type ProfileDTO,
  } from '../lib/api';
  import { loadControlsByProfile } from '../lib/store';
  import { focusTrap } from '../lib/focusTrap';

  let {
    profiles,
    controls,
    presetControl,
    presetProfileId = '',
    defaultDeviceInstanceGUID = '',
    onClose,
  }: {
    profiles: ProfileDTO[];
    controls: ControlDTO[];
    presetControl: string;
    presetProfileId?: string;
    defaultDeviceInstanceGUID?: string;
    onClose: () => void;
  } = $props();

  let selected = $state<Record<string, boolean>>(
    untrack(() => {
      const onlyPreset =
        presetProfileId !== '' && profiles.some((p) => p.id === presetProfileId);
      return Object.fromEntries(profiles.map((p) => [p.id, !onlyPreset || p.id === presetProfileId]));
    }),
  );
  let controlName = $state(untrack(() => presetControl));
  let controlsByProfile = $state<Record<string, ControlDTO[]>>({});
  let binding = $state<BindingInput>({});
  let plan = $state<PlanResp | null>(null);
  let force = $state(false);
  let forceOverrides = $state<Record<string, boolean>>({});
  let message = $state('');
  let error = $state('');

  async function onControl(name: string) {
    controlName = name;
    plan = null;
    controlsByProfile = await loadControlsByProfile(profiles, api.fetchControls);
  }

  function toggle(id: string) {
    selected = { ...selected, [id]: !selected[id] };
    plan = null;
  }

  function toggleAll() {
    const all = profiles.every((p) => selected[p.id]);
    const sel: Record<string, boolean> = {};
    for (const p of profiles) sel[p.id] = !all;
    selected = sel;
    plan = null;
  }

  function onBinding(b: BindingInput) {
    binding = b;
    plan = null;
  }

  async function onPlan() {
    error = '';
    message = '';
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
    error = '';
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

<div
  class="overlay"
  role="presentation"
  onclick={(e) => e.target === e.currentTarget && onClose()}
>
  <div
    class="modal binding-modal"
    use:focusTrap={onClose}
    tabindex="-1"
    role="dialog"
    aria-modal="true"
    aria-labelledby="binding-title"
  >
    <div class="modal-header">
      <div>
        <h2 id="binding-title">{presetControl ? 'Edit binding' : 'New binding'}</h2>
        <p class="modal-sub">
          {presetControl
            ? 'Update the selected control or switch to another one, then preview the changes.'
            : 'Choose a control, assign a key or button, and preview the changes.'}
        </p>
      </div>
      <button class="icon-btn" aria-label="Close binding dialog" onclick={onClose}>
        <Icon name="close" />
      </button>
    </div>

    <div class="modal-body">
      {#if error}<div class="banner banner-error" role="alert">{error}</div>{/if}
      {#if message}<div class="banner banner-ok">{message}</div>{/if}

      <div class="binding-layout">
        <section class="panel">
          <div class="panel-head">
            <h3>Profiles</h3>
            <span class="badge">current binding</span>
          </div>
          <ProfileList
            {profiles}
            {selected}
            {controlName}
            {controlsByProfile}
            onToggle={toggle}
            onToggleAll={toggleAll}
          />
        </section>

        <section class="panel">
          <div class="panel-head">
            <h3>Binding</h3>
            <span class="badge badge-accent">new value</span>
          </div>
          <BindingWizard
            {controls}
            {controlName}
            {binding}
            {defaultDeviceInstanceGUID}
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
  </div>
</div>

<style>
  .binding-modal {
    max-width: 72rem;
  }
  .modal-sub {
    margin-top: 0.1rem;
    color: var(--muted);
    font-size: 0.78rem;
  }
  .binding-layout {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.15fr);
    gap: 1rem;
  }
  .panel {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    min-width: 0;
    padding: 0.8rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface-2);
  }
  .panel-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }
  .panel-head h3 {
    font-size: 0.9rem;
    font-weight: 700;
  }
  @media (max-width: 860px) {
    .binding-layout {
      grid-template-columns: 1fr;
    }
  }
</style>
