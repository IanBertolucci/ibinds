<script lang="ts">
  import ProfileList from './components/ProfileList.svelte';
  import BindingWizard from './components/BindingWizard.svelte';
  import CalibCopy from './components/CalibCopy.svelte';
  import PreviewModal from './components/PreviewModal.svelte';
  import Settings from './components/Settings.svelte';
  import {
    api,
    type BindingInput,
    type ControlDTO,
    type PlanResp,
    type ProfileDTO,
  } from './lib/api';
  import { loadControlsByProfile } from './lib/store';

  let profiles = $state<ProfileDTO[]>([]);
  let selected = $state<Record<string, boolean>>({});
  let controlName = $state('');
  let controls = $state<ControlDTO[]>([]);
  let controlsByProfile = $state<Record<string, ControlDTO[]>>({});
  let binding = $state<BindingInput>({});
  let plan = $state<PlanResp | null>(null);
  let force = $state(false);
  let forceOverrides = $state<Record<string, boolean>>({});
  let mode = $state<'binding' | 'calib'>('binding');
  let message = $state('');
  let error = $state('');
  let profilesDir = $state('');

  async function loadProfiles() {
    try {
      profiles = await api.fetchProfiles();
      const sel: Record<string, boolean> = {};
      for (const p of profiles) sel[p.id] = true;
      selected = sel;
      // Load the control list (shared names) from the first profile that has
      // a controls.cfg, so the function picker is populated.
      for (const p of profiles) {
        if (!p.hasControls) continue;
        try {
          controls = await api.fetchControls(p.id);
        } catch {
          /* ignore */
        }
        break;
      }
    } catch (e) {
      error = String(e);
    }
  }

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
    loadProfiles();
  });
</script>

<main>
  <h1>iBinds</h1>

  <nav>
    <button onclick={() => (mode = 'binding')} class:active={mode === 'binding'}>Bindings</button>
    <button onclick={() => (mode = 'calib')} class:active={mode === 'calib'}>Calibration</button>
  </nav>

  {#if error}<p class="error">{error}</p>{/if}
  {#if message}<p class="ok">{message}</p>{/if}

  {#if mode === 'binding'}
    <div class="layout">
      <section>
        <h2>Profiles</h2>
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
        <h2>Binding</h2>
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
  {:else}
    <CalibCopy {profiles} {selected} onToggle={toggle} />
  {/if}

  <Settings profilesDir={profilesDir} onSaved={(d) => (profilesDir = d)} />

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
</main>

<style>
  main {
    font-family: system-ui, sans-serif;
    max-width: 70rem;
    margin: 0 auto;
    padding: 1rem;
  }
  nav {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
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
  button.active {
    font-weight: bold;
  }
</style>
